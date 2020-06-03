package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const AGENTPROXYCONNBUFFER = 10
const TIMEOUTSECFORCONNPOOL = 1
const OPERATIONRETRYCOUNT = 5

func newErrTimeoutWaitingProxyConn(s string) error {
	return fmt.Errorf("timeout while waiting for proxy conn for %v", s)
}

var ErrProxyConnBufferFull = fmt.Errorf("proxy conn buffer is full")

type Agent struct {
	Id               string
	version          string
	RemoteAddr       net.Addr
	AdminConn        *conn.BaseConn
	ProxyConfigs     map[string]*proxyConfig
	proxyConfigMutex sync.Mutex
	proxyConfigChan  chan *proxyConfig
	chanProxyConns   map[string]chan *conn.BaseConn
	errChan          chan error
	CloseChan        chan struct{}
	joinedConns      *conn.JoinedConnList
	connectCount     uint64
}
type proxyConfig struct {
	*model.ProxyConfig
	acl       *util.WhiteList
	closeChan chan struct{}
}

func (c *proxyConfig) AddNetworkFlow(remoteToLocalBytes, localToRemoteBytes uint64) {
	atomic.AddUint64(&c.NetworkFlowLocalToRemoteInBytes, localToRemoteBytes)
	atomic.AddUint64(&c.NetworkFlowRemoteToLocalInBytes, remoteToLocalBytes)
}

func (c *proxyConfig) AddConnectCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectCount, nums)
}

func (c *proxyConfig) AddConnectRejectedCount(nums uint64) {
	atomic.AddUint64(&c.ProxyConnectRejectCount, nums)
}

func NewAgentInfo(agentId string, c net.Conn, errChan chan error) *Agent {

	a := &Agent{
		Id:              agentId,
		version:         model.AnywhereVersion,
		RemoteAddr:      c.RemoteAddr(),
		AdminConn:       conn.NewBaseConn(c),
		ProxyConfigs:    make(map[string]*proxyConfig, 0),
		chanProxyConns:  make(map[string]chan *conn.BaseConn, 5),
		proxyConfigChan: make(chan *proxyConfig, 0),
		errChan:         errChan,
		CloseChan:       make(chan struct{}, 0),
		joinedConns:     conn.NewJoinedConnList(),
	}
	return a
}

func (a *Agent) requestNewProxyConn(localAddr string) {
	h := log.NewHeader("requestNewProxyConn")
	if err := a.AdminConn.Send(model.NewTunnelBeginMsg(a.Id, localAddr)); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Id, err)
		log.Errorf(h, "%v", err)
		a.errChan <- errMsg
	}
}

func (a *Agent) ProxyConfigHandleLoop() {
	h := log.NewHeader("proxyConfigHandleLoop")
	log.Infof(h, "started loop for agent %v, addr %v", a.Id, a.AdminConn.RemoteAddr())
	go func() {
		<-a.CloseChan
		close(a.proxyConfigChan)
	}()
	defer log.Infof(h, "stopped loop for agent %v, %v", a.Id, a.AdminConn.RemoteAddr())
	for p := range a.proxyConfigChan {
		go a.proxyConfigHandler(p, h)
	}
}

func (a *Agent) proxyConfigHandler(config *proxyConfig, h *log.Header) {
	ln, err := util.ListenTcp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
	if err != nil {
		errMsg := fmt.Errorf("agent %v proxyConfigHandler got error %v", a.Id, err)
		log.Errorf(h, "%v", errMsg)
		a.errChan <- errMsg
		return
	}
	go a.handelTunnelConnection(ln, config)
}

func (a *Agent) AddProxyConfig(config *model.ProxyConfig) error {
	h := log.NewHeader("AddProxyConfig")
	if _, exist := a.ProxyConfigs[config.LocalAddr]; exist {
		return fmt.Errorf("proxy config %v is already exist in %v", config.LocalAddr, a.Id)
	}
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	log.Infof(h, "adding proxy config: %v", config)
	closeChan := make(chan struct{}, 0)
	pConfig := &proxyConfig{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	a.proxyConfigChan <- pConfig
	log.Infof(h, "add %v done", config)
	a.ProxyConfigs[config.LocalAddr] = pConfig
	return nil
}

func (a *Agent) RemoveProxyConfig(localAddr string) error {
	c, ok := a.ProxyConfigs[localAddr]
	if !ok {
		return fmt.Errorf("no such proxy config")
	}
	close(c.closeChan)
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	delete(a.ProxyConfigs, localAddr)
	return nil
}

func (a *Agent) GetProxyConn(proxyAddr string) (*conn.BaseConn, error) {
	h := log.NewHeader("GetProxyConn")
	if _, ok := a.chanProxyConns[proxyAddr]; !ok {
		a.chanProxyConns[proxyAddr] = make(chan *conn.BaseConn, AGENTPROXYCONNBUFFER)
	}
	for i := 0; i < OPERATIONRETRYCOUNT; i++ {

		//request a new proxy conn
		a.requestNewProxyConn(proxyAddr)
		select {
		case c := <-a.chanProxyConns[proxyAddr]:
			return c, nil
		case <-time.After(200 * time.Millisecond):
			continue
		}
	}
	log.Infof(h, "get conn from agent %v proxy addr %v failed, maybe proxy address is down or agent is dead", a.Id, proxyAddr)

	return nil, newErrTimeoutWaitingProxyConn(proxyAddr)
}

func (a *Agent) PutProxyConn(proxyAddr string, c *conn.BaseConn) error {
	if _, ok := a.chanProxyConns[proxyAddr]; !ok {
		a.chanProxyConns[proxyAddr] = make(chan *conn.BaseConn, AGENTPROXYCONNBUFFER)
	}
	select {
	case a.chanProxyConns[proxyAddr] <- c:
		return nil
	case <-time.After(TIMEOUTSECFORCONNPOOL * time.Second):
		a.errChan <- ErrProxyConnBufferFull
		_ = c.Close()
		return ErrProxyConnBufferFull
	}
}

func (a *Agent) handelTunnelConnection(ln *net.TCPListener, config *proxyConfig) {
	h := log.NewHeader(fmt.Sprintf("tunnel_%v_handler", config.LocalAddr))
	closeFlag := false
	go func() {
		<-config.closeChan
		_ = ln.Close()
		closeFlag = true
	}()

	//always try to get a whitelist
	whiteList, err := util.NewWhiteList(config.WhiteCidrList, config.IsWhiteListOn)
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
	idx := a.joinedConns.Add(conn.NewBaseConn(c), dst)
	localToRemoteBytes, remoteToLocalBytes := conn.JoinConn(dst.Conn, c)
	fnOnEnd(localToRemoteBytes, remoteToLocalBytes)
	if err := a.joinedConns.Remove(idx); err != nil {
		log.Errorf(h, "remove conn from list error: %v", err)
	}
	log.Infof(h, "proxy conn closed")

}

func (a *Agent) handleAdminConnection() {
	h := log.NewHeader("handleAdminConnection")
	if a.AdminConn == nil {
		log.Errorf(h, "agent %v admin connection is nil, skip handle loop", a.Id)
		return
	}
	msg := &model.RequestMsg{}
	for {
		if err := a.AdminConn.Receive(&msg); err != nil {
			log.Errorf(h, "receive from agent %v admin conn error: %v, wait client reconnecting", a.Id, err)
			_ = a.AdminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeatPing:
			m, err := model.ParseHeartBeatPkg(msg.Message)
			if err != nil {
				log.Errorf(h, "got corrupted heartbeat ping packet from agent %v admin conn, will close it", a.Id)
				_ = a.AdminConn.Close()
				return
			}
			if err := a.AdminConn.Send(model.NewHeartBeatPongMsg(a.AdminConn, a.Id)); err != nil {
				log.Errorf(h, "send pong msg to %v admin conn error, will close it", a.Id)
				_ = a.AdminConn.Close()
				return
			}
			a.AdminConn.SetAck(m.SendTime, time.Now())
		default:
			log.Errorf(h, "got unknown ReqType: %v ,body: %v, will close admin conn", msg.ReqType, msg.Message)
			_ = a.AdminConn.Close()
			return
		}
	}
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

func (a *Agent) UpdateProxyConfigWhiteListConfig(localAddr, whiteCidrs string, whiteListEnable bool) error {
	config, ok := a.ProxyConfigs[localAddr]
	if !ok {
		return fmt.Errorf("no such proxy config %v in agent %v", localAddr, a.Id)
	}
	config.acl.SetEnable(whiteListEnable)
	err := config.acl.AddCidrToList(whiteCidrs, true)
	if err == nil {
		config.IsWhiteListOn = whiteListEnable
		config.WhiteCidrList = whiteCidrs
	}
	return err

}

func (a *Agent) AddProxyConfigWhiteListConfig(localAddr, whiteCidrs string) error {
	config, ok := a.ProxyConfigs[localAddr]
	if !ok {
		return fmt.Errorf("no such proxy config %v in agent %v", localAddr, a.Id)
	}
	return config.acl.AddCidrToList(whiteCidrs, false)

}
