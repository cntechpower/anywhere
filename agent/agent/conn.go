package agent

import (
	"context"
	_tls "crypto/tls"
	"net"
	"time"

	"github.com/cntechpower/anywhere/util"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/tls"
	log "github.com/cntechpower/utils/log.v2"
)

func (a *Agent) mustGetTlsConnToServer() *_tls.Conn {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Agent.mustGetTlsConnToServer",
	}
	var c *_tls.Conn
	for {
		var err error
		c, err = tls.DialTlsServer(a.addr.IP.String(), a.addr.Port, a.credential)
		if err != nil {
			log.Errorf(fields, "can not connect to server %v, error: %v", a.addr, err)
			// sleep 5 second and retry
			time.Sleep(5 * time.Second)
			continue
		}
		return c
	}
}

func (a *Agent) newProxyConn(localAddr string) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Agent.newProxyConn",
		"local_addr":         localAddr,
	}
	dst, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Errorf(fields, "error while dial to localAddr %v", err)
		return
	}
	//let server use this local connection
	c := conn.NewWrappedConn("server", a.mustGetTlsConnToServer())
	if err := c.Send(model.NewTunnelBeginMsg(a.user, a.zone, a.id, localAddr)); err != nil {
		log.Errorf(fields, "error while send tunnel pkg : %v", err)
		_ = c.Close()
		_ = dst.Close()
	}
	log.Infof(fields, "called newProxyConn for %v", localAddr)
	idx := a.joinedConns.Add(nil, c, conn.NewWrappedConn("server", dst))
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "Agent.initControlConn",
	}
	if a.adminConn == nil {
		a.adminConn = conn.NewWrappedConn("", nil)
	}
CONNECT:
	c := a.mustGetTlsConnToServer()
	a.status = "INIT"
	a.adminConn.ResetConn(c)
	if err := a.SendControlConnRegisterPkg(); err != nil {
		log.Errorf(fields, "can not send register pkg to server %v, error: %v", a.addr, err)
		_ = c.Close()
		time.Sleep(time.Duration(dur) * time.Second)
		goto CONNECT
	}
	log.Infof(fields, "init control connection to server %v success", a.addr)
	a.status = "RUNNING"

}

func (a *Agent) ControlConnHeartBeatSendLoop(dur int, ctx context.Context) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Agent.ControlConnHeartBeatSendLoop",
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := a.sendHeartBeatPkg(); err != nil {
				_ = a.adminConn.Close()
				log.Errorf(fields, "send heartbeat error: %v, sleep %v s and try again", err, dur)

			} else {
				a.lastAckSendTime = time.Now()
				a.adminConn.SetAck(time.Now(), time.Now())
			}
			time.Sleep(time.Duration(dur) * time.Second)
		}
	}()

}

func (a *Agent) handleAdminConnection(ctx context.Context) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Agent.handleAdminConnection",
	}
	if !a.adminConn.IsValid() {
		log.Errorf(fields, "admin connection is invalid")
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
			log.Errorf(fields, "receive from admin conn error: %v, call reconnecting", err)
			_ = a.adminConn.Close()
			time.Sleep(time.Second)
			a.initControlConn(1)
			continue
		}
		switch msg.ReqType {
		case model.PkgTunnelBegin:
			m, _ := model.ParseTunnelBeginPkg(msg.Message)
			log.Infof(fields, "got PkgDataConnTunnel for : %v", m.LocalAddr)
			go a.newProxyConn(m.LocalAddr)
		case model.PkgReqHeartBeatPong:
			a.lastAckRcvTime = time.Now()

		case model.PkgAuthenticationFail:
			m, _ := model.ParseAuthenticationFailMsg(msg.Message)
			log.Fatalf(fields, "authentication fail: %v", m)

		case model.PkgUDPData:
			err := util.SendUDP(msg.To, msg.Message)
			if err != nil {
				log.Errorf(fields, "send udp data to %v error: %v", msg.To, err)
			}
		default:
			log.Errorf(fields, "got unknown ReqType: %v, message is: %v", msg.ReqType, string(msg.Message))
			_ = a.adminConn.Close()
		}
	}
}
