package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/cntechpower/utils/log"

	configDao "github.com/cntechpower/anywhere/dao/config"
	"github.com/cntechpower/anywhere/internal/trace"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/server/zone"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/util"
)

type Server struct {
	serverId   string
	serverAddr *net.TCPAddr
	credential *tls.Config
	listener   net.Listener

	// zonesMu guards zones. Every access to the map must go through
	// getZone/getOrCreateZone while holding it.
	zonesMu sync.RWMutex
	zones   map[string] /*user*/ map[string] /*zone*/ zone.IZone

	userValidator *auth.UserValidator

	statusRwMutex sync.RWMutex
	statusCache   model.ServerSummary

	closeCh   chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup

	// idleTimeout closes proxied TCP connections that stay idle for too long.
	// Zero disables the timeout.
	idleTimeout time.Duration
}

// New builds a server instance. There is no package level singleton: callers
// own the instance and inject it into the HTTP/RPC layers.
func New(serverId string, port int, users *model.UserConfig, idleTimeout time.Duration) (*Server, error) {
	if serverId == "" {
		return nil, fmt.Errorf("serverId is empty")
	}
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		return nil, fmt.Errorf("invalid server port %d: %w", port, err)
	}
	return &Server{
		serverId:      serverId,
		serverAddr:    addr,
		zones:         make(map[string]map[string]zone.IZone),
		userValidator: auth.NewUserValidator(users),
		closeCh:       make(chan struct{}),
		idleTimeout:   idleTimeout,
	}, nil
}

func (s *Server) GetUserValidator() *auth.UserValidator {
	return s.userValidator
}

func (s *Server) SetCredentials(config *tls.Config) {
	s.credential = config
}

func (s *Server) checkServerInit() error {
	if s.credential == nil {
		return fmt.Errorf("credential is empty")
	}
	if s.serverId == "" {
		return fmt.Errorf("serverId is empty")
	}
	if s.serverAddr == nil {
		return fmt.Errorf("serverAddr is empty")
	}
	return nil
}

// Start opens the agent facing TLS listener and starts the maintenance loops.
func (s *Server) Start(ctx context.Context) error {
	if err := s.checkServerInit(); err != nil {
		return err
	}
	ln, err := tls.Listen("tcp", s.serverAddr.String(), s.credential)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.serverAddr, err)
	}
	s.listener = ln

	s.wg.Add(1)
	go s.acceptLoop()
	s.wg.Add(1)
	go s.maintenanceLoop(ctx)
	return nil
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	h := log.NewHeader("serverStart")
	for {
		c, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			select {
			case <-s.closeCh:
				return
			case <-time.After(100 * time.Millisecond):
			}
			h.Errorf("accept conn error: %v", err)
			continue
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleNewServerConnection(c)
		}()
	}
}

func (s *Server) handleNewServerConnection(c net.Conn) {
	span, ctx := trace.New(context.Background(), "handleNewServerConnection")
	defer span.Finish()
	h := log.NewHeader("handleNewServerConnection")

	wc := conn.NewWrappedConn(s.serverId, c)
	var msg model.RequestMsg
	if err := wc.Receive(&msg); err != nil {
		h.Errorf("unmarshal init pkg from %s error: %v", c.RemoteAddr(), err)
		_ = c.Close()
		return
	}

	switch msg.ReqType {
	case model.PkgControlConnRegister:
		m, err := model.ParseControlRegisterPkg(msg.Message)
		if err != nil {
			h.Errorf("corrupted ControlRegister packet from %s: %v", c.RemoteAddr(), err)
			_ = c.Close()
			return
		}
		if !s.userValidator.ValidateUserPass(m.UserName, m.PassWord) {
			h.Errorf("validate userName and password from %s failed", c.RemoteAddr())
			_ = wc.Send(model.NewAuthenticationFailMsg("validate userName and password fail"))
			_ = c.Close()
			return
		}
		if isUpdate := s.RegisterAgent(m.UserName, m.AgentGroup, m.AgentId, c); isUpdate {
			h.Infof("rebuilt control connection for zone %s, agent %s", m.AgentGroup, m.AgentId)
		} else {
			h.Infof("accepted control connection from zone %s, agent %s", m.AgentGroup, m.AgentId)
		}
	case model.PkgTunnelBegin:
		m, err := model.ParseTunnelBeginPkg(msg.Message)
		if err != nil {
			h.Errorf("corrupted PkgTunnelBegin packet from %s: %v", c.RemoteAddr(), err)
			_ = c.Close()
			return
		}
		z, ok := s.getZone(m.UserName, m.AgentGroup)
		if !ok {
			h.Errorf("data conn from unknown user %s / zone %s", m.UserName, m.AgentGroup)
			_ = c.Close()
			return
		}
		if err := z.PutProxyConnection(ctx, m.AgentId, m.LocalAddr, m.Token, c); err != nil {
			h.Errorf("put proxy conn from agent %s error: %v", m.AgentId, err)
			_ = c.Close()
			return
		}
		h.Infof("added data conn for %s from user %s, group %s", m.LocalAddr, m.UserName, m.AgentGroup)
	default:
		h.Errorf("unknown msg type %s from %s", msg.ReqType, c.RemoteAddr())
		_ = c.Close()
	}
}

func (s *Server) getZone(userName, zoneName string) (zone.IZone, bool) {
	s.zonesMu.RLock()
	defer s.zonesMu.RUnlock()
	userZones, ok := s.zones[userName]
	if !ok {
		return nil, false
	}
	z, ok := userZones[zoneName]
	return z, ok
}

// getOrCreateZone returns the zone, creating it if needed.
func (s *Server) getOrCreateZone(userName, zoneName string) zone.IZone {
	s.zonesMu.Lock()
	defer s.zonesMu.Unlock()
	userZones, ok := s.zones[userName]
	if !ok {
		userZones = make(map[string]zone.IZone)
		s.zones[userName] = userZones
	}
	if z, ok := userZones[zoneName]; ok {
		return z
	}
	z := zone.NewZone(userName, zoneName, s.idleTimeout)
	userZones[zoneName] = z
	return z
}

func (s *Server) RegisterAgent(userName, zoneName, agentId string, c net.Conn) (isUpdate bool) {
	return s.getOrCreateZone(userName, zoneName).RegisterAgent(agentId, c)
}

func (s *Server) ListAgentInfo() []*model.AgentInfoInServer {
	s.zonesMu.RLock()
	res := make([]*model.AgentInfoInServer, 0)
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.Infos()...)
		}
	}
	s.zonesMu.RUnlock()

	sort.Slice(res, func(i, j int) bool { return res[i].Id > res[j].Id })
	return res
}

func (s *Server) ListProxyConfigs() []*model.ProxyConfig {
	res := make([]*model.ProxyConfig, 0, 100)
	s.zonesMu.RLock()
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.ListProxyConfigs()...)
		}
	}
	s.zonesMu.RUnlock()
	return res
}

func (s *Server) ListZones() []*model.ZoneInfo {
	res := make([]*model.ZoneInfo, 0, 100)
	s.zonesMu.RLock()
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.Info())
		}
	}
	s.zonesMu.RUnlock()
	return res
}

func (s *Server) ListJoinedConns(userName, zoneName string) ([]*model.GroupConnList, error) {
	res := make([]*model.GroupConnList, 0)

	if userName != "" && zoneName != "" {
		z, ok := s.getZone(userName, zoneName)
		if !ok {
			return nil, fmt.Errorf("no such zone %s", zoneName)
		}
		return append(res, &model.GroupConnList{
			UserName: userName,
			ZoneName: zoneName,
			List:     z.ListJoinedConnections(),
		}), nil
	}

	s.zonesMu.RLock()
	for user, zones := range s.zones {
		for zName, z := range zones {
			if userName != "" && user != userName {
				continue
			}
			if zoneName != "" && zName != zoneName {
				continue
			}
			res = append(res, &model.GroupConnList{
				UserName: user,
				ZoneName: zName,
				List:     z.ListJoinedConnections(),
			})
		}
	}
	s.zonesMu.RUnlock()
	return res, nil
}

// KillJoinedConnById searches every zone because connection IDs are process
// wide unique but the zone is not part of the API request.
func (s *Server) KillJoinedConnById(id int64) error {
	s.zonesMu.RLock()
	zones := make([]zone.IZone, 0, len(s.zones))
	for _, userZones := range s.zones {
		for _, z := range userZones {
			zones = append(zones, z)
		}
	}
	s.zonesMu.RUnlock()

	for _, z := range zones {
		if err := z.KillJoinedConnById(uint(id)); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no such connection %d", id)
}

func (s *Server) FlushJoinedConns() {
	s.zonesMu.RLock()
	zones := make([]zone.IZone, 0, len(s.zones))
	for _, userZones := range s.zones {
		for _, z := range userZones {
			zones = append(zones, z)
		}
	}
	s.zonesMu.RUnlock()
	for _, z := range zones {
		z.FlushJoinedConnections()
	}
}

func (s *Server) UpdateProxyConfigWhiteList(userName, zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	z, ok := s.getZone(userName, zoneName)
	if !ok {
		return fmt.Errorf("no such zone %s", zoneName)
	}
	if err := z.UpdateProxyConfigWhiteListConfig(remotePort, localAddr, whiteCidrs, whiteListEnable); err != nil {
		return err
	}
	return configDao.Update(&model.ProxyConfig{
		UserName:      userName,
		ZoneName:      zoneName,
		RemotePort:    remotePort,
		LocalAddr:     localAddr,
		IsWhiteListOn: whiteListEnable,
		WhiteCidrList: whiteCidrs,
	})
}

// LoadProxyConfigFile is kept for API compatibility. Configs are restored
// per-zone when the zone is created, so there is nothing left to load here.
func (s *Server) LoadProxyConfigFile() error {
	return configDao.Iterator(func(cfg *model.ProxyConfig) {
		if err := s.AddProxyConfigByModel(cfg); err != nil {
			log.Errorf(log.NewHeader("LoadProxyConfigFile"), "load config %+v error: %v", cfg, err)
		}
	})
}

// Shutdown stops accepting connections and closes every zone.
func (s *Server) Shutdown(ctx context.Context) error {
	s.closeOnce.Do(func() {
		close(s.closeCh)
		if s.listener != nil {
			_ = s.listener.Close()
		}
		s.zonesMu.Lock()
		zones := make([]zone.IZone, 0, len(s.zones))
		for user, userZones := range s.zones {
			for name, z := range userZones {
				zones = append(zones, z)
				delete(userZones, name)
			}
			delete(s.zones, user)
		}
		s.zonesMu.Unlock()
		for _, z := range zones {
			z.Shutdown()
		}
	})

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
