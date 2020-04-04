package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	_tls "crypto/tls"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	serverId      string
	serverAddr    *net.TCPAddr
	credential    *_tls.Config
	listener      net.Listener
	proxyListener []net.Listener
	proxyMutex    sync.Mutex
	httpMutex     sync.Mutex
	agents        map[string]*Agent
	agentsRwMutex sync.RWMutex
	ExitChan      chan error
	ErrChan       chan error
}

var serverInstance *Server

func GetServerInstance() *Server {
	return serverInstance
}

func InitServerInstance(serverId string, port int) *Server {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		panic(err)
	}
	serverInstance = &Server{
		serverId:      serverId,
		serverAddr:    addr,
		proxyMutex:    sync.Mutex{},
		httpMutex:     sync.Mutex{},
		agents:        make(map[string]*Agent, 0),
		agentsRwMutex: sync.RWMutex{},
		ExitChan:      make(chan error, 1),
		ErrChan:       make(chan error, 10000),
	}
	return serverInstance
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
	if err := s.checkServerInit(); err != nil {
		panic(err)
	}
	ln, err := _tls.Listen("tcp", s.serverAddr.String(), s.credential)
	if err != nil {
		panic(err)
	}
	s.listener = ln
	l := log.GetCustomLogger("anyWhereServerMainLoop")

	go func() {
		for {
			c, err := s.listener.Accept()
			if err != nil {
				l.Infof("accept conn error: %v", err)
				continue
			}
			go s.handleNewConnection(c)

		}
	}()

}

func (s *Server) isAgentExist(id string) bool {
	if _, ok := s.agents[id]; ok {
		return true
	}
	return false
}

func (s *Server) ListAgentInfo() []*model.AgentInfo {
	res := make([]*model.AgentInfo, 0)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, agent := range s.agents {
		res = append(res, &model.AgentInfo{
			Id:          agent.Id,
			RemoteAddr:  agent.RemoteAddr.String(),
			LastAckRcv:  agent.AdminConn.LastAckRcvTime.Format("2006-01-02 15:04:05"),
			LastAckSend: agent.AdminConn.LastAckSendTime.Format("2006-01-02 15:04:05"),
		})
	}
	return res
}

func (s *Server) ListProxyConfigs() []*model.ProxyConfig {
	res := make([]*model.ProxyConfig, 0)
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	for _, agent := range s.agents {
		for _, config := range agent.ProxyConfigs {
			res = append(res, &model.ProxyConfig{
				AgentId:       agent.Id,
				RemotePort:    config.RemotePort,
				LocalAddr:     config.LocalAddr,
				IsWhiteListOn: config.IsWhiteListOn,
				WhiteCidrList: config.WhiteCidrList,
			})
		}
	}
	return res
}

func (s *Server) RegisterAgent(info *Agent) (isUpdate bool) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	isUpdate = s.isAgentExist(info.Id)
	if isUpdate {
		//close(s.agents[info.Id].CloseChan)
		s.agents[info.Id].AdminConn = info.AdminConn
	} else {
		s.agents[info.Id] = info
		go s.agents[info.Id].ProxyConfigHandleLoop()
	}

	return isUpdate
}

func (s *Server) ListJoinedConns(agentId string) (map[string][]*conn.JoinedConnListItem, error) {
	res := make(map[string][]*conn.JoinedConnListItem, 0)
	if agentId != "" { //only get specified agent
		if !s.isAgentExist(agentId) {
			return nil, fmt.Errorf("no such agent id %v", agentId)
		}
		res[agentId] = s.agents[agentId].joinedConns.List()
		return res, nil
	}
	for agentId, agent := range s.agents {
		res[agentId] = agent.joinedConns.List()
	}
	return res, nil
}

func (s *Server) KillJoinedConnById(agentId string, id int) error {
	if agentId == "" {
		return fmt.Errorf("agent id is empty")
	}
	if !s.isAgentExist(agentId) {
		return fmt.Errorf("no such agent id %v", agentId)
	}
	return s.agents[agentId].joinedConns.KillById(id)
}

func (s *Server) FlushJoinedConns() {
	for _, agent := range s.agents {
		agent.joinedConns.Flush()
	}
}
