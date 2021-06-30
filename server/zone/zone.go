package zone

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cntechpower/anywhere/dao/connlist"

	"github.com/cntechpower/anywhere/server/zone/agent"

	"github.com/cntechpower/anywhere/dao/config"

	"github.com/cntechpower/anywhere/dao/whitelist"

	"github.com/cntechpower/anywhere/server/api/auth"

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
	ListJoinedConns() ([]*model.JoinedConnListItem, error)
	KillJoinedConnById(id uint) error
	FlushJoinedConns()
	GetCurrentConnectionCount() (int64, error)
	UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error
	//status
	Infos() []*model.AgentInfoInServer
	Info() *model.ZoneInfo
	GetProxyConfigCount() int
	ListProxyConfigs() []*model.ProxyConfig

	PutProxyConn(fromAgentId, localAddr string, c net.Conn) error
}

type Zone struct {
	zoneName         string
	userName         string
	agentsRwMutex    sync.RWMutex
	agents           map[string]agent.IAgent
	proxyConfigs     map[string]*ProxyConfigStats
	proxyConfigMutex sync.Mutex
	connectionPool   conn.ConnectionPool
	errChan          chan error
	CloseChan        chan struct{}
	joinedConns      *connlist.JoinedConnList
	connectCount     uint64
}

func NewZone(userName, zoneName string) IZone {
	z := &Zone{
		userName:         userName,
		zoneName:         zoneName,
		agents:           make(map[string]agent.IAgent, 0),
		proxyConfigs:     make(map[string]*ProxyConfigStats, 0),
		proxyConfigMutex: sync.Mutex{},
		errChan:          make(chan error, 1),
		CloseChan:        make(chan struct{}, 1),
		joinedConns:      connlist.NewJoinedConnList(userName, zoneName),
		connectCount:     0,
	}
	z.connectionPool = conn.NewConnectionPool(z.requestNewProxyConn)
	_ = z.restoreProxyConfig()
	go z.houseKeepLoop()
	return z
}

func (z *Zone) houseKeepLoop() {
	ticker := time.NewTicker(time.Second * 60)
	h := log.NewHeader("houseKeepLoop")
	for range ticker.C {
		z.agentsRwMutex.Lock()
		for name, a := range z.agents {
			if a.LastAckRcvTime().Add(time.Minute * 5).Before(time.Now()) {
				h.Infof("a %v not receive ack for 5 min, will be delete", name)
				delete(z.agents, name)
			}
		}
		z.agentsRwMutex.Unlock()
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
		z.agents[agentId] = agent.NewAgentInfo(z.userName, z.zoneName, agentId, c, make(chan error, 99))
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

func (z *Zone) GetCurrentConnectionCount() (int64, error) {
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
	pConfig := &ProxyConfigStats{
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

func (z *Zone) Info() (res *model.ZoneInfo) {
	res = &model.ZoneInfo{
		UserName:    z.userName,
		ZoneName:    z.zoneName,
		AgentsCount: int64(len(z.agents)),
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
	for _, c := range z.proxyConfigs {
		//fmt.Printf("ListProxyConfigs: %v\n", c.NetworkFlowLocalToRemoteInBytes)
		//fmt.Printf("ListProxyConfigs: %v\n", c.NetworkFlowRemoteToLocalInBytes)
		tmpC := &model.ProxyConfig{
			UserName:                        z.userName,
			ZoneName:                        z.zoneName,
			RemotePort:                      c.RemotePort,
			LocalAddr:                       c.LocalAddr,
			IsWhiteListOn:                   c.IsWhiteListOn,
			WhiteCidrList:                   c.WhiteCidrList,
			NetworkFlowRemoteToLocalInBytes: c.NetworkFlowRemoteToLocalInBytes,
			NetworkFlowLocalToRemoteInBytes: c.NetworkFlowLocalToRemoteInBytes,
			ProxyConnectCount:               c.ProxyConnectCount,
			ProxyConnectRejectCount:         c.ProxyConnectRejectCount,
		}
		tmpC.ID = c.ID
		tmpC.CreatedAt = c.CreatedAt
		tmpC.UpdatedAt = c.UpdatedAt
		res = append(res, tmpC)
	}
	return res
}

func (z *Zone) UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	key := z.getProxyConfigMapKey(remotePort, localAddr)
	c, ok := z.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy c %v in zone %v", localAddr, z.zoneName)
	}
	c.acl.SetEnable(whiteListEnable)
	err := c.acl.UpdateCidrs(whiteCidrs)
	if err == nil {
		c.IsWhiteListOn = whiteListEnable
		c.WhiteCidrList = whiteCidrs
	}
	return err

}

func (z *Zone) handleAddProxyConfig(config *ProxyConfigStats) {
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

func (z *Zone) handleTunnelConnection(h *log.Header, ln *net.TCPListener, config *ProxyConfigStats) {
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
		ip := strings.Split(c.RemoteAddr().String(), ":")[0]
		if !whiteList.IpInWhiteList(ip) {
			_ = c.Close()
			log.Infof(h, "refused %v connection because it is not in white list", c.RemoteAddr())
			config.AddConnectRejectedCount(1)
			go func() {
				_ = whitelist.AddWhiteListDenyIp(config.RemotePort, config.ZoneName, config.LocalAddr, ip)
			}()
			continue
		}
		go z.handleProxyConnection(c, config.LocalAddr, onConnectionEnd)

	}
}

func (z *Zone) handleProxyConnection(c net.Conn, localAddr string, fnOnEnd func(localToRemoteBytes, remoteToLocalBytes uint64)) {
	h := log.NewHeader(fmt.Sprintf("proxy: %v->%v", c.RemoteAddr().String(), localAddr))
	h.Infof("handle new request")
	dst, err := z.connectionPool.Get(localAddr)
	if err != nil {
		log.Infof(h, "get conn error: %v", err)
		_ = c.Close()
		return
	}
	h.Infof("get conn from pool success")
	idx := z.joinedConns.Add(conn.NewWrappedConn(localAddr, c), dst)
	h.Infof("joinedConns.Add success")
	localToRemoteBytes, remoteToLocalBytes := conn.JoinConn(dst.GetConn(), c)
	fnOnEnd(localToRemoteBytes, remoteToLocalBytes)
	if err := z.joinedConns.Remove(idx); err != nil {
		log.Errorf(h, "remove conn from list error: %v", err)
	}
	log.Infof(h, "proxy conn closed")

}

func (z *Zone) chooseAgent() (a agent.IAgent) {
	h := log.NewHeader("chooseAgent")
	z.agentsRwMutex.RLock()
	defer z.agentsRwMutex.RUnlock()
	tmpList := make([]string, 0, len(z.agents))
	for id, i := range z.agents {
		if i != nil && i.IsHealthy() {
			tmpList = append(tmpList, id)
		}
	}
	if len(tmpList) == 0 {
		h.Errorf("can't find agent")
		return
	}
	var randIdx int
	if len(tmpList) == 1 {
		randIdx = 0
	} else {
		randIdx = rand.Intn(len(tmpList) - 1)
	}
	h.Infof("choose agent %v", tmpList[randIdx])
	a = z.agents[tmpList[randIdx]]

	return
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

func (z *Zone) ListJoinedConns() ([]*model.JoinedConnListItem, error) {
	return z.joinedConns.List()
}

func (z *Zone) KillJoinedConnById(id uint) error {
	return z.joinedConns.KillById(id)
}

func (z *Zone) FlushJoinedConns() {
	z.joinedConns.Flush()
}

func (z *Zone) restoreProxyConfig() error {
	header := log.NewHeader(fmt.Sprintf("restoreProxyConfig_%s_%s", z.userName, z.zoneName))
	if err := config.Iterator(func(config *model.ProxyConfig) {
		var err error
		if z.userName == config.UserName && z.zoneName == config.ZoneName {
			err = z.AddProxyConfig(config)
			header.Infof("restore config for user %v, zone %v remotePort(%v), localAddr(%v), error: %v",
				config.UserName, config.ZoneName, config.RemotePort, config.LocalAddr, err)
		}
		if err != nil {
			header.Errorf("restore config %+v error: %v", config, err)
		}
	}); err != nil {
		return err
	}
	return nil
}

func (z *Zone) PutProxyConn(fromAgentId, localAddr string, c net.Conn) error {
	return z.connectionPool.Put(localAddr, conn.NewWrappedConn(fromAgentId, c))
}
