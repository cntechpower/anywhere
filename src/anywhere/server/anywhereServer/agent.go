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

var ErrTimeoutWaitingProxyConn = fmt.Errorf("timeout when waiting for proxy conn")
var ErrProxyConnBufferFull = fmt.Errorf("proxy conn buffer is full")

type Agent struct {
	Id               string
	RemoteAddr       net.Addr
	AdminConn        *conn.BaseConn
	ProxyConfigs     []*model.ProxyConfig
	proxyConfigMutex sync.Mutex
	proxyConfigChan  chan model.ProxyConfig
	chanProxyConns   map[string]chan *conn.BaseConn
	errChan          chan error
}

func NewAgentInfo(agentId string, c net.Conn, errChan chan error) *Agent {

	a := &Agent{
		Id:              agentId,
		RemoteAddr:      c.RemoteAddr(),
		AdminConn:       conn.NewBaseConn(c),
		ProxyConfigs:    make([]*model.ProxyConfig, 0),
		chanProxyConns:  make(map[string]chan *conn.BaseConn, 0),
		proxyConfigChan: make(chan model.ProxyConfig, 5),
		errChan:         errChan,
	}
	return a
}

func (a *Agent) requestNewProxyConn(localAddr string) {
	p := model.NewTunnelBeginMsg(a.Id, localAddr)
	pkg := model.NewRequestMsg("0.0.1", model.PkgTunnelBegin, a.Id, "", p)
	if err := a.AdminConn.Send(pkg); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.Id, err)
		log.Error("%v", err)
		a.errChan <- errMsg
	}
}

func (a *Agent) ProxyConfigHandleLoop() {
	for p := range a.proxyConfigChan {
		log.Info("got proxyConfigChan: %v", p)
		go a.proxyConfigHandler(p)
	}
}

func (a *Agent) proxyConfigHandler(config model.ProxyConfig) {
	ln, err := util.ListenTcp(config.RemoteAddr)
	if err != nil {
		errMsg := fmt.Errorf("agent %v proxyConfigHandler got error %v", a.Id, err)
		log.Error("%v", errMsg)
		a.errChan <- errMsg
	}
	go a.handelTunnelConnection(ln, config.LocalAddr)
}

func (a *Agent) AddProxyConfig(config *model.ProxyConfig) {
	a.proxyConfigMutex.Lock()
	defer a.proxyConfigMutex.Unlock()
	log.Info("adding %v", config)
	a.proxyConfigChan <- *config
	log.Info("add %v done", config)
	a.ProxyConfigs = append(a.ProxyConfigs, config)
}

func (a *Agent) GetProxyConn(proxyAddr string) (*conn.BaseConn, error) {

	//request for new conn when not exist
	for i := 0; i < OPERATIONRETRYCOUNT; i++ {
		if _, ok := a.chanProxyConns[proxyAddr]; !ok {
			time.Sleep(TIMEOUTSECFORCONNPOOL * time.Second)
			a.requestNewProxyConn(proxyAddr)
			continue
		} else {
			select {
			case c := <-a.chanProxyConns[proxyAddr]:
				return c, nil
			case <-time.After(TIMEOUTSECFORCONNPOOL * time.Second):
			}
		}

	}
	return nil, fmt.Errorf("timeout while waiting for proxy conn for %v", proxyAddr)

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

func (a *Agent) handelTunnelConnection(ln *net.TCPListener, localAddr string) {
	for {
		c, err := ln.AcceptTCP()
		if err != nil {
			log.Error("got conn from %v err: %v", ln.Addr(), err)
			continue
		}
		go a.tunnelHandlerWithPool(c, localAddr)

	}
}

func (a *Agent) tunnelHandlerWithPool(c net.Conn, localAddr string) {

	dst, err := a.GetProxyConn(localAddr)
	if err != nil {
		log.Error("get conn error: %v", err)
		_ = c.Close()
		return
	}
	conn.JoinConn(dst.Conn, c)

}
