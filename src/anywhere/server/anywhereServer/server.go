package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/util"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	. "anywhere/model"
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
	agents        map[string]AgentInfo
	agentsRwMutex sync.RWMutex
	ExitChan      chan error
}

type AgentInfo struct {
	Id         string
	ServerId   string
	RemoteAddr net.Addr
	AdminConn  *conn.AdminConn
	DataConn   []net.Conn
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
		agents:        make(map[string]AgentInfo, 0),
		agentsRwMutex: sync.RWMutex{},
		ExitChan:      make(chan error, 0),
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
	ln, err := _tls.Listen("tcp", s.serverAddr.String(), s.credential)
	if err != nil {
		s.ExitChan <- err
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
			s.RegisterAgent(AgentInfo{
				Id:         "agent-id",
				ServerId:   s.serverId,
				RemoteAddr: c.RemoteAddr(),
				AdminConn:  conn.NewAdminConn(c),
			})
			go handleConnection(c)
		}
	}()
}

func handleConnection(c net.Conn) {
	msg, err := conn.ReadRequest(c)
	if err != nil {
		log.Error("read from %v error %v ", c.RemoteAddr().String(), err)
		fmt.Println(c.RemoteAddr().String())
		_ = c.Close()
		return
	}
	switch msg.ReqType {
	case PkgReqNewproxy:
	default:

	}

	log.Info("%v", msg)
	if err := conn.SendResponse(c, 200, "Got It"); err != nil {
		log.Error("send response error: %v", err)
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
	table.SetHeader([]string{"Agent_Id", "Agent_Addr"})
	for _, agent := range s.agents {
		table.Append([]string{agent.Id, agent.RemoteAddr.String()})
	}
	table.Render()
}

func (s *anyWhereServer) RegisterAgent(info AgentInfo) (isUpdate bool) {
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
