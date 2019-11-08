package anywhereServer

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/util"
	"fmt"
	"net"
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
		proxyConfigChan: make(chan *proxyConfig, 1),
		errChan:         errChan,
	}
	return a
}

func (a *Agent) requestNewProxyConn(localAddr string) {
	p := model.NewTunnelBeginMsg(a.Id, localAddr)
	pkg := model.NewRequestMsg(a.version, model.PkgTunnelBegin, a.Id, "", p)
	if err := a.AdminConn.Send(pkg); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Id, err)
		log.Error("%v", err)
		a.errChan <- errMsg
	}
}

func (a *Agent) ProxyConfigHandleLoop() {
	for p := range a.proxyConfigChan {
		go a.proxyConfigHandler(p)
	}
}

func (a *Agent) proxyConfigHandler(config *proxyConfig) {
	ln, err := util.ListenTcp(config.RemoteAddr)
	if err != nil {
		errMsg := fmt.Errorf("agent %v proxyConfigHandler got error %v", a.Id, err)
		log.Error("%v", errMsg)
		a.errChan <- errMsg
		return
	}
	go a.handelTunnelConnection(ln, config.LocalAddr, config.closeChan)
}

func (a *Agent) AddProxyConfig(config *model.ProxyConfig) {
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	log.Info("adding proxy config: %v", config)
	closeChan := make(chan struct{}, 0)
	pConfig := &proxyConfig{
		ProxyConfig: config,
		closeChan:   closeChan,
	}
	a.proxyConfigChan <- pConfig
	log.Info("add %v done", config)
	a.ProxyConfigs[config.LocalAddr] = pConfig
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
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
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
		return ErrProxyConnBufferFull
	}
}

func (a *Agent) handelTunnelConnection(ln *net.TCPListener, localAddr string, closeChan chan struct{}) {
	go func() {
		<-closeChan
		_ = ln.Close()
	}()
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Info("removed proxy config %v, %v", ln.Addr(), localAddr)
			return
		}
		go a.handelProxyConnection(c, localAddr)

	}
}

func (a *Agent) handelProxyConnection(c net.Conn, localAddr string) {

	dst, err := a.GetProxyConn(localAddr)
	if err != nil {
		log.Error("get conn error: %v", err)
		_ = c.Close()
		return
	}
	conn.JoinConn(dst.Conn, c)

}
