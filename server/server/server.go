package server

import (
	"context"
	_tls "crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"sync"

	"github.com/cntechpower/utils/tracing"

	"github.com/cntechpower/anywhere/dao/connlist"

	"github.com/cntechpower/anywhere/server/zone"

	configDao "github.com/cntechpower/anywhere/dao/config"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/util"
)

type Server struct {
	serverId           string
	serverAddr         *net.TCPAddr
	credential         *_tls.Config
	listener           net.Listener
	zones              map[string] /*user*/ map[string] /*zone*/ zone.IZone
	agentsRwMutex      sync.RWMutex
	ExitChan           chan error
	ErrChan            chan error
	statusRwMutex      sync.RWMutex
	statusCache        model.ServerSummary
	userValidator      *auth.UserValidator
	allProxyConfigList []*model.ProxyConfig
}

var serverInstance *Server

func GetServerInstance() *Server {
	return serverInstance
}

func InitServerInstance(serverId string, port int, users *model.UserConfig) *Server {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		panic(err)
	}
	serverInstance = &Server{
		serverId:      serverId,
		serverAddr:    addr,
		zones:         make(map[string]map[string]zone.IZone, 0),
		agentsRwMutex: sync.RWMutex{},
		ExitChan:      make(chan error, 1),
		ErrChan:       make(chan error, 10000),
		statusCache:   model.ServerSummary{},
		userValidator: auth.NewUserValidator(users),
	}
	return serverInstance
}

func (s *Server) GetUserValidator() *auth.UserValidator {
	return s.userValidator
}

func (s *Server) SetCredentials(config *_tls.Config) {
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

func (s *Server) Start(ctx context.Context) {
	h := log.NewHeader("serverStart")
	if err := s.checkServerInit(); err != nil {
		panic(err)
	}
	ln, err := _tls.Listen("tcp", s.serverAddr.String(), s.credential)
	if err != nil {
		panic(err)
	}
	s.listener = ln
	go s.RefreshSummaryLoop(ctx)
	go func() {
		for {
			c, err := s.listener.Accept()
			if err != nil {
				log.Infof(h, "server port accept conn error: %v", err)
				continue
			}
			go s.handleNewServerConnection(c)

		}
	}()

}

func (s *Server) handleNewServerConnection(c net.Conn) {
	span, ctx := tracing.New(context.TODO(), "handleNewServerConnection")
	defer span.Finish()
	h := log.NewHeader("handleNewServerConnection")
	var msg model.RequestMsg
	d := json.NewDecoder(c)

	if err := d.Decode(&msg); err != nil {
		h.Errorc(ctx, "unmarshal init pkg from %s error: %v", c.RemoteAddr(), err)
		_ = c.Close()
		return
	}
	switch msg.ReqType {
	case model.PkgControlConnRegister:
		m, err := model.ParseControlRegisterPkg(msg.Message)
		if err != nil {
			h.Errorc(ctx, "get corrupted ControlRegister packet from %v", c.RemoteAddr())
			_ = c.Close()
			return
		}
		if !s.userValidator.ValidateUserPass(m.UserName, m.PassWord) {
			h.Errorc(ctx, "validate userName and password from %v fail", c.RemoteAddr())
			_ = conn.NewWrappedConn(m.AgentId, c).Send(model.NewAuthenticationFailMsg("validate userName and password fail"))
			_ = c.Close()
			return
		}
		if isUpdate := s.RegisterAgent(m.UserName, m.AgentGroup, m.AgentId, c); isUpdate {
			h.Errorc(ctx, "rebuild control connection for zone: %v, agent: %v", m.AgentGroup, m.AgentId)
		} else {
			h.Errorc(ctx, "accept control connection from zone: %v, agent: %v", m.AgentGroup, m.AgentId)
		}
	case model.PkgTunnelBegin:
		m, err := model.ParseTunnelBeginPkg(msg.Message)
		if err != nil {
			h.Errorc(ctx, "get corrupted PkgTunnelBegin packet from %v", c.RemoteAddr())
			_ = c.Close()
			return
		}
		if !s.isZoneExist(m.UserName, m.AgentGroup) {
			h.Errorc(ctx, "got data conn register pkg from unknown user %v, zone %v", m.UserName, m.AgentGroup)
			_ = c.Close()
		} else {
			h.Infoc(ctx, "add data conn for %v from user %v, group %v", m.UserName, m.LocalAddr, m.AgentGroup)
			if err := s.zones[m.UserName][m.AgentGroup].PutProxyConn(ctx, m.AgentId, m.LocalAddr, c); err != nil {
				h.Errorc(ctx, "put proxy conn to agent error: %v", err)
			}
		}
	default:
		h.Errorc(ctx, "unknown msg type %v from %v", msg.ReqType, c.RemoteAddr())
		_ = c.Close()

	}

}

func (s *Server) isZoneExist(userName, zoneName string) (exists bool) {
	if _, userExist := s.zones[userName]; !userExist {
		exists = false
		return
	}
	_, exists = s.zones[userName][zoneName]
	return
}

func (s *Server) ListAgentInfo() (res []*model.AgentInfoInServer) {
	res = make([]*model.AgentInfoInServer, 0)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.Infos()...)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Id > res[j].Id
	})
	return res
}

func (s *Server) ListProxyConfigs() []*model.ProxyConfig {
	// we assume that we had 100 proxy config.
	res := make([]*model.ProxyConfig, 0, 100)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.ListProxyConfigs()...)
		}
	}
	return res
}

func (s *Server) ListZones() []*model.ZoneInfo {
	// we assume that we had 100 proxy config.
	res := make([]*model.ZoneInfo, 0, 100)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.Info())
		}
	}
	return res
}

func (s *Server) RegisterAgent(userName, zoneName, agentId string, c net.Conn) (isUpdate bool) {
	if _, ok := s.zones[userName]; !ok {
		s.zones[userName] = make(map[string]zone.IZone)
	}
	if _, ok := s.zones[userName][zoneName]; !ok {
		s.zones[userName][zoneName] = zone.NewZone(userName, zoneName)
	}
	return s.zones[userName][zoneName].RegisterAgent(agentId, c)

}

func (s *Server) ListJoinedConns(userName, zoneName string) (res []*model.GroupConnList, err error) {
	res = make([]*model.GroupConnList, 0)
	if userName != "" && zoneName != "" { // only get specified zone
		if !s.isZoneExist(userName, zoneName) {
			return nil, fmt.Errorf("no such zone %v", zoneName)
		}
		var list []*model.JoinedConnListItem
		list, err = s.zones[userName][zoneName].ListJoinedConns()
		if err != nil {
			return
		}
		res = append(res, &model.GroupConnList{
			UserName: userName,
			ZoneName: zoneName,
			List:     list,
		})
		return res, nil
	}
	// get all userName's group conn
	for userName, zones := range s.zones {
		for zoneName, z := range zones {
			var list []*model.JoinedConnListItem
			list, err = z.ListJoinedConns()
			if err != nil {
				return
			}
			res = append(res, &model.GroupConnList{
				UserName: userName,
				ZoneName: zoneName,
				List:     list,
			})

		}
	}
	return res, nil
}

func (s *Server) killJoinedConn(userName, zoneName string, id uint) error {
	if zoneName == "" {
		return fmt.Errorf("zone is empty")
	}
	if !s.isZoneExist(userName, zoneName) {
		return fmt.Errorf("no such zone %v", zoneName)
	}
	return s.zones[userName][zoneName].KillJoinedConnById(id)
}

func (s *Server) KillJoinedConnById(id int64) (err error) {
	c, err := connlist.GetJoinedConnById(id)
	if err != nil {
		return
	}
	return s.killJoinedConn(c.UserName, c.ZoneName, c.ID)
}

func (s *Server) FlushJoinedConns() {
	for _, zones := range s.zones {
		for _, z := range zones {
			z.FlushJoinedConns()
		}
	}
}

func (s *Server) UpdateProxyConfigWhiteList(userName, zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) (err error) {
	if zoneName == "" {
		return fmt.Errorf("zone is empty")
	}
	if !s.isZoneExist(userName, zoneName) {
		return fmt.Errorf("no such zone %v", zoneName)
	}
	err = s.zones[userName][zoneName].UpdateProxyConfigWhiteListConfig(remotePort, localAddr, whiteCidrs, whiteListEnable)
	if err == nil {
		err = configDao.Update(&model.ProxyConfig{
			UserName:      userName,
			ZoneName:      zoneName,
			RemotePort:    remotePort,
			LocalAddr:     localAddr,
			IsWhiteListOn: whiteListEnable,
			WhiteCidrList: whiteCidrs,
		})
	}
	return
}

func (s *Server) LoadProxyConfigFile() error {
	h := log.NewHeader("LoadProxyConfigFile")
	if err := configDao.Iterator(func(config *model.ProxyConfig) {
		err := s.AddProxyConfigByModel(config)
		if err != nil {
			h.Errorf("load config %+v error: %v", config, err)
		}
	}); err != nil {
		return err
	}
	return nil
}
