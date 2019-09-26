package anywhereServer

import (
	"anywhere/conn"
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
	serverId         string
	serverAddr       *net.TCPAddr
	credential       *_tls.Config
	listener         net.Listener
	isProxyOn        bool
	proxyMutex       sync.Mutex
	isHttpOn         bool
	httpMutex        sync.Mutex
	agents           map[string]Agent
	agentListRwMutex sync.RWMutex
	ExitChan         chan struct{}
}

var serverInstance *anyWhereServer

func InitServerInstance(serverId string, port int, isProxyOn, isHttpOn bool) *anyWhereServer {
	addrString := fmt.Sprintf("0.0.0.0:%v", port)
	addr, _ := net.ResolveTCPAddr("tcp", addrString)
	serverInstance = &anyWhereServer{
		serverId:         serverId,
		serverAddr:       addr,
		isProxyOn:        isProxyOn,
		proxyMutex:       sync.Mutex{},
		isHttpOn:         isHttpOn,
		httpMutex:        sync.Mutex{},
		agents:           make(map[string]Agent, 0),
		agentListRwMutex: sync.RWMutex{},
		ExitChan:         make(chan struct{}, 0),
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
		fmt.Println(err)
		s.ExitChan <- struct{}{}
	}
	s.listener = ln
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				fmt.Printf("accept conn error: %v", err)
				continue
			}
			if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				fmt.Printf("set readtimeout error: %v", err)
			}
			s.RegisterAgent(Agent{
				Id:           "agent-id",
				ServerId:     s.serverId,
				Addr:         conn.RemoteAddr(),
				ProxyConfigs: nil,
			})
			go handleConnection(conn)
		}
	}()
}

func handleConnection(c net.Conn) {
	msg, err := conn.ReadRequest(c)
	if err != nil {
		fmt.Printf("read error %v", err)
	} else {
		fmt.Println(msg)
		if err := conn.SendResponse(c, 200, "Got It"); err != nil {
			fmt.Printf("send response error: %v", err)
		}
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
	s.agentListRwMutex.RLock()
	defer s.agentListRwMutex.RUnlock()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Agent_Id", "Agent_Addr"})
	for _, agent := range s.agents {
		table.Append([]string{agent.Id, agent.Addr.String()})
	}
	table.Render()
}

func (s *anyWhereServer) RegisterAgent(info Agent) (isUpdate bool) {
	s.agentListRwMutex.Lock()
	defer s.agentListRwMutex.Unlock()
	if _, ok := s.agents[info.Addr.String()]; ok {
		isUpdate = true
	}
	s.agents[info.Addr.String()] = info
	return isUpdate
}

func (s *anyWhereServer) RemoveAgent(info Agent) {
	s.agentListRwMutex.Lock()
	defer s.agentListRwMutex.Unlock()
	delete(s.agents, info.Addr.String())
}
