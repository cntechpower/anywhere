package agent

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/util"
)

var ErrProxyConnBufferFull = fmt.Errorf("proxy conn buffer is full")

type Interface interface {
	ResetAdminConn(c net.Conn)
	AddProxyConfig(config *model.ProxyConfig) error
	RemoveProxyConfig(remotePort int, localAddr string) error
	PutProxyConn(proxyAddr string, c *conn.WrappedConn) error
	GetProxyConn(proxyAddr string) (*conn.WrappedConn, error)
	ListJoinedConns() []*conn.JoinedConnListItem
	KillJoinedConnById(id int) error
	FlushJoinedConns()
	GetCurrentConnectionCount() int
	UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error
	AddProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string) error
	//status
	Info() *model.AgentInfoInServer
	GetProxyConfigCount() int
	ListProxyConfigs() []*model.ProxyConfig
}

type ProxyConfig struct {
	*model.ProxyConfig
	acl         *util.WhiteList
	joinedConns *conn.JoinedConnList
	closeChan   chan struct{}
}

func (c *ProxyConfig) AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes uint64) {
	atomic.AddUint64(&c.NetworkFlowLocalToRemoteInBytes, localToRemoteBytes)
	atomic.AddUint64(&c.NetworkFlowRemoteToLocalInBytes, remoteToLocalBytes)
}

func (c *ProxyConfig) AddConnectCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectCount, nums)
}

func (c *ProxyConfig) AddConnectRejectedCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectRejectCount, nums)
}

func (c *ProxyConfig) GetCurrentConnectionCount() int {
	return c.joinedConns.Count()

}

type Agent struct {
	id               string
	userName         string
	version          string
	RemoteAddr       net.Addr
	adminConn        *conn.WrappedConn
	proxyConfigs     map[string]*ProxyConfig
	proxyConfigMutex sync.Mutex
	proxyConfigChan  chan *ProxyConfig
	connectionPool   conn.ConnectionPool
	errChan          chan error
	CloseChan        chan struct{}
	joinedConns      *conn.JoinedConnList
	connectCount     uint64
}

func NewAgentInfo(userName, agentId string, c net.Conn, errChan chan error) *Agent {
	a := &Agent{
		id:              agentId,
		userName:        userName,
		version:         constants.AnywhereVersion,
		RemoteAddr:      c.RemoteAddr(),
		adminConn:       conn.NewWrappedConn(c),
		proxyConfigs:    make(map[string]*ProxyConfig, 0),
		proxyConfigChan: make(chan *ProxyConfig, 0),
		errChan:         errChan,
		CloseChan:       make(chan struct{}, 0),
		joinedConns:     conn.NewJoinedConnList(),
	}
	a.connectionPool = conn.NewConnectionPool(a.requestNewProxyConn)
	go a.proxyConfigHandleLoop()
	go a.handleAdminConnection()
	return a
}

func (a *Agent) Info() *model.AgentInfoInServer {
	return &model.AgentInfoInServer{
		UserName:         a.userName,
		Id:               a.id,
		RemoteAddr:       a.RemoteAddr.String(),
		LastAckRcv:       a.adminConn.LastAckRcvTime.Format(constants.DefaultTimeFormat),
		LastAckSend:      a.adminConn.LastAckSendTime.Format(constants.DefaultTimeFormat),
		ProxyConfigCount: a.GetProxyConfigCount(),
	}
}

func (a *Agent) GetProxyConfigCount() int {
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	return len(a.proxyConfigs)
}

func (a *Agent) ListProxyConfigs() []*model.ProxyConfig {
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	if len(a.proxyConfigs) == 0 {
		return nil
	}
	res := make([]*model.ProxyConfig, 0, len(a.proxyConfigs))
	for _, config := range a.proxyConfigs {
		//fmt.Printf("ListProxyConfigs: %v\n", config.NetworkFlowLocalToRemoteInBytes)
		//fmt.Printf("ListProxyConfigs: %v\n", config.NetworkFlowRemoteToLocalInBytes)
		res = append(res, &model.ProxyConfig{
			AgentId:                         config.AgentId,
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

func (a *Agent) ResetAdminConn(c net.Conn) {
	a.adminConn.ResetConn(c)
	go a.handleAdminConnection()
}

func (a *Agent) AddProxyConfig(config *model.ProxyConfig) error {
	h := log.NewHeader("AddProxyConfig")
	key := a.getProxyConfigMapKey(config.RemotePort, config.LocalAddr)
	if _, exist := a.proxyConfigs[key]; exist {
		return fmt.Errorf("proxy config %v is already exist in %v", key, a.id)
	}
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	log.Infof(h, "adding proxy config: %v", config)
	closeChan := make(chan struct{}, 0)
	pConfig := &ProxyConfig{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	a.proxyConfigChan <- pConfig
	log.Infof(h, "add %v done", config)
	a.proxyConfigs[key] = pConfig
	return nil
}

func (a *Agent) getProxyConfigMapKey(remotePort int, localAddr string) string {
	return fmt.Sprintf("%v:%v", remotePort, localAddr)
}

func (a *Agent) RemoveProxyConfig(remotePort int, localAddr string) error {
	key := a.getProxyConfigMapKey(remotePort, localAddr)
	c, ok := a.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config")
	}
	close(c.closeChan)
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	delete(a.proxyConfigs, key)
	return nil
}

func (a *Agent) PutProxyConn(proxyAddr string, c *conn.WrappedConn) error {
	return a.connectionPool.Put(proxyAddr, c)
}

func (a *Agent) GetProxyConn(proxyAddr string) (*conn.WrappedConn, error) {
	return a.connectionPool.Get(proxyAddr)
}

func (a *Agent) ListJoinedConns() []*conn.JoinedConnListItem {
	return a.joinedConns.List()
}

func (a *Agent) KillJoinedConnById(id int) error {
	return a.joinedConns.KillById(id)
}

func (a *Agent) FlushJoinedConns() {
	a.joinedConns.Flush()
}

func (a *Agent) GetCurrentConnectionCount() int {
	count := 0
	for _, c := range a.proxyConfigs {
		count += c.GetCurrentConnectionCount()
	}
	return count
}

func (a *Agent) UpdateProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string, whiteListEnable bool) error {
	key := a.getProxyConfigMapKey(remotePort, localAddr)
	config, ok := a.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config %v in agent %v", localAddr, a.id)
	}
	config.acl.SetEnable(whiteListEnable)
	err := config.acl.AddCidrToList(whiteCidrs, true)
	if err == nil {
		config.IsWhiteListOn = whiteListEnable
		config.WhiteCidrList = whiteCidrs
	}
	return err

}

func (a *Agent) AddProxyConfigWhiteListConfig(remotePort int, localAddr, whiteCidrs string) error {
	key := a.getProxyConfigMapKey(remotePort, localAddr)
	config, ok := a.proxyConfigs[key]
	if !ok {
		return fmt.Errorf("no such proxy config %v in agent %v", localAddr, a.id)
	}
	return config.acl.AddCidrToList(whiteCidrs, false)

}

func (a *Agent) requestNewProxyConn(localAddr string) {
	h := log.NewHeader("requestNewProxyConn")
	if err := a.adminConn.Send(model.NewTunnelBeginMsg(a.userName, a.id, localAddr)); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.id, err)
		log.Errorf(h, "%v", err)
		a.errChan <- errMsg
	}
}

func (a *Agent) proxyConfigHandleLoop() {
	h := log.NewHeader("proxyConfigHandleLoop")
	log.Infof(h, "started loop for agent %v, addr %v", a.id, a.adminConn.GetRemoteAddr())
	go func() {
		<-a.CloseChan
		close(a.proxyConfigChan)
	}()
	defer log.Infof(h, "stopped loop for agent %v, %v", a.id, a.adminConn.GetRemoteAddr())
	for p := range a.proxyConfigChan {
		go a.proxyConfigHandler(p, h)
	}
}

func (a *Agent) proxyConfigHandler(config *ProxyConfig, h *log.Header) {
	ln, err := util.ListenTcp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
	if err != nil {
		errMsg := fmt.Errorf("agent %v proxyConfigHandler got error %v", a.id, err)
		log.Errorf(h, "%v", errMsg)
		a.errChan <- errMsg
		return
	}
	go a.handelTunnelConnection(ln, config)
}

func (a *Agent) handelTunnelConnection(ln *net.TCPListener, config *ProxyConfig) {
	h := log.NewHeader(fmt.Sprintf("tunnel_%v_handler", config.LocalAddr))
	closeFlag := false
	go func() {
		<-config.closeChan
		_ = ln.Close()
		closeFlag = true
	}()

	//always try to get a whitelist
	whiteList, err := util.NewWhiteList(config.RemotePort, config.AgentId, config.LocalAddr, config.WhiteCidrList, config.IsWhiteListOn)
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
		go a.handelProxyConnection(c, config.LocalAddr, onConnectionEnd)

	}
}

func (a *Agent) handelProxyConnection(c net.Conn, localAddr string, fnOnEnd func(localToRemoteBytes, remoteToLocalBytes uint64)) {
	h := log.NewHeader(fmt.Sprintf("proxy: %v->%v", c.RemoteAddr().String(), localAddr))
	dst, err := a.GetProxyConn(localAddr)
	if err != nil {
		log.Infof(h, "get conn error: %v", err)
		_ = c.Close()
		return
	}
	idx := a.joinedConns.Add(conn.NewWrappedConn(c), dst)
	localToRemoteBytes, remoteToLocalBytes := conn.JoinConn(dst.GetConn(), c)
	fnOnEnd(localToRemoteBytes, remoteToLocalBytes)
	if err := a.joinedConns.Remove(idx); err != nil {
		log.Errorf(h, "remove conn from list error: %v", err)
	}
	log.Infof(h, "proxy conn closed")

}

func (a *Agent) handleAdminConnection() {
	h := log.NewHeader("handleAdminConnection")
	if !a.adminConn.IsValid() {
		log.Errorf(h, "agent %v admin connection is invalid, skip handle loop", a.id)
		return
	}
	defer func() {
		// handleAdminConnection will not exit in normal
		// when handleAdminConnection there is always error happen.
		// so we need close adminConn and wait client reconnect.
		log.Warnf(h, "handleAdminConnection for %v closed", a.id)
		a.adminConn.Close()
	}()
	msg := &model.RequestMsg{}
	for {
		if err := a.adminConn.Receive(&msg); err != nil {
			if err == conn.ErrNilConn {
				log.Errorf(h, "receive from agent %v admin conn error: %v, wait client reconnecting", a.id, err)
			} else {
				log.Errorf(h, "receive from agent %v admin conn error: %v, will close this connection.", a.id, err)
				_ = a.adminConn.Close()
			}
			//TODO: make this configurable
			time.Sleep(5 * time.Second)
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeatPing:
			m, err := model.ParseHeartBeatPkg(msg.Message)
			if err != nil {
				log.Errorf(h, "got corrupted heartbeat ping packet from agent %v admin conn, will close it", a.id)
				return
			}
			if err := a.adminConn.Send(model.NewHeartBeatPongMsg(a.adminConn.GetLocalAddr(), a.adminConn.GetRemoteAddr(), a.id)); err != nil {
				log.Errorf(h, "send pong msg to %v admin conn error, will close it", a.id)
				return
			} else {
				a.adminConn.SetAck(m.SendTime, time.Now())
			}

		default:
			log.Errorf(h, "got unknown ReqType: %v ,body: %v, will close admin conn", msg.ReqType, msg.Message)
			return
		}
	}
}
