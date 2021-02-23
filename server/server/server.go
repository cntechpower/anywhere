package server

import (
	_tls "crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/agent"
	"github.com/cntechpower/anywhere/server/auth"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
)

type Server struct {
	serverId                         string
	serverAddr                       *net.TCPAddr
	credential                       *_tls.Config
	listener                         net.Listener
	zones                            map[string] /*user*/ map[string] /*zone*/ agent.IZone
	agentsRwMutex                    sync.RWMutex
	ExitChan                         chan error
	ErrChan                          chan error
	statusRwMutex                    sync.RWMutex
	statusCache                      model.ServerSummary
	userValidator                    *auth.UserValidator
	allProxyConfigList               []*model.ProxyConfig
	networkFlowSortedProxyConfigList []*model.ProxyConfig
	rejectCountSortedProxyConfigList []*model.ProxyConfig
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
		zones:         make(map[string]map[string]agent.IZone, 0),
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

func (s *Server) Start() {
	h := log.NewHeader("serverStart")
	if err := s.checkServerInit(); err != nil {
		panic(err)
	}
	ln, err := _tls.Listen("tcp", s.serverAddr.String(), s.credential)
	if err != nil {
		panic(err)
	}
	s.listener = ln
	go s.RefreshSummaryLoop()
	go s.StartReportCron()

	go func() {
		for {
			c, err := s.listener.Accept()
			if err != nil {
				log.Infof(h, "server port accept conn error: %v", err)
				continue
			}
			go s.handleNewConnection(c)

		}
	}()

}

func (s *Server) handleNewConnection(c net.Conn) {
	h := log.NewHeader("handleNewAgentConn")
	var msg model.RequestMsg
	d := json.NewDecoder(c)

	if err := d.Decode(&msg); err != nil {
		log.Errorf(h, "unmarshal init pkg from %s error: %v", c.RemoteAddr(), err)
		_ = c.Close()
		return
	}
	switch msg.ReqType {
	case model.PkgControlConnRegister:
		m, _ := model.ParseControlRegisterPkg(msg.Message)
		if !s.userValidator.ValidateUserPass(m.UserName, m.PassWord) {
			log.Errorf(h, "validate userName and password from %v fail", c.RemoteAddr())
			_ = conn.NewWrappedConn(m.AgentId, c).Send(model.NewAuthenticationFailMsg("validate userName and password fail"))
			_ = c.Close()
			return
		}
		if isUpdate := s.RegisterAgent(m.UserName, m.AgentGroup, m.AgentId, c); isUpdate {
			log.Errorf(h, "rebuild control connection for agent: %v", m.AgentId)
		} else {
			log.Infof(h, "accept control connection from agent: %v", m.AgentId)
		}
	case model.PkgTunnelBegin:
		m, err := model.ParseTunnelBeginPkg(msg.Message)
		if err != nil {
			log.Errorf(h, "get corrupted PkgTunnelBegin packet from %v", c.RemoteAddr())
			_ = c.Close()
			return
		}
		if !s.isZoneExist(m.UserName, m.AgentGroup) {
			log.Errorf(h, "got data conn register pkg from unknown user %v, group %v", m.UserName, m.AgentGroup)
			_ = c.Close()
		} else {
			log.Infof(h, "add data conn for %v from user %v, group %v", m.UserName, m.LocalAddr, m.AgentGroup)
			if err := s.zones[m.UserName][m.AgentGroup].PutProxyConn(m.AgentId, m.LocalAddr, c); err != nil {
				log.Errorf(h, "put proxy conn to agent error: %v", err)
			}
		}
	default:
		log.Errorf(h, "unknown msg type %v from %v", msg.ReqType, c.RemoteAddr())
		_ = c.Close()

	}

}

func (s *Server) isZoneExist(userName, zoneName string) (exists bool) {
	if _, userExist := s.zones[userName]; !userExist {
		s.zones[userName] = make(map[string]agent.IZone, 0)
	}
	_, exists = s.zones[userName][zoneName]
	return exists
}

func (s *Server) ListAgentInfo() []*model.AgentInfoInServer {
	res := make([]*model.AgentInfoInServer, 0)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, zones := range s.zones {
		for _, z := range zones {
			res = append(res, z.Infos()...)
		}
	}
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

func (s *Server) RegisterAgent(userName, zoneName, agentId string, c net.Conn) (isUpdate bool) {
	if _, ok := s.zones[userName]; !ok {
		s.zones[userName] = make(map[string]agent.IZone, 0)
	}
	if _, ok := s.zones[userName][zoneName]; !ok {
		s.zones[userName][zoneName] = agent.NewZone(userName, zoneName)
	}
	return s.zones[userName][zoneName].RegisterAgent(agentId, c)

}

func (s *Server) ListJoinedConns(userName, zoneName string) ([]*model.GroupConnList, error) {
	res := make([]*model.GroupConnList, 0)
	if userName != "" && zoneName != "" { //only get specified zone
		if !s.isZoneExist(userName, zoneName) {
			return nil, fmt.Errorf("no such zone %v", zoneName)
		}
		res = append(res, &model.GroupConnList{
			UserName: userName,
			ZoneName: zoneName,
			List:     s.zones[userName][zoneName].ListJoinedConns(),
		})
		return res, nil
	}
	//get all userName's group conn
	for userName, zones := range s.zones {
		for zoneName, zone := range zones {
			res = append(res, &model.GroupConnList{
				UserName: userName,
				ZoneName: zoneName,
				List:     zone.ListJoinedConns(),
			})

		}
	}
	return res, nil
}

func (s *Server) KillJoinedConnById(userName, zoneName string, id int) error {
	if zoneName == "" {
		return fmt.Errorf("zone is empty")
	}
	if !s.isZoneExist(userName, zoneName) {
		return fmt.Errorf("no such zone %v", zoneName)
	}
	return s.zones[userName][zoneName].KillJoinedConnById(id)
}

func (s *Server) FlushJoinedConns() {
	for _, zones := range s.zones {
		for _, zone := range zones {
			zone.FlushJoinedConns()
		}
	}
}

func (s *Server) UpdateProxyConfigWhiteList(userName ,zoneName string, remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	if zoneName == "" {
		return fmt.Errorf("zone is empty")
	}
	if !s.isZoneExist(userName, zoneName) {
		return fmt.Errorf("no such zone %v", zoneName)
	}
	return s.zones[userName][zoneName].UpdateProxyConfigWhiteListConfig(remotePort, localAddr, whiteCidrs, whiteListEnable)
}

func (s *Server) LoadProxyConfigFile() error {
	configs, err := conf.ParseProxyConfigFile()
	if err != nil {
		return err
	}
	if err := configs.ProxyConfigIterator(func(userName string, config *model.ProxyConfig) error {
		return s.AddProxyConfigByModel(config)
	}); err != nil {
		return err
	}
	return nil
}
