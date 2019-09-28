package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"anywhere/tls"
	_tls "crypto/tls"

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
			if err := c.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				log.Error("set readTimeout error: %v", err)
			}
			agent := NewAgentInfo("agent-id", "server-id", c)
			s.RegisterAgent(agent)

			go handleConnection(agent.AdminConn, func(c conn.Conn, err2 error) {
				log.Error("handel connection error: %v", err2.Error())
				c.Close()
				s.ExitChan <- fmt.Errorf("test error")
			})
		}
	}()
}

func handleConnection(c conn.Conn, funcOnError func(c conn.Conn, err error)) {
	msg := &model.RequestMsg{}
	err := c.Receive(msg)
	if err != nil {
		funcOnError(c, err)
	}
	switch msg.ReqType {
	case model.PkgReqNewproxy:
		m, _ := model.ParseProxyConfig(msg.Message)
		log.Info("got PkgReqNewproxy: %v, %v", m.RemoteAddr, m.LocalAddr)
	default:
	}
	rsp := model.NewResponseMsg(200, "got it")
	if err := c.Send(rsp); err != nil {
		funcOnError(c, err)
	}
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
	table.SetHeader([]string{"Id", "Addr", "Status"})
	for _, agent := range s.agents {
		table.Append([]string{agent.Id, agent.RemoteAddr.String(), agent.AdminConn.GetStatus().String()})
	}
	table.Render()
}

func (s *anyWhereServer) RegisterAgent(info *AgentInfo) (isUpdate bool) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	if _, ok := s.agents[info.RemoteAddr.String()]; ok {
		isUpdate = true
	}
	s.agents[info.RemoteAddr.String()] = info
	return isUpdate
}

func (s *anyWhereServer) RemoveAgent(info AgentInfo) {
	s.agentsRwMutex.Lock()
	defer s.agentsRwMutex.Unlock()
	delete(s.agents, info.RemoteAddr.String())
}
