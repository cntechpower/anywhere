package zone

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/cntechpower/utils/tracing"

	"github.com/cntechpower/anywhere/dao/connlist"

	"github.com/cntechpower/anywhere/server/zone/agent"

	"github.com/cntechpower/anywhere/dao/config"

	"github.com/cntechpower/anywhere/dao/whitelist"

	"github.com/cntechpower/anywhere/server/api/auth"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
	log "github.com/cntechpower/utils/log.v2"
	"github.com/opentracing/opentracing-go/ext"
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

	PutProxyConn(ctx context.Context, fromAgentId, localAddr string, c net.Conn) error
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
		agents:           make(map[string]agent.IAgent),
		proxyConfigs:     make(map[string]*ProxyConfigStats),
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.houseKeepLoop",
	}
	for range ticker.C {
		z.agentsRwMutex.Lock()
		for name, a := range z.agents {
			if a.LastAckRcvTime().Add(time.Minute * 5).Before(time.Now()) {
				log.Infof(fields, "agent %v not receive ack for 5 min, will be delete", name)
				delete(z.agents, name)
			}
		}
		z.agentsRwMutex.Unlock()
	}
}

func (z *Zone) RegisterAgent(agentId string, c net.Conn) (isUpdate bool) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.RegisterAgent",
		"agent_id":           agentId,
	}
	z.agentsRwMutex.Lock()
	a, ok := z.agents[agentId]
	isUpdate = ok
	if isUpdate {
		//close(s.agents[info.id].CloseChan)
		log.Infof(fields, "reset admin conn for user: %v, zoneName: %v, agentId: %v", z.userName, z.zoneName, agentId)
		a.ResetAdminConn(c)
	} else {
		log.Infof(fields, "build admin conn for user: %v, zoneName: %v, agentId: %v", z.userName, z.zoneName, agentId)
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.AddProxyConfig",
		"remote_port":        config.RemotePort,
		"local_addr":         config.LocalAddr,
	}
	key := z.getProxyConfigMapKey(config.RemotePort, config.LocalAddr)
	if _, exist := z.proxyConfigs[key]; exist {
		return fmt.Errorf("proxy config %v is already exist in zone  %v", key, z.zoneName)
	}
	z.proxyConfigMutex.Lock()
	defer z.proxyConfigMutex.Unlock()
	log.Infof(fields, "adding proxy config: %v", config)
	closeChan := make(chan struct{})
	pConfig := &ProxyConfigStats{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	go z.handleAddProxyConfig(pConfig)
	log.Infof(fields, "add %v done", config)
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
			ListenType:                      c.ListenType,
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.handleAddProxyConfig",
		"user_name":          config.UserName,
		"remote_port":        config.RemotePort,
		"local_addr":         config.LocalAddr,
	}
	log.Infof(fields, "starting new %v port listening", config.ListenType)

	if config.ListenType == model.ListenTypeUDP {
		ln, err := util.ListenUdp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
		if err != nil {
			errMsg := fmt.Errorf("zone %v handleAddProxyConfig got error %v", z.zoneName, err)
			log.Errorf(fields, "%v", errMsg)
			z.errChan <- errMsg
			return
		}
		go z.handleUDPTunnelConnection(fields, ln, config)
	} else {
		ln, err := util.ListenTcp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
		if err != nil {
			errMsg := fmt.Errorf("zone %v handleAddProxyConfig got error %v", z.zoneName, err)
			log.Errorf(fields, "%v", errMsg)
			z.errChan <- errMsg
			return
		}
		go z.handleTCPTunnelConnection(fields, ln, config)
	}

}

func (z *Zone) handleTCPTunnelConnection(fields map[string]interface{}, ln *net.TCPListener, config *ProxyConfigStats) {
	closeFlag := false
	go func() {
		<-config.closeChan
		_ = ln.Close()
		closeFlag = true
	}()

	//always try to get a whitelist
	whiteList, err := auth.NewWhiteListValidator(config.RemotePort, config.ZoneName, config.LocalAddr, config.WhiteCidrList, config.IsWhiteListOn)
	if err != nil {
		log.Errorf(fields, "init white list error: %v", err)
		return
	}
	config.acl = whiteList
	onConnectionEnd := func(localToRemoteBytes, remoteToLocalBytes uint64) {
		config.AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes)
		config.AddConnectCount(1)
	}
	onConnectionStart := func(span opentracing.Span) func() { return func() { span.Finish() } }
	waitTime := time.Millisecond //default error wait time 1ms
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			//if got conn error, make a limiting
			waitTime = waitTime * 2 //double wait time
			time.Sleep(waitTime)
			if closeFlag {
				log.Infof(fields, "handler closed")
				return
			}
			log.Errorf(fields, "accept new conn error: %v", err)
			continue
		}
		span, ctx := tracing.New(context.TODO(), "handleTCPTunnelConnection")
		waitTime = time.Millisecond
		ip := strings.Split(c.RemoteAddr().String(), ":")[0]
		span.SetTag("remote-ip", ip)
		span.SetTag("local-addr", config.LocalAddr)
		span.SetTag("server-port", config.RemotePort)
		span.SetTag("zone-name", config.ZoneName)
		if !whiteList.IpInWhiteList(ctx, ip) {
			_ = c.Close()
			log.InfoC(ctx, fields, "refused %v connection because it is not in white list", c.RemoteAddr())
			ext.HTTPStatusCode.Set(span, http.StatusForbidden)
			config.AddConnectRejectedCount(1)
			go func() {
				_ = whitelist.AddWhiteListDenyIp(config.RemotePort, config.UserName, config.ZoneName, config.LocalAddr, ip)
			}()
			ext.Error.Set(span, true)
			span.Finish()
			continue
		}
		go z.handleTCPProxyConnection(ctx, c, config.LocalAddr, onConnectionStart(span), onConnectionEnd)

	}
}

func (z *Zone) handleUDPTunnelConnection(fields map[string]interface{}, ln *net.UDPConn, config *ProxyConfigStats) {
	closeFlag := false
	go func() {
		<-config.closeChan
		_ = ln.Close()
		closeFlag = true
	}()

	//always try to get a whitelist
	whiteList, err := auth.NewWhiteListValidator(config.RemotePort, config.ZoneName, config.LocalAddr, config.WhiteCidrList, config.IsWhiteListOn)
	if err != nil {
		log.Errorf(fields, "init white list error: %v", err)
		return
	}
	config.acl = whiteList
	onConnectionEnd := func(localToRemoteBytes, remoteToLocalBytes uint64) {
		config.AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes)
		config.AddConnectCount(1)
	}

	waitTime := time.Millisecond //default error wait time 1ms
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := ln.ReadFromUDP(data)
		if err != nil {
			//if got conn error, make a limiting
			waitTime = waitTime * 2 //double wait time
			time.Sleep(waitTime)
			if closeFlag {
				log.Infof(fields, "handler closed")
				return
			}
			log.Errorf(fields, "accept new conn error: %v", err)
			continue
		}
		waitTime = time.Millisecond
		ip := strings.Split(remoteAddr.String(), ":")[0]
		if !whiteList.IpInWhiteList(nil, ip) {
			log.Infof(fields, "refused %v connection because it is not in white list", remoteAddr.String())
			config.AddConnectRejectedCount(1)
			go func() {
				_ = whitelist.AddWhiteListDenyIp(config.RemotePort, config.UserName, config.ZoneName, config.LocalAddr, ip)
			}()
			continue
		}
		go z.handleUDPProxyConnection(remoteAddr.String(), config.LocalAddr, data[:n], onConnectionEnd)
	}
}

func (z *Zone) handleTCPProxyConnection(ctx context.Context, c net.Conn, localAddr string, fnOnStart func(), fnOnEnd func(localToRemoteBytes, remoteToLocalBytes uint64)) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.handleTCPProxyConnection",
		"remote_addr":        c.RemoteAddr().String(),
		"local_addr":         localAddr,
	}
	log.InfoC(ctx, fields, "handle new request")
	dst, err := z.connectionPool.Get(ctx, localAddr)
	if err != nil {
		fnOnEnd(0, 0)
		log.InfoC(ctx, fields, "get conn error: %v", err)
		_ = c.Close()
		return
	}
	log.InfoC(ctx, fields, "get conn from pool success")
	idx := z.joinedConns.Add(ctx, conn.NewWrappedConn(localAddr, c), dst)
	log.InfoC(ctx, fields, "joinedConns.Add success")
	fnOnStart()
	localToRemoteBytes, remoteToLocalBytes := conn.JoinConn(dst.GetConn(), c)
	fnOnEnd(localToRemoteBytes, remoteToLocalBytes)
	if err := z.joinedConns.Remove(idx); err != nil {
		log.ErrorC(ctx, fields, "remove conn from list error: %v", err)
	}
	log.InfoC(ctx, fields, "proxy conn closed")

}

func (z *Zone) handleUDPProxyConnection(remoteAddr, localAddr string, data []byte, fnOnEnd func(localToRemoteBytes, remoteToLocalBytes uint64)) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.handleUDPProxyConnection",
		"remote_addr":        remoteAddr,
		"local_addr":         localAddr,
	}
	log.Infof(fields, "handle new request")
	a := z.chooseAgent()
	if a == nil {
		return
	}
	err := a.SendUDPData(localAddr, data)
	if err != nil {
		log.Errorf(fields, "send udp data error: %v", err)
	}
	fnOnEnd(0, uint64(len(data)))
}

func (z *Zone) chooseAgent() (a agent.IAgent) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.chooseAgent",
	}
	z.agentsRwMutex.RLock()
	defer z.agentsRwMutex.RUnlock()
	tmpList := make([]string, 0, len(z.agents))
	for id, i := range z.agents {
		if i != nil && i.IsHealthy() {
			tmpList = append(tmpList, id)
		}
	}
	if len(tmpList) == 0 {
		log.Errorf(fields, "can't find agent")
		return
	}
	var randIdx int
	if len(tmpList) == 1 {
		randIdx = 0
	} else {
		randIdx = rand.Intn(len(tmpList) - 1) //nolint:gosec
	}
	log.Infof(fields, "choose agent %v", tmpList[randIdx])
	a = z.agents[tmpList[randIdx]]

	return
}

func (z *Zone) requestNewProxyConn(localAddr string) {
	a := z.chooseAgent()
	if a == nil {
		return
	}
	if err := a.AskProxyConn(localAddr); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Info().Id, err)
		log.Errorf(nil, "%v", err)
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "Zone.restoreProxyConfig",
		"user_name":          z.userName,
		"zone_name":          z.zoneName,
	}
	if err := config.Iterator(func(config *model.ProxyConfig) {
		var err error
		if z.userName == config.UserName && z.zoneName == config.ZoneName {
			err = z.AddProxyConfig(config)
			log.Infof(fields, "restore config for user %v, zone %v remotePort(%v), localAddr(%v), error: %v",
				config.UserName, config.ZoneName, config.RemotePort, config.LocalAddr, err)
		}
		if err != nil {
			log.Errorf(fields, "restore config %+v error: %v", config, err)
		}
	}); err != nil {
		return err
	}
	return nil
}

func (z *Zone) PutProxyConn(ctx context.Context, fromAgentId, localAddr string, c net.Conn) error {
	return z.connectionPool.Put(ctx, localAddr, conn.NewWrappedConn(fromAgentId, c))
}
