package anywhereServer

import (
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	_tls "crypto/tls"
	"fmt"
	"net"
	"sync"
)

type anyWhereServer struct {
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

var serverInstance *anyWhereServer

func GetServerInstance() *anyWhereServer {
	return serverInstance
}

func InitServerInstance(serverId string, port int) *anyWhereServer {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		panic(err)
	}
	serverInstance = &anyWhereServer{
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

func (s *anyWhereServer) SetCredentials(config *_tls.Config) {
	s.credential = config
}

func (s *anyWhereServer) checkServerInit() error {
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

func (s *anyWhereServer) Start() {
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

func (s *anyWhereServer) isAgentExist(id string) bool {
	if _, ok := s.agents[id]; ok {
		return true
	}
	return false
}

func (s *anyWhereServer) ListAgentInfo() []*model.AgentInfo {
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

func (s *anyWhereServer) ListProxyConfigs() []*model.ProxyConfig {
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
				WhiteListIps:  config.WhiteListIps,
			})
		}
	}
	return res
}

func (s *anyWhereServer) RegisterAgent(info *Agent) (isUpdate bool) {
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
