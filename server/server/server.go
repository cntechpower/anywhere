package server

import (
	_tls "crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/agent"
	"github.com/cntechpower/anywhere/server/auth"
	"github.com/cntechpower/anywhere/server/conf"
	"github.com/cntechpower/anywhere/util"
)

type Server struct {
	serverId                         string
	serverAddr                       *net.TCPAddr
	credential                       *_tls.Config
	listener                         net.Listener
	agents                           map[string] /*userName*/ map[string] /*agentId*/ agent.Interface
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
		agents:        make(map[string]map[string]agent.Interface, 0),
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
			_ = conn.NewWrappedConn(c).Send(model.NewAuthenticationFailMsg("validate userName and password fail"))
			_ = c.Close()
			return
		}
		if isUpdate := s.RegisterAgent(m.UserName, m.AgentId, c); isUpdate {
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
		if !s.isAgentExist(m.UserName, m.AgentId) {
			log.Errorf(h, "got data conn register pkg from unknown user %v, agent %v", m.UserName, m.AgentId)
			_ = c.Close()
		} else {
			log.Infof(h, "add data conn for %v from user %v, agent %v", m.UserName, m.LocalAddr, m.AgentId)
			if err := s.agents[m.UserName][m.AgentId].PutProxyConn(m.LocalAddr, conn.NewWrappedConn(c)); err != nil {
				log.Errorf(h, "put proxy conn to agent error: %v", err)
			}
		}
	default:
		log.Errorf(h, "unknown msg type %v from %v", msg.ReqType, c.RemoteAddr())
		_ = c.Close()

	}

}

func (s *Server) isAgentExist(userName, id string) bool {
	if _, userExist := s.agents[userName]; userExist {
		if _, agentExist := s.agents[userName][id]; agentExist {
			return true
		}
	} else {
		s.agents[userName] = make(map[string]agent.Interface, 0)
	}
	return false
}

func (s *Server) ListAgentInfo() []*model.AgentInfoInServer {
	res := make([]*model.AgentInfoInServer, 0)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, user := range s.agents {
		for _, a := range user {
			res = append(res, a.Info())
		}
	}
	return res
}

func (s *Server) ListProxyConfigs() []*model.ProxyConfig {
	// we assume that we had 100 proxy config.
	res := make([]*model.ProxyConfig, 0, 100)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, user := range s.agents {
		for _, a := range user {
			res = append(res, a.ListProxyConfigs()...)
		}
	}
	return res
}

func (s *Server) RegisterAgent(user, agentId string, c net.Conn) (isUpdate bool) {
	s.agentsRwMutex.Lock()
	isUpdate = s.isAgentExist(user, agentId)
	if isUpdate {
		//close(s.agents[info.id].CloseChan)
		s.agents[user][agentId].ResetAdminConn(c)
	} else {
		s.agents[user][agentId] = agent.NewAgentInfo(user, agentId, c, make(chan error, 99))
	}
	s.agentsRwMutex.Unlock()
	_ = s.LoadProxyConfigByAgent(log.NewHeader("RegisterAgent"), agentId)

	return isUpdate
}

func (s *Server) ListJoinedConns(user, agentId string) ([]*model.AgentConnList, error) {
	res := make([]*model.AgentConnList, 0)
	if user != "" && agentId != "" { //only get specified agent
		if !s.isAgentExist(user, agentId) {
			return nil, fmt.Errorf("no such agent id %v", agentId)
		}
		res = append(res, &model.AgentConnList{
			UserName: user,
			AgentId:  agentId,
			List:     s.agents[user][agentId].ListJoinedConns(),
		})
		return res, nil
	}
	//get all agent's conn
	for _, agents := range s.agents {
		for _, a := range agents {
			aInfo := a.Info()
			res = append(res, &model.AgentConnList{
				UserName: aInfo.UserName,
				AgentId:  aInfo.Id,
				List:     a.ListJoinedConns(),
			})
		}
	}
	return res, nil
}

func (s *Server) KillJoinedConnById(user, agentId string, id int) error {
	if agentId == "" {
		return fmt.Errorf("agent id is empty")
	}
	if !s.isAgentExist(user, agentId) {
		return fmt.Errorf("no such agent id %v", agentId)
	}
	return s.agents[agentId][agentId].KillJoinedConnById(id)
}

func (s *Server) FlushJoinedConns() {
	for _, user := range s.agents {
		for _, a := range user {
			a.FlushJoinedConns()
		}
	}
}

func (s *Server) UpdateProxyConfigWhiteList(userName string, remotePort int, agentId, localAddr, whiteCidrs string, whiteListEnable bool) error {
	if agentId == "" {
		return fmt.Errorf("agent id is empty")
	}
	if !s.isAgentExist(userName, agentId) {
		return fmt.Errorf("no such agent id %v", agentId)
	}
	return s.agents[userName][agentId].UpdateProxyConfigWhiteListConfig(remotePort, localAddr, whiteCidrs, whiteListEnable)
}

func (s *Server) LoadProxyConfigFile() error {
	configs, err := conf.ParseProxyConfigFile()
	if err != nil {
		return err
	}
	if err := configs.ProxyConfigIterator(func(userName string, config *model.ProxyConfig) error {
		return s.AddProxyConfigToAgentByModel(config)
	}); err != nil {
		return err
	}
	return nil
}

func (s *Server) LoadProxyConfigByAgent(header *log.Header, agentId string) error {
	configs, err := conf.ParseProxyConfigFile()
	if err != nil {
		return err
	}
	if err := configs.ProxyConfigIterator(func(userName string, config *model.ProxyConfig) error {
		var err error
		if agentId == config.AgentId {
			err = s.AddProxyConfigToAgentByModel(config)
			header.Infof("restore config for agent %v remotePort(%v), localAddr(%v), error: %v",
				config.AgentId, config.RemotePort, config.LocalAddr, err)
		}
		return err
	}); err != nil {
		return err
	}
	return nil
}
