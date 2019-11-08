package anywhereAgent

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/tls"
	_tls "crypto/tls"
	"fmt"
	"net"
	"time"
)

func (a *Agent) mustGetTlsConnToServer() *_tls.Conn {
	var c *_tls.Conn
	for {
		var err error
		c, err = tls.DialTlsServer(a.Addr.IP.String(), a.Addr.Port, a.credential)
		if err != nil {
			log.Error("can not connect to server %v, error: %v", a.Addr, err)
			// sleep 5 second and retry
			time.Sleep(5 * time.Second)
			continue
		}
		return c
	}
}

func (a *Agent) newProxyConn(localAddr string) {
	dst, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Error("error while dial to localAddr %v", err)
		return
	}
	c := conn.NewBaseConn(a.mustGetTlsConnToServer())
	p := model.NewTunnelBeginMsg(a.Id, localAddr)
	pkg := model.NewRequestMsg(a.version, model.PkgTunnelBegin, a.Id, "", p)
	if err := c.Send(pkg); err != nil {
		log.Error("error while send tunnel pkg : %v", err)
		_ = c.Close()
	}
	log.Info("called newProxyConn for %v", localAddr)
	conn.JoinConn(c, dst)
}

func (a *Agent) getTlsConnToServer() (*_tls.Conn, error) {
	var c *_tls.Conn
	c, err := tls.DialTlsServer(a.Addr.IP.String(), a.Addr.Port, a.credential)
	if err != nil {
		return nil, err
	}
	return c, nil

}

func (a *Agent) initControlConn(dur int) {
CONNECT:
	c := a.mustGetTlsConnToServer()
	a.status = "INIT"
	a.AdminConn = conn.NewBaseConn(c)
	if err := a.SendControlConnRegisterPkg(); err != nil {
		log.Error("can not send register pkg to server %v, error: %v", a.Addr, err)
		_ = c.Close()
		a.AdminConn = nil
		time.Sleep(time.Duration(dur) * time.Second)
		goto CONNECT
	}
	log.Info("init control connection to server %v success", a.Addr)
	a.status = "RUNNING"

}

func (a *Agent) ControlConnHeartBeatSendLoop(dur int, errChan chan error) {
	go func() {
		for {
			//check conn status first
			if a.AdminConn.GetStatus() == conn.CStatusBad {
				errMsg := fmt.Errorf("control connection not healthy, status: %v, failReason: %v", a.AdminConn.GetStatus(), a.AdminConn.GetFailReason())
				log.Error("admin status :%v", errMsg)
				_ = a.AdminConn.Close()
				errChan <- errMsg
				a.initControlConn(dur)
				return
			}

			//if conn status is ok ,generate pkg and send
			if err := a.SendHeartBeatPkg(); err != nil {
				a.AdminConn.SetBad(err.Error())
			} else {
				a.AdminConn.SetHealthy()
			}
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) handleAdminConnection() {
	if a.AdminConn == nil {
		log.Fatal("handle on nil admin connection")
	}
	msg := &model.RequestMsg{}
	for {
		if err := a.AdminConn.Receive(&msg); err != nil {
			log.Error("receive from admin conn error: %v, call reconnecting", err)
			_ = a.AdminConn.Close()
			a.initControlConn(1)
		}
		switch msg.ReqType {
		case model.PkgTunnelBegin:
			m, _ := model.ParseTunnelBeginPkg(msg.Message)
			log.Info("got PkgDataConnTunnel for : %v", m.LocalAddr)
			go a.newProxyConn(m.LocalAddr)
		default:
			log.Error("got unknown ReqType: %v, message is: %v", msg.ReqType,string(msg.Message))
			_ = a.AdminConn.Close()
		}
	}
}
