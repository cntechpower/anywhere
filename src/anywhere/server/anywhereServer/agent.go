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
}
type proxyConfig struct {
	*model.ProxyConfig
	closeChan chan struct{}
}

func NewAgentInfo(agentId string, c net.Conn, errChan chan error) *Agent {

	a := &Agent{
		Id:              agentId,
		version:         "0.0.1",
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
	p := model.NewTunnelBeginMsg(a.Id, localAddr)
	pkg := model.NewRequestMsg(a.version, model.PkgTunnelBegin, a.Id, "", p)
	if err := a.AdminConn.Send(pkg); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Id, err)
		log.GetDefaultLogger().Errorf("%v", err)
		a.errChan <- errMsg
	}
}

func (a *Agent) ProxyConfigHandleLoop() {
	log.Infof("started loop for agent %v, addr %v", a.Id, a.AdminConn.RemoteAddr())
	go func() {
		<-a.CloseChan
		close(a.proxyConfigChan)
	}()
	defer log.Infof("stopped loop for agent %v, %v", a.Id, a.AdminConn.RemoteAddr())
	for p := range a.proxyConfigChan {
		go a.proxyConfigHandler(p)
	}
}

func (a *Agent) proxyConfigHandler(config *proxyConfig) {
	ln, err := util.ListenTcp("0.0.0.0:" + strconv.Itoa(config.RemotePort))
	if err != nil {
		errMsg := fmt.Errorf("agent %v proxyConfigHandler got error %v", a.Id, err)
		log.GetDefaultLogger().Errorf("%v", errMsg)
		a.errChan <- errMsg
		return
	}
	go a.handelTunnelConnection(ln, config.LocalAddr, config.closeChan, config.ProxyConfig.IsWhiteListOn, config.ProxyConfig.WhiteCidrList)
}

func (a *Agent) AddProxyConfig(config *model.ProxyConfig) error {
	if _, exist := a.ProxyConfigs[config.LocalAddr]; exist {
		return fmt.Errorf("proxy config %v is already exist in %v", config.LocalAddr, a.Id)
	}
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	log.GetDefaultLogger().Infof("adding proxy config: %v", config)
	closeChan := make(chan struct{}, 0)
	pConfig := &proxyConfig{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	a.proxyConfigChan <- pConfig
	log.GetDefaultLogger().Infof("add %v done", config)
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
	//http://10.0.0.8/self-code/anywhere/issues/15
	err := a.AdminConn.Close()
	log.Infof("get conn from agent %v proxy addr %v failed, close admin conn, err: %v", a.Id, proxyAddr, err)

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

func (a *Agent) handelTunnelConnection(ln *net.TCPListener, localAddr string, closeChan chan struct{}, isWhiteListOn bool, whiteListIps string) {
	header := fmt.Sprintf("tunnel_%v_handler", localAddr)
	closeFlag := false
	go func() {
		<-closeChan
		_ = ln.Close()
		closeFlag = true
	}()

	//always try to get a whitelist
	whiteList, err := util.NewWhiteList(whiteListIps)
	if err != nil {
		log.Errorf("%v: init white list error: %v", header, err)
		return
	}
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			if closeFlag {
				log.Infof("%v: handler closed", header)
				return
			}
			log.Infof("%v: accept new conn error: %v", header, err)
			continue
		}
		if isWhiteListOn && !whiteList.AddrInWhiteList(c.RemoteAddr().String()) {
			_ = c.Close()
			log.Infof("%v: refused %v connection because it is not in white list", header, c.RemoteAddr())
			continue
		}
		go a.handelProxyConnection(c, localAddr)

	}
}

func (a *Agent) handelProxyConnection(c net.Conn, localAddr string) {
	header := fmt.Sprintf("proxy: %v->%v", c.RemoteAddr().String(), localAddr)
	dst, err := a.GetProxyConn(localAddr)
	if err != nil {
		log.Infof("%v: get conn error: %v", header, err)
		_ = c.Close()
		return
	}
	idx := a.joinedConns.Add(conn.NewBaseConn(c), dst)
	conn.JoinConn(dst.Conn, c)
	if err := a.joinedConns.Remove(idx); err != nil {
		log.Errorf("%v: remove conn from list error: %v", header, err)
	}
	log.Infof("%v: proxy conn closed", header)

}

func (a *Agent) handleAdminConnection() {
	if a.AdminConn == nil {
		log.Errorf("agent %v admin connection is nil, skip handle loop", a.Id)
		return
	}
	msg := &model.RequestMsg{}
	for {
		if err := a.AdminConn.Receive(&msg); err != nil {
			log.Errorf("receive from agent %v admin conn error: %v, wait client reconnecting", a.Id, err)
			_ = a.AdminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeat:
			m, _ := model.ParseHeartBeatPkg(msg.Message)
			a.AdminConn.SetAck(m.SendTime, time.Now())
		default:
			log.GetDefaultLogger().Errorf("got unknown ReqType: %v ,body: %v, will close admin conn", msg.ReqType, msg.Message)
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
