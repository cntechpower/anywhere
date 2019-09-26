package anywhereServer

import (
	"anywhere/conn"
	"encoding/json"
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
	serverId         string
	serverAddr       *net.TCPAddr
	credential       *_tls.Config
	listener         net.Listener
	isProxyOn        bool
	proxyMutex       sync.Mutex
	isHttpOn         bool
	httpMutex        sync.Mutex
	agentList        map[string]agentInfo
	agentListRwMutex sync.RWMutex
	ExitChan         chan struct{}
}

type agentInfo struct {
	agentId         string
	serverId        string
	agentAddr       net.Addr
	proxyConfigList []proxyConfig
}

type proxyConfig struct {
	remoteAddr net.Addr
	localAddr  net.Addr
}

var serverInstance *anyWhereServer

func InitServerInstance(serverId, port string, isProxyOn, isHttpOn bool) *anyWhereServer {
	addrString := fmt.Sprintf("0.0.0.0:%v", port)
	addr, _ := net.ResolveTCPAddr("tcp", addrString)
	serverInstance = &anyWhereServer{
		serverId:         serverId,
		serverAddr:       addr,
		isProxyOn:        isProxyOn,
		proxyMutex:       sync.Mutex{},
		isHttpOn:         isHttpOn,
		httpMutex:        sync.Mutex{},
		agentList:        make(map[string]agentInfo, 0),
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
			s.RegisterAgent(agentInfo{
				agentId:         "agent-id",
				serverId:        s.serverId,
				agentAddr:       conn.RemoteAddr(),
				proxyConfigList: nil,
			})
			go handleConnection(conn)
		}
	}()
}

func handleConnection(c net.Conn) {
	d := json.NewDecoder(c)
	var msg conn.Package
	if err := d.Decode(&msg); err != nil {
		fmt.Println("Decode Package Error")
	}
	fmt.Println(msg)
	if err := c.Close(); err != nil {
		fmt.Printf("Error Close Conn: %v\n", err)

	}
}

func (s *anyWhereServer) SetProxyEnable() {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()
	s.isProxyOn = true

}

func (s *anyWhereServer) ListAgentInfo() {
	if s.agentList == nil {
		return
	}
	s.agentListRwMutex.RLock()
	defer s.agentListRwMutex.RUnlock()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Agent_Id", "Agent_Addr"})
	for _, agent := range s.agentList {
		table.Append([]string{agent.agentId, agent.agentAddr.String()})
	}
	table.Render()
}

func (s *anyWhereServer) RegisterAgent(info agentInfo) (isUpdate bool) {
	s.agentListRwMutex.Lock()
	defer s.agentListRwMutex.Unlock()
	if _, ok := s.agentList[info.agentAddr.String()]; ok {
		isUpdate = true
	}
	s.agentList[info.agentAddr.String()] = info
	return isUpdate
}

func (s *anyWhereServer) RemoveAgent(info agentInfo) {
	s.agentListRwMutex.Lock()
	defer s.agentListRwMutex.Unlock()
	delete(s.agentList, info.agentAddr.String())
}
