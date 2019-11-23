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
	var c *_tls.Conn
	for {
		var err error
		c, err = tls.DialTlsServer(a.Addr.IP.String(), a.Addr.Port, a.credential)
		if err != nil {
			log.GetDefaultLogger().Errorf("can not connect to server %v, error: %v", a.Addr, err)
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
		log.GetDefaultLogger().Errorf("error while dial to localAddr %v", err)
		return
	}
	c := conn.NewBaseConn(a.mustGetTlsConnToServer())
	p := model.NewTunnelBeginMsg(a.Id, localAddr)
	pkg := model.NewRequestMsg(a.version, model.PkgTunnelBegin, a.Id, "", p)
	if err := c.Send(pkg); err != nil {
		log.GetDefaultLogger().Errorf("error while send tunnel pkg : %v", err)
		_ = c.Close()
	}
	log.GetDefaultLogger().Infof("called newProxyConn for %v", localAddr)
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
		log.GetDefaultLogger().Errorf("can not send register pkg to server %v, error: %v", a.Addr, err)
		_ = c.Close()
		time.Sleep(time.Duration(dur) * time.Second)
		goto CONNECT
	}
	log.GetDefaultLogger().Infof("init control connection to server %v success", a.Addr)
	a.status = "RUNNING"

}

func (a *Agent) ControlConnHeartBeatSendLoop(dur int, errChan chan error) {
	l := log.GetCustomLogger("agent_heartBeater")
	go func() {
		for {
			if err := a.SendHeartBeatPkg(); err != nil {
				_ = a.AdminConn.Close()
				l.Error("send heartbeat error: %v, sleep %v s and try again", err, dur)

			} else {
				a.AdminConn.SetAck(time.Now(), time.Now())
			}
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) handleAdminConnection() {
	if a.AdminConn == nil {
		log.GetDefaultLogger().Errorf("handle on nil admin connection")
		return
	}
	msg := &model.RequestMsg{}
	for {
		if err := a.AdminConn.Receive(&msg); err != nil {
			log.GetDefaultLogger().Errorf("receive from admin conn error: %v, call reconnecting", err)
			_ = a.AdminConn.Close()
			time.Sleep(time.Second)
			a.initControlConn(1)
			continue
		}
		switch msg.ReqType {
		case model.PkgTunnelBegin:
			m, _ := model.ParseTunnelBeginPkg(msg.Message)
			log.GetDefaultLogger().Infof("got PkgDataConnTunnel for : %v", m.LocalAddr)
			go a.newProxyConn(m.LocalAddr)
		default:
			log.GetDefaultLogger().Errorf("got unknown ReqType: %v, message is: %v", msg.ReqType, string(msg.Message))
			_ = a.AdminConn.Close()
		}
	}
}
