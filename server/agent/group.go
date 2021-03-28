package agent

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cntechpower/anywhere/server/conf"

	"github.com/cntechpower/anywhere/server/auth"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
	"github.com/cntechpower/utils/log"
)

type IZone interface {
	//Agent
	IsAgentExists(agentId string) bool
	RegisterAgent(agentId string, c net.Conn) (isUpdate bool)
	AddProxyConfig(config *model.ProxyConfig) error
	RemoveProxyConfig(remotePort int, localAddr string) error
	ListJoinedConns() []*model.JoinedConnListItem
	KillJoinedConnById(id int) error
	FlushJoinedConns()
	GetCurrentConnectionCount() int
	UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error
	AddProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string) error
	//status
	Infos() []*model.AgentInfoInServer
	GetProxyConfigCount() int
	ListProxyConfigs() []*model.ProxyConfig

	PutProxyConn(fromAgentId, localAddr string, c net.Conn) error
}

type Zone struct {
	zoneName         string
	userName         string
	agentsRwMutex    sync.RWMutex
	agents           map[string]Interface
	proxyConfigs     map[string]*ProxyConfig
	proxyConfigMutex sync.Mutex
	connectionPool   conn.ConnectionPool
	errChan          chan error
	CloseChan        chan struct{}
	joinedConns      *conn.JoinedConnList
	connectCount     uint64
}

func NewZone(userName, zoneName string) IZone {
	z := &Zone{
		userName:         userName,
		zoneName:         zoneName,
		agents:           make(map[string]Interface, 0),
		proxyConfigs:     make(map[string]*ProxyConfig, 0),
		proxyConfigMutex: sync.Mutex{},
		errChan:          make(chan error, 1),
		CloseChan:        make(chan struct{}, 1),
		joinedConns:      conn.NewJoinedConnList(),
		connectCount:     0,
	}
	z.connectionPool = conn.NewConnectionPool(z.requestNewProxyConn)
	_ = z.restoreProxyConfig()
	return z
}

func (g *Zone) RegisterAgent(agentId string, c net.Conn) (isUpdate bool) {
	h := log.NewHeader("RegisterAgent")
	g.agentsRwMutex.Lock()
	a, ok := g.agents[agentId]
	isUpdate = ok
	if isUpdate {
		//close(s.agents[info.id].CloseChan)
		h.Info("reset admin conn for user: %v, zoneName: %v, agentId: %v", g.userName, g.zoneName, agentId)
		a.ResetAdminConn(c)
	} else {
		h.Info("build admin conn for user: %v, zoneName: %v, agentId: %v", g.userName, g.zoneName, agentId)
		g.agents[agentId] = NewAgentInfo(g.userName, g.zoneName, agentId, c, make(chan error, 99))
	}
	g.agentsRwMutex.Unlock()

	return isUpdate
}

func (g *Zone) IsAgentExists(agentId string) bool {
	g.agentsRwMutex.Lock()
	defer g.agentsRwMutex.Unlock()
	_, ok := g.agents[agentId]
	return ok
}

func (g *Zone) GetCurrentConnectionCount() int {
	return g.joinedConns.Count()
}

func (g *Zone) getProxyConfigMapKey(remotePort int, localAddr string) string {
	return fmt.Sprintf("%v:%v", remotePort, localAddr)
}
func (g *Zone) AddProxyConfig(config *model.ProxyConfig) error {
	h := log.NewHeader("AddProxyConfig")
	key := g.getProxyConfigMapKey(config.RemotePort, config.LocalAddr)
	if _, exist := g.proxyConfigs[key]; exist {
		return fmt.Errorf("proxy config %v is already exist in zone  %v", key, g.zoneName)
	}
	g.proxyConfigMutex.Lock()
	defer g.proxyConfigMutex.Unlock()
	log.Infof(h, "adding proxy config: %v", config)
	closeChan := make(chan struct{}, 0)
	pConfig := &ProxyConfig{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	go g.handleAddProxyConfig(pConfig)
	log.Infof(h, "add %v done", config)
	g.proxyConfigs[key] = pConfig
	return nil
}

func (g *Zone) RemoveProxyConfig(remotePort int, localAddr string) error {
	key := g.getProxyConfigMapKey(remotePort, localAddr)
	c, ok := g.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config")
	}
	close(c.closeChan)
	g.proxyConfigMutex.Lock()
	defer g.proxyConfigMutex.Unlock()
	delete(g.proxyConfigs, key)
	return nil
}

func (g *Zone) Infos() (res []*model.AgentInfoInServer) {
	res = make([]*model.AgentInfoInServer, 0)
	for _, a := range g.agents {
		res = append(res, a.Info())
	}
	return
}

func (g *Zone) GetProxyConfigCount() int {
	g.proxyConfigMutex.Lock()
	defer g.proxyConfigMutex.Unlock()
	return len(g.proxyConfigs)
}

func (g *Zone) ListProxyConfigs() []*model.ProxyConfig {
	g.proxyConfigMutex.Lock()
	defer g.proxyConfigMutex.Unlock()
	if len(g.proxyConfigs) == 0 {
		return nil
	}
	res := make([]*model.ProxyConfig, 0, len(g.proxyConfigs))
	for _, config := range g.proxyConfigs {
		//fmt.Printf("ListProxyConfigs: %v\n", config.NetworkFlowLocalToRemoteInBytes)
		//fmt.Printf("ListProxyConfigs: %v\n", config.NetworkFlowRemoteToLocalInBytes)
		res = append(res, &model.ProxyConfig{
			UserName:                        g.userName,
			ZoneName:                        g.zoneName,
			RemotePort:                      config.RemotePort,
			LocalAddr:                       config.LocalAddr,
			IsWhiteListOn:                   config.IsWhiteListOn,
			WhiteCidrList:                   config.WhiteCidrList,
			NetworkFlowRemoteToLocalInBytes: config.NetworkFlowRemoteToLocalInBytes,
			NetworkFlowLocalToRemoteInBytes: config.NetworkFlowLocalToRemoteInBytes,
			ProxyConnectCount:               config.ProxyConnectCount,
			ProxyConnectRejectCount:         config.ProxyConnectRejectCount,
		})
	}
	return res
}

func (g *Zone) UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	key := g.getProxyConfigMapKey(remotePort, localAddr)
	config, ok := g.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config %v in zone %v", localAddr, g.zoneName)
	}
	config.acl.SetEnable(whiteListEnable)
	err := config.acl.AddCidrToList(whiteCidrs, true)
	if err == nil {
		config.IsWhiteListOn = whiteListEnable
		config.WhiteCidrList = whiteCidrs
	}
	return err

}

func (g *Zone) AddProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string) error {
	key := g.getProxyConfigMapKey(remotePort, localAddr)
	config, ok := g.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config %v in zone %v", localAddr, g.zoneName)
	}
	return config.acl.AddCidrToList(whiteCidrs, false)

}

func (g *Zone) handleAddProxyConfig(config *ProxyConfig) {
	h := log.NewHeader(fmt.Sprintf("tunnel-%v-(%v->%v)", config.UserName, config.RemotePort, config.LocalAddr))
	h.Infof("starting new port listening")
	ln, err := util.ListenTcp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
	if err != nil {
		errMsg := fmt.Errorf("zone %v handleAddProxyConfig got error %v", g.zoneName, err)
		log.Errorf(h, "%v", errMsg)
		g.errChan <- errMsg
		return
	}
	go g.handleTunnelConnection(h, ln, config)
}

func (g *Zone) handleTunnelConnection(h *log.Header, ln *net.TCPListener, config *ProxyConfig) {
	closeFlag := false
	go func() {
		<-config.closeChan
		_ = ln.Close()
		closeFlag = true
	}()

	//always try to get a whitelist
	whiteList, err := auth.NewWhiteListValidator(config.RemotePort, config.ZoneName, config.LocalAddr, config.WhiteCidrList, config.IsWhiteListOn)
	if err != nil {
		log.Errorf(h, "init white list error: %v", err)
		return
	}
	config.acl = whiteList
	onConnectionEnd := func(localToRemoteBytes, remoteToLocalBytes uint64) {
		config.AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes)
		config.AddConnectCount(1)
	}
	for {
		waitTime := time.Millisecond //default error wait time 1ms
		c, err := ln.AcceptTCP()
		if err != nil {
			//if got conn error, make a limiting
			time.Sleep(waitTime)
			waitTime = waitTime * 2 //double wait time
			if closeFlag {
				log.Infof(h, "handler closed")
				return
			}
			log.Infof(h, "accept new conn error: %v", err)
			continue
		}
		waitTime = time.Millisecond
		if !whiteList.AddrInWhiteList(c.RemoteAddr().String()) {
			_ = c.Close()
			log.Infof(h, "refused %v connection because it is not in white list", c.RemoteAddr())
			config.AddConnectRejectedCount(1)
			continue
		}
		go g.handleProxyConnection(c, config.LocalAddr, onConnectionEnd)

	}
}

func (g *Zone) handleProxyConnection(c net.Conn, localAddr string, fnOnEnd func(localToRemoteBytes, remoteToLocalBytes uint64)) {
	h := log.NewHeader(fmt.Sprintf("proxy: %v->%v", c.RemoteAddr().String(), localAddr))
	dst, err := g.connectionPool.Get(localAddr)
	if err != nil {
		log.Infof(h, "get conn error: %v", err)
		_ = c.Close()
		return
	}
	idx := g.joinedConns.Add(conn.NewWrappedConn("other", c), dst)
	localToRemoteBytes, remoteToLocalBytes := conn.JoinConn(dst.GetConn(), c)
	fnOnEnd(localToRemoteBytes, remoteToLocalBytes)
	if err := g.joinedConns.Remove(idx); err != nil {
		log.Errorf(h, "remove conn from list error: %v", err)
	}
	log.Infof(h, "proxy conn closed")

}

func (g *Zone) chooseAgent() Interface {
	h := log.NewHeader("chooseAgent")
	for _, i := range g.agents {
		if i != nil && i.IsHealthy() {
			h.Infof("chosen agent %v", i.Info().Id)
			return i
		}
	}
	h.Errorf("can't find agent")
	return nil
}

func (g *Zone) requestNewProxyConn(localAddr string) {
	h := log.NewHeader("requestNewProxyConn")
	a := g.chooseAgent()
	if a == nil {
		return
	}
	if err := a.AskProxyConn(localAddr); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Info().Id, err)
		log.Errorf(h, "%v", err)
		g.errChan <- errMsg
	}
}

func (g *Zone) ListJoinedConns() []*model.JoinedConnListItem {
	return g.joinedConns.List()
}

func (g *Zone) KillJoinedConnById(id int) error {
	return g.joinedConns.KillById(id)
}

func (g *Zone) FlushJoinedConns() {
	g.joinedConns.Flush()
}

func (g *Zone) restoreProxyConfig() error {
	header := log.NewHeader(fmt.Sprintf("restoreProxyConfig_%s_%s", g.userName, g.zoneName))
	configs, err := conf.ParseProxyConfigFile()
	if err != nil {
		return err
	}
	if err := configs.ProxyConfigIterator(func(userName string, config *model.ProxyConfig) error {
		var err error
		if g.userName == config.UserName && g.zoneName == config.ZoneName {
			err = g.AddProxyConfig(config)
			header.Infof("restore config for user %v, zone %v remotePort(%v), localAddr(%v), error: %v",
				config.UserName, config.ZoneName, config.RemotePort, config.LocalAddr, err)
		}
		return err
	}); err != nil {
		return err
	}
	return nil
}

func (g *Zone) PutProxyConn(fromAgentId, localAddr string, c net.Conn) error {
	return g.connectionPool.Put(localAddr, conn.NewWrappedConn(fromAgentId, c))
}