package anywhereAgent

import (
	"context"
	_tls "crypto/tls"
	"net"
	"time"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/tls"
	"github.com/cntechpower/utils/log"
)

func (a *Agent) mustGetTlsConnToServer() *_tls.Conn {
	h := log.NewHeader("mustGetTlsConnToServer")
	var c *_tls.Conn
	for {
		var err error
		c, err = tls.DialTlsServer(a.addr.IP.String(), a.addr.Port, a.credential)
		if err != nil {
			h.Errorf("can not connect to server %v, error: %v", a.addr, err)
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
		h.Errorf("error while dial to localAddr %v", err)
		return
	}
	//let server use this local connection
	c := conn.NewWrappedConn("server", a.mustGetTlsConnToServer())
	if err := c.Send(model.NewTunnelBeginMsg(a.user, a.group, a.id, localAddr)); err != nil {
		h.Errorf("error while send tunnel pkg : %v", err)
		_ = c.Close()
		_ = dst.Close()
	}
	h.Infof("called newProxyConn for %v", localAddr)
	idx := a.joinedConns.Add(c, conn.NewWrappedConn("server", dst))
	conn.JoinConn(c.GetConn(), dst)
	_ = a.joinedConns.Remove(idx)
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
	if a.adminConn == nil {
		a.adminConn = conn.NewWrappedConn("", nil)
	}
CONNECT:
	c := a.mustGetTlsConnToServer()
	a.status = "INIT"
	a.adminConn.ResetConn(c)
	if err := a.SendControlConnRegisterPkg(); err != nil {
		h.Errorf("can not send register pkg to server %v, error: %v", a.addr, err)
		_ = c.Close()
		time.Sleep(time.Duration(dur) * time.Second)
		goto CONNECT
	}
	h.Infof("init control connection to server %v success", a.addr)
	a.status = "RUNNING"

}

func (a *Agent) ControlConnHeartBeatSendLoop(dur int, ctx context.Context) {
	h := log.NewHeader("ControlConnHeartBeatSendLoop")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := a.sendHeartBeatPkg(); err != nil {
				_ = a.adminConn.Close()
				h.Errorf("send heartbeat error: %v, sleep %v s and try again", err, dur)

			} else {
				a.lastAckSendTime = time.Now()
				a.adminConn.SetAck(time.Now(), time.Now())
			}
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) handleAdminConnection(ctx context.Context) {
	h := log.NewHeader("handleAdminConnection")
	if !a.adminConn.IsValid() {
		h.Errorf("admin connection is invalid")
		return
	}
	msg := &model.RequestMsg{}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := a.adminConn.Receive(&msg); err != nil {
			h.Errorf("receive from admin conn error: %v, call reconnecting", err)
			_ = a.adminConn.Close()
			time.Sleep(time.Second)
			a.initControlConn(1)
			continue
		}
		switch msg.ReqType {
		case model.PkgTunnelBegin:
			m, _ := model.ParseTunnelBeginPkg(msg.Message)
			h.Infof("got PkgDataConnTunnel for : %v", m.LocalAddr)
			go a.newProxyConn(m.LocalAddr)
		case model.PkgReqHeartBeatPong:
			a.lastAckRcvTime = time.Now()

		case model.PkgAuthenticationFail:
			m, _ := model.ParseAuthenticationFailMsg(msg.Message)
			h.Fatalf("authentication fail: %v", m)

		default:
			h.Errorf("got unknown ReqType: %v, message is: %v", msg.ReqType, string(msg.Message))
			_ = a.adminConn.Close()
		}
	}
}
