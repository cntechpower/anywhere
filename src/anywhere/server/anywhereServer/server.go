package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/tls"
	"anywhere/util"
	_tls "crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/olekukonko/tablewriter"
)

type anyWhereServer struct {
	serverId      string
	serverAddr    *net.TCPAddr
	isTls         bool
	credential    *_tls.Config
	listener      net.Listener
	proxyListener []net.Listener
	proxyMutex    sync.Mutex
	isHttpOn      bool
	httpMutex     sync.Mutex
	agents        map[string]*Agent
	agentsRwMutex sync.RWMutex
	ExitChan      chan error
}

var serverInstance *anyWhereServer

func InitServerInstance(serverId, port string, isHttpOn, isTls bool) *anyWhereServer {
	addr, err := util.GetAddrByIpPort("0.0.0.0", port)
	if err != nil {
		panic(err)
	}
	serverInstance = &anyWhereServer{
		serverId:      serverId,
		serverAddr:    addr,
		proxyMutex:    sync.Mutex{},
		isHttpOn:      isHttpOn,
		isTls:         isTls,
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

func (s *anyWhereServer) checkServerInit() error {
	if s.isTls && s.credential == nil {
		return fmt.Errorf("credential is empty, but server is using tls")
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
	conn.InitConnPool(10)
	go conn.HealthyCheck(conn.HeartBeatCheckFunc)
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

func (s *anyWhereServer) isAgentExist(id string) bool {
	if _, ok := s.agents[id]; ok {
		return true
	}
	//test commit
	return false
}

func (s *anyWhereServer) ListAgentInfo() {
	if s.agents == nil {
		return
	}
	s.agentsRwMutex.RLock()
	defer s.agentsRwMutex.RUnlock()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "AgentAddr", "LastAck", "Status"})
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
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"AgentId", "Port", "Addr"})
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
	table.SetHeader([]string{"AgentId", "LocalAddr", "RemoteAddr", "Status", "InUsed", "LastAckSend", "LastACKRcv"})
	for _, agent := range s.agents {
		for _, c := range agent.DataConn {
			table.Append([]string{agent.Id, c.LocalAddr().String(), c.RemoteAddr().String(),
				c.GetStatus().String(), strconv.FormatBool(c.InUsed),
				c.LastAckSendTime.Format("2006-01-02 15:04:05"), c.LastAckRcvTime.Format("2006-01-02 15:04:05")})
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
