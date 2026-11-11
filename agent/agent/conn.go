package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/model"
)

const (
	dialTimeout       = 5 * time.Second
	handshakeTimeout  = 5 * time.Second
	heartbeatInterval = time.Second
	tunnelDialTimeout = 10 * time.Second
)

func (a *Agent) setStatus(status string) {
	a.statusMu.Lock()
	a.status = status
	a.statusMu.Unlock()
}

// sleep waits for d or until the agent stops. It returns false when stopped.
func (a *Agent) sleep(d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-a.ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// controlLoop keeps exactly one control connection alive. Reconnection is
// handled here only, so there can never be two goroutines competing for the
// same admin connection.
func (a *Agent) controlLoop() {
	defer a.wg.Done()
	h := log.NewHeader("agentControlLoop")
	wait := reconnectMinWait
	for {
		if a.ctx.Err() != nil {
			return
		}
		if err := a.connectOnce(); err != nil {
			h.Errorf("connect to server %s failed: %v, retry in %v", a.addr, err, wait)
			if !a.sleep(wait) {
				return
			}
			wait *= 2
			if wait > reconnectMaxWait {
				wait = reconnectMaxWait
			}
			continue
		}
		wait = reconnectMinWait
		a.runControlConn()
		if a.ctx.Err() != nil {
			return
		}
		h.Infof("control connection to %s lost, reconnecting", a.addr)
		if !a.sleep(reconnectMinWait) {
			return
		}
	}
}

func (a *Agent) connectOnce() error {
	c, err := a.dialServer()
	if err != nil {
		return err
	}
	a.adminConn.ResetConn(c)
	a.setStatus(statusInit)
	if err := a.adminConn.Send(model.NewAgentRegisterMsg(a.zone, a.id, a.user, a.password)); err != nil {
		_ = c.Close()
		return fmt.Errorf("send register pkg: %w", err)
	}
	now := time.Now().UnixNano()
	a.lastAckSend.Store(now)
	a.lastAckRcv.Store(now)
	a.setStatus(statusRunning)
	log.Infof(log.NewHeader("agent"), "control connection to %s established", a.addr)
	return nil
}

func (a *Agent) dialServer() (*tls.Conn, error) {
	raw, err := net.DialTimeout("tcp", a.addr.String(), dialTimeout)
	if err != nil {
		return nil, err
	}
	if tc, ok := raw.(*net.TCPConn); ok {
		_ = tc.SetKeepAlive(true)
		_ = tc.SetKeepAlivePeriod(30 * time.Second)
	}
	tlsConn := tls.Client(raw, a.credential)
	ctx, cancel := context.WithTimeout(a.ctx, handshakeTimeout)
	defer cancel()
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("tls handshake: %w", err)
	}
	return tlsConn, nil
}

func (a *Agent) runControlConn() {
	connCtx, cancel := context.WithCancel(a.ctx)
	defer cancel()

	var hb sync.WaitGroup
	hb.Add(1)
	go func() {
		defer hb.Done()
		a.heartbeatLoop(connCtx)
	}()

	h := log.NewHeader("agentControlConn")
	msg := &model.RequestMsg{}
loop:
	for {
		if connCtx.Err() != nil {
			break
		}
		if err := a.adminConn.Receive(msg); err != nil {
			if connCtx.Err() == nil {
				h.Infof("receive from admin conn error: %v", err)
			}
			break
		}
		switch msg.ReqType {
		case model.PkgTunnelBegin:
			m, err := model.ParseTunnelBeginPkg(msg.Message)
			if err != nil {
				h.Errorf("corrupted tunnel begin pkg: %v", err)
				continue
			}
			go a.newProxyConn(m.LocalAddr, m.Token)
		case model.PkgReqHeartBeatPong:
			a.lastAckRcv.Store(time.Now().UnixNano())
		case model.PkgAuthenticationFail:
			m, _ := model.ParseAuthenticationFailMsg(msg.Message)
			h.Errorf("authentication failed: %s, stopping agent", m.ErrorMsg)
			a.setStatus(statusStopped)
			cancel()
			a.cancel()
			break loop
		case model.PkgUDPData:
			m, err := model.ParseUDPDataMsg(msg.Message)
			if err != nil {
				h.Errorf("corrupted udp data pkg: %v", err)
				continue
			}
			a.udp.Handle(m)
		default:
			h.Errorf("unknown message type %s", msg.ReqType)
		}
	}

	cancel()
	_ = a.adminConn.Close()
	hb.Wait()
}

func (a *Agent) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.adminConn.Send(model.NewHeartBeatPingMsg(
				a.adminConn.GetLocalAddr(), a.adminConn.GetRemoteAddr(), a.zone, a.id)); err != nil {
				log.Infof(log.NewHeader("agentHeartbeat"), "send heartbeat error: %v", err)
				// Closing unblocks the reader, which triggers a reconnect.
				_ = a.adminConn.Close()
				return
			}
			a.lastAckSend.Store(time.Now().UnixNano())
		}
	}
}

// newProxyConn opens the local service connection and attaches it to the
// server side connection pool.
func (a *Agent) newProxyConn(localAddr, token string) {
	h := log.NewHeader("agentNewProxyConn")
	dst, err := net.DialTimeout("tcp", localAddr, tunnelDialTimeout)
	if err != nil {
		h.Errorf("dial localAddr %s error: %v", localAddr, err)
		return
	}

	c, err := a.dialServer()
	if err != nil {
		h.Errorf("dial server for tunnel error: %v", err)
		_ = dst.Close()
		return
	}
	wc := conn.NewWrappedConn("server", c)
	if err := wc.Send(model.NewTunnelBeginMsg(a.user, a.zone, a.id, localAddr, token)); err != nil {
		h.Errorf("send tunnel begin pkg error: %v", err)
		_ = wc.Close()
		_ = dst.Close()
		return
	}

	idx := a.joinedConn.Add(context.Background(), wc, conn.NewWrappedConn("server", dst))
	conn.JoinConn(wc.GetConn(), dst)
	if err := a.joinedConn.Remove(idx); err != nil {
		h.Infof("remove joined conn %d: %v", idx, err)
	}
}
