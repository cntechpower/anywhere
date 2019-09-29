package anywhereServer

import (
	"anywhere/log"
	"anywhere/tls"
	"anywhere/util"
	_tls "crypto/tls"
	"net"
	"os"
	"sync"

	"github.com/olekukonko/tablewriter"
)

type anyWhereServer struct {
	serverId      string
	serverAddr    *net.TCPAddr
	credential    *_tls.Config
	listener      net.Listener
	isProxyOn     bool
	proxyMutex    sync.Mutex
	isHttpOn      bool
	httpMutex     sync.Mutex
	agents        map[string]*AgentInfo
	agentsRwMutex sync.RWMutex
	ExitChan      chan error
}

var serverInstance *anyWhereServer

func InitServerInstance(serverId, port string, isProxyOn, isHttpOn bool) *anyWhereServer {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		panic(err)
	}
	serverInstance = &anyWhereServer{
		serverId:      serverId,
		serverAddr:    addr,
		isProxyOn:     isProxyOn,
		proxyMutex:    sync.Mutex{},
		isHttpOn:      isHttpOn,
		httpMutex:     sync.Mutex{},
		agents:        make(map[string]*AgentInfo, 0),
		agentsRwMutex: sync.RWMutex{},
		ExitChan:      make(chan error, 1),
	}
	return serverInstance
}

func GetServerInstance() *anyWhereServer {
	return serverInstance
}

func (s *anyWhereServer) SetCredentials(certFile, keyFile, caFile string) error {
	tlsConfig, err := tls.ParseTlsConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}
	s.credential = tlsConfig
	return nil
}

func (s *anyWhereServer) Start() {
	if s.credential == nil || s.serverId == "" {
		panic("server not init")
	}
	ln, err := _tls.Listen("tcp", s.serverAddr.String(), s.credential)
	if err != nil {
		panic(err)
	}
	s.listener = ln

	go func() {
		for {
			c, err := s.listener.Accept()
			if err != nil {
				log.Error("accept c error: %v", err)
				continue
			}
			go s.handleNewConnection(c, func(c net.Conn, err2 error) {
				log.Error("handel connection error: %v", err2.Error())
				_ = c.Close()
			})

		}
	}()
}

func (s *anyWhereServer) SetProxyEnable() {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()
	s.isProxyOn = true

}

func (s *anyWhereServer) ListAgentInfo() {
	if s.agents == nil {
		return
	}
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Addr", "LastAck", "Status"})
	for _, agent := range s.agents {
		table.Append([]string{agent.Id, agent.RemoteAddr.String(), agent.AdminConn.LastAckRcvTime.Format("2006-01-02 15:04:05"), agent.AdminConn.GetStatus().String()})
	}
	table.Render()
}

func (s *anyWhereServer) RegisterAgent(info *AgentInfo) (isUpdate bool) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if _, ok := s.agents[info.Id]; ok {
		isUpdate = true
	}
	s.agents[info.Id] = info
	return isUpdate
}

func (s *anyWhereServer) RemoveAgent(info AgentInfo) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	delete(s.agents, info.RemoteAddr.String())
}
