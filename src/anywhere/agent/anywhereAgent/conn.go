package anywhereAgent

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/tls"
	_tls "crypto/tls"
	"net"
	"time"
)

func (a *Agent) mustGetTlsConnToServer() *_tls.Conn {
	h := log.NewHeader("mustGetTlsConnToServer")
	var c *_tls.Conn
	for {
		var err error
		c, err = tls.DialTlsServer(a.addr.IP.String(), a.addr.Port, a.credential)
		if err != nil {
			log.Errorf(h, "can not connect to server %v, error: %v", a.addr, err)
			// sleep 5 second and retry
			time.Sleep(5 * time.Second)
			continue
		}
		return c
	}
}

func (a *Agent) newProxyConn(localAddr string) {
	h := log.NewHeader("newProxyConn")
	dst, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Errorf(h, "error while dial to localAddr %v", err)
		return
	}
	//let server use this local connection
	c := conn.NewBaseConn(a.mustGetTlsConnToServer())
	//TODO: optimize this package generate
	p := model.NewTunnelBeginMsg(a.id, localAddr)
	pkg := model.NewRequestMsg(a.version, model.PkgTunnelBegin, a.id, "", p)
	if err := c.Send(pkg); err != nil {
		log.Errorf(h, "error while send tunnel pkg : %v", err)
		_ = c.Close()
		_ = dst.Close()
	}
	log.Infof(h, "called newProxyConn for %v", localAddr)
	idx := a.joinedConns.Add(c, conn.NewBaseConn(dst))
	conn.JoinConn(c, dst)
	a.joinedConns.Remove(idx)
}

func (a *Agent) getTlsConnToServer() (*_tls.Conn, error) {
	var c *_tls.Conn
	c, err := tls.DialTlsServer(a.addr.IP.String(), a.addr.Port, a.credential)
	if err != nil {
		return nil, err
	}
	return c, nil

}

func (a *Agent) initControlConn(dur int) {
	h := log.NewHeader("initControlConn")
CONNECT:
	c := a.mustGetTlsConnToServer()
	a.status = "INIT"
	a.adminConn = conn.NewBaseConn(c)
	if err := a.SendControlConnRegisterPkg(); err != nil {
		log.Errorf(h, "can not send register pkg to server %v, error: %v", a.addr, err)
		_ = c.Close()
		time.Sleep(time.Duration(dur) * time.Second)
		goto CONNECT
	}
	log.Infof(h, "init control connection to server %v success", a.addr)
	a.status = "RUNNING"

}

func (a *Agent) ControlConnHeartBeatSendLoop(dur int, errChan chan error) {
	h := log.NewHeader("ControlConnHeartBeatSendLoop")
	go func() {
		for {
			if err := a.sendHeartBeatPkg(); err != nil {
				_ = a.adminConn.Close()
				log.Errorf(h, "send heartbeat error: %v, sleep %v s and try again", err, dur)

			} else {
				a.adminConn.SetAck(time.Now(), time.Now())
			}
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) handleAdminConnection() {
	h := log.NewHeader("handleAdminConnection")
	if a.adminConn == nil {
		log.Errorf(h, "handle on nil admin connection")
		return
	}
	msg := &model.RequestMsg{}
	for {
		if err := a.adminConn.Receive(&msg); err != nil {
			log.Errorf(h, "receive from admin conn error: %v, call reconnecting", err)
			_ = a.adminConn.Close()
			time.Sleep(time.Second)
			a.initControlConn(1)
			continue
		}
		switch msg.ReqType {
		case model.PkgTunnelBegin:
			m, _ := model.ParseTunnelBeginPkg(msg.Message)
			log.Infof(h, "got PkgDataConnTunnel for : %v", m.LocalAddr)
			go a.newProxyConn(m.LocalAddr)
		default:
			log.Errorf(h, "got unknown ReqType: %v, message is: %v", msg.ReqType, string(msg.Message))
			_ = a.adminConn.Close()
		}
	}
}
