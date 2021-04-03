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
	go z.houseKeepLoop()
	return z
}

func (z *Zone) houseKeepLoop() {
	ticker := time.NewTicker(time.Second * 60)
	for range ticker.C {
		z.agentsRwMutex.Lock()
		for name, agent := range z.agents {
			if agent.LastAckTime().Add(time.Minute * 5).Before(time.Now()) {
				delete(z.agents, name)
			}
		}
	}
}

func (z *Zone) RegisterAgent(agentId string, c net.Conn) (isUpdate bool) {
	h := log.NewHeader("RegisterAgent")
	z.agentsRwMutex.Lock()
	a, ok := z.agents[agentId]
	isUpdate = ok
	if isUpdate {
		//close(s.agents[info.id].CloseChan)
		h.Info("reset admin conn for user: %v, zoneName: %v, agentId: %v", z.userName, z.zoneName, agentId)
		a.ResetAdminConn(c)
	} else {
		h.Info("build admin conn for user: %v, zoneName: %v, agentId: %v", z.userName, z.zoneName, agentId)
		z.agents[agentId] = NewAgentInfo(z.userName, z.zoneName, agentId, c, make(chan error, 99))
	}
	z.agentsRwMutex.Unlock()

	return isUpdate
}

func (z *Zone) IsAgentExists(agentId string) bool {
	z.agentsRwMutex.Lock()
	defer z.agentsRwMutex.Unlock()
	_, ok := z.agents[agentId]
	return ok
}

func (z *Zone) GetCurrentConnectionCount() int {
	return z.joinedConns.Count()
}

func (z *Zone) getProxyConfigMapKey(remotePort int, localAddr string) string {
	return fmt.Sprintf("%v:%v", remotePort, localAddr)
}
func (z *Zone) AddProxyConfig(config *model.ProxyConfig) error {
	h := log.NewHeader("AddProxyConfig")
	key := z.getProxyConfigMapKey(config.RemotePort, config.LocalAddr)
	if _, exist := z.proxyConfigs[key]; exist {
		return fmt.Errorf("proxy config %v is already exist in zone  %v", key, z.zoneName)
	}
	z.proxyConfigMutex.Lock()
	defer z.proxyConfigMutex.Unlock()
	log.Infof(h, "adding proxy config: %v", config)
	closeChan := make(chan struct{}, 0)
	pConfig := &ProxyConfig{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	go z.handleAddProxyConfig(pConfig)
	log.Infof(h, "add %v done", config)
	z.proxyConfigs[key] = pConfig
	return nil
}

func (z *Zone) RemoveProxyConfig(remotePort int, localAddr string) error {
	key := z.getProxyConfigMapKey(remotePort, localAddr)
	c, ok := z.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config")
	}
	close(c.closeChan)
	z.proxyConfigMutex.Lock()
	defer z.proxyConfigMutex.Unlock()
	delete(z.proxyConfigs, key)
	return nil
}

func (z *Zone) Infos() (res []*model.AgentInfoInServer) {
	res = make([]*model.AgentInfoInServer, 0)
	for _, a := range z.agents {
		res = append(res, a.Info())
	}
	return
}

func (z *Zone) GetProxyConfigCount() int {
	z.proxyConfigMutex.Lock()
	defer z.proxyConfigMutex.Unlock()
	return len(z.proxyConfigs)
}

func (z *Zone) ListProxyConfigs() []*model.ProxyConfig {
	z.proxyConfigMutex.Lock()
	defer z.proxyConfigMutex.Unlock()
	if len(z.proxyConfigs) == 0 {
		return nil
	}
	res := make([]*model.ProxyConfig, 0, len(z.proxyConfigs))
	for _, config := range z.proxyConfigs {
		//fmt.Printf("ListProxyConfigs: %v\n", config.NetworkFlowLocalToRemoteInBytes)
		//fmt.Printf("ListProxyConfigs: %v\n", config.NetworkFlowRemoteToLocalInBytes)
		res = append(res, &model.ProxyConfig{
			UserName:                        z.userName,
			ZoneName:                        z.zoneName,
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

func (z *Zone) UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	key := z.getProxyConfigMapKey(remotePort, localAddr)
	config, ok := z.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config %v in zone %v", localAddr, z.zoneName)
	}
	config.acl.SetEnable(whiteListEnable)
	err := config.acl.AddCidrToList(whiteCidrs, true)
	if err == nil {
		config.IsWhiteListOn = whiteListEnable
		config.WhiteCidrList = whiteCidrs
	}
	return err

}

func (z *Zone) AddProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string) error {
	key := z.getProxyConfigMapKey(remotePort, localAddr)
	config, ok := z.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config %v in zone %v", localAddr, z.zoneName)
	}
	return config.acl.AddCidrToList(whiteCidrs, false)

}

func (z *Zone) handleAddProxyConfig(config *ProxyConfig) {
	h := log.NewHeader(fmt.Sprintf("tunnel-%v-(%v->%v)", config.UserName, config.RemotePort, config.LocalAddr))
	h.Infof("starting new port listening")
	ln, err := util.ListenTcp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
	if err != nil {
		errMsg := fmt.Errorf("zone %v handleAddProxyConfig got error %v", z.zoneName, err)
		log.Errorf(h, "%v", errMsg)
		z.errChan <- errMsg
		return
	}
	go z.handleTunnelConnection(h, ln, config)
}

func (z *Zone) handleTunnelConnection(h *log.Header, ln *net.TCPListener, config *ProxyConfig) {
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
		go z.handleProxyConnection(c, config.LocalAddr, onConnectionEnd)

	}
}

func (z *Zone) handleProxyConnection(c net.Conn, localAddr string, fnOnEnd func(localToRemoteBytes, remoteToLocalBytes uint64)) {
	h := log.NewHeader(fmt.Sprintf("proxy: %v->%v", c.RemoteAddr().String(), localAddr))
	dst, err := z.connectionPool.Get(localAddr)
	if err != nil {
		log.Infof(h, "get conn error: %v", err)
		_ = c.Close()
		return
	}
	idx := z.joinedConns.Add(conn.NewWrappedConn("other", c), dst)
	localToRemoteBytes, remoteToLocalBytes := conn.JoinConn(dst.GetConn(), c)
	fnOnEnd(localToRemoteBytes, remoteToLocalBytes)
	if err := z.joinedConns.Remove(idx); err != nil {
		log.Errorf(h, "remove conn from list error: %v", err)
	}
	log.Infof(h, "proxy conn closed")

}

func (z *Zone) chooseAgent() Interface {
	h := log.NewHeader("chooseAgent")
	for _, i := range z.agents {
		if i != nil && i.IsHealthy() {
			h.Infof("chosen agent %v", i.Info().Id)
			return i
		}
	}
	h.Errorf("can't find agent")
	return nil
}

func (z *Zone) requestNewProxyConn(localAddr string) {
	h := log.NewHeader("requestNewProxyConn")
	a := z.chooseAgent()
	if a == nil {
		return
	}
	if err := a.AskProxyConn(localAddr); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Info().Id, err)
		log.Errorf(h, "%v", err)
		z.errChan <- errMsg
	}
}

func (z *Zone) ListJoinedConns() []*model.JoinedConnListItem {
	return z.joinedConns.List()
}

func (z *Zone) KillJoinedConnById(id int) error {
	return z.joinedConns.KillById(id)
}

func (z *Zone) FlushJoinedConns() {
	z.joinedConns.Flush()
}

func (z *Zone) restoreProxyConfig() error {
	header := log.NewHeader(fmt.Sprintf("restoreProxyConfig_%s_%s", z.userName, z.zoneName))
	configs, err := conf.ParseProxyConfigFile()
	if err != nil {
		return err
	}
	if err := configs.ProxyConfigIterator(func(userName string, config *model.ProxyConfig) error {
		var err error
		if z.userName == config.UserName && z.zoneName == config.ZoneName {
			err = z.AddProxyConfig(config)
			header.Infof("restore config for user %v, zone %v remotePort(%v), localAddr(%v), error: %v",
				config.UserName, config.ZoneName, config.RemotePort, config.LocalAddr, err)
		}
		return err
	}); err != nil {
		return err
	}
	return nil
}

func (z *Zone) PutProxyConn(fromAgentId, localAddr string, c net.Conn) error {
	return z.connectionPool.Put(localAddr, conn.NewWrappedConn(fromAgentId, c))
}
