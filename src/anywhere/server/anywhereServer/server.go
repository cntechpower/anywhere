package anywhereServer

import (
	"anywhere/log"
	"anywhere/tls"
	"anywhere/util"
	_tls "crypto/tls"
	"net"
	"os"
	"strconv"
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
	agents        map[string]*Agent
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
		agents:        make(map[string]*Agent, 0),
		agentsRwMutex: sync.RWMutex{},
		ExitChan:      make(chan error, 1),
	}
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
			go s.handleNewConnection(c)

		}
	}()
}

func (s *anyWhereServer) SetProxyEnable() {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()
	s.isProxyOn = true

}

func (s *anyWhereServer) isAgentExist(id string) bool {
	if _, ok := s.agents[id]; ok {
		return true
	}
	return false
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

func (s *anyWhereServer) ListProxyConfig() {
	if s.agents == nil {
		return
	}
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"AgentId", "PORT", "Addr"})
	for _, agent := range s.agents {
		for _, proxyConfig := range agent.ProxyConfigs {
			table.Append([]string{agent.Id, proxyConfig.RemoteAddr, proxyConfig.LocalAddr})
		}
	}
	table.Render()
}

func (s *anyWhereServer) ListDataConn() {
	if s.agents == nil {
		return
	}
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "LocalAddr", "RemoteAddr", "InUsed"})
	for _, agent := range s.agents {
		for _, c := range agent.DataConn {
			table.Append([]string{agent.Id, c.GetRawConn().LocalAddr().String(), c.GetRawConn().RemoteAddr().String(), strconv.FormatBool(c.InUsed)})
		}
	}
	table.Render()
}

func (s *anyWhereServer) RegisterAgent(info *Agent) (isUpdate bool) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	isUpdate = s.isAgentExist(info.Id)
	s.agents[info.Id] = info
	return isUpdate
}

func (s *anyWhereServer) RemoveAgent(info Agent) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	delete(s.agents, info.RemoteAddr.String())
}
