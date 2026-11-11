package agent

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

// healthyTimeout is how long an agent may stay silent (no heartbeat) before it
// is considered unhealthy. Agents send a heartbeat every second.
const healthyTimeout = 30 * time.Second

type IAgent interface {
	ResetAdminConn(c net.Conn)
	AskProxyConn(proxyAddr string) error
	ConsumeProxyToken(proxyAddr, token string) bool
	Info() *model.AgentInfoInServer
	IsHealthy() bool
	LastAckRcvTime() time.Time
	SendUDPData(sessionID uint32, localAddr string, data []byte) error
	Close()
}

// Agent represents one connected agent on the server side.
//
// The admin connection may be replaced (agent reconnects) while readers of the
// old connection are still blocked in Receive. Each connection generation has
// its own context; replacing the connection cancels the previous reader so
// exactly one goroutine consumes messages at a time.
type Agent struct {
	zone       string
	id         string
	userName   string
	remoteAddr string

	mu        sync.Mutex
	adminConn *conn.WrappedConn
	cancel    context.CancelFunc

	tokenMu sync.Mutex
	// tokens maps proxyAddr -> token -> expiry. Tokens are single use and bind
	// a data connection to this authenticated control connection.
	tokens map[string]map[string]time.Time

	onUDPResp func(*model.UDPDataMsg)

	lastAckRcv  atomic.Int64 // unix nano
	lastAckSend atomic.Int64 // unix nano
}

func NewAgentInfo(userName, zoneName, agentId string, c net.Conn, onUDPResp func(*model.UDPDataMsg)) *Agent {
	a := &Agent{
		zone:       zoneName,
		id:         agentId,
		userName:   userName,
		remoteAddr: c.RemoteAddr().String(),
		tokens:     make(map[string]map[string]time.Time),
		onUDPResp:  onUDPResp,
	}
	a.lastAckRcv.Store(time.Now().UnixNano())
	a.setAdminConn(c)
	return a
}

func (a *Agent) setAdminConn(c net.Conn) {
	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	if a.adminConn == nil {
		a.adminConn = conn.NewWrappedConn(a.id, c)
	} else {
		a.adminConn.ResetConn(c)
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	a.mu.Unlock()
	a.lastAckRcv.Store(time.Now().UnixNano())
	go a.run(ctx)
}

func (a *Agent) ResetAdminConn(c net.Conn) {
	a.setAdminConn(c)
}

func (a *Agent) run(ctx context.Context) {
	h := log.NewHeader("agentAdminConn")
	msg := &model.RequestMsg{}
	for {
		if ctx.Err() != nil {
			return
		}
		if err := a.adminConn.Receive(msg); err != nil {
			if ctx.Err() != nil {
				return
			}
			h.Infof("agent %s admin conn closed: %v, waiting for reconnect", a.id, err)
			_ = a.adminConn.Close()
			return
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeatPing:
			if _, err := model.ParseHeartBeatPkg(msg.Message); err != nil {
				h.Errorf("agent %s sent a corrupted heartbeat: %v", a.id, err)
				_ = a.adminConn.Close()
				return
			}
			if err := a.adminConn.Send(model.NewHeartBeatPongMsg(
				a.adminConn.GetLocalAddr(), a.adminConn.GetRemoteAddr(), a.zone, a.id)); err != nil {
				h.Errorf("send heartbeat pong to agent %s error: %v", a.id, err)
				_ = a.adminConn.Close()
				return
			}
			now := time.Now().UnixNano()
			a.lastAckSend.Store(now)
			a.lastAckRcv.Store(now)
		case model.PkgUDPDataResp:
			m, err := model.ParseUDPDataMsg(msg.Message)
			if err != nil {
				h.Errorf("agent %s sent a corrupted udp response: %v", a.id, err)
				continue
			}
			if a.onUDPResp != nil {
				a.onUDPResp(m)
			}
		default:
			h.Errorf("agent %s sent unknown message type %s, closing admin conn", a.id, msg.ReqType)
			_ = a.adminConn.Close()
			return
		}
	}
}

func (a *Agent) AskProxyConn(proxyAddr string) error {
	token, err := newToken()
	if err != nil {
		return err
	}
	a.storeToken(proxyAddr, token)
	if err := a.adminConn.Send(model.NewTunnelBeginMsg(a.userName, a.zone, a.id, proxyAddr, token)); err != nil {
		a.dropToken(proxyAddr, token)
		return fmt.Errorf("agent %s request for new proxy conn error: %w", a.id, err)
	}
	return nil
}

func (a *Agent) storeToken(proxyAddr, token string) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	m, ok := a.tokens[proxyAddr]
	if !ok {
		m = make(map[string]time.Time)
		a.tokens[proxyAddr] = m
	}
	m[token] = time.Now().Add(constants.ProxyTokenTTLSeconds * time.Second)
}

func (a *Agent) dropToken(proxyAddr, token string) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	if m, ok := a.tokens[proxyAddr]; ok {
		delete(m, token)
		if len(m) == 0 {
			delete(a.tokens, proxyAddr)
		}
	}
}

// ConsumeProxyToken validates and burns a one time token.
func (a *Agent) ConsumeProxyToken(proxyAddr, token string) bool {
	if token == "" {
		return false
	}
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	now := time.Now()
	m, ok := a.tokens[proxyAddr]
	if !ok {
		return false
	}
	expiry, ok := m[token]
	if !ok {
		return false
	}
	delete(m, token)
	if len(m) == 0 {
		delete(a.tokens, proxyAddr)
	}
	return expiry.After(now)
}

func (a *Agent) Info() *model.AgentInfoInServer {
	return &model.AgentInfoInServer{
		UserName:    a.userName,
		ZoneName:    a.zone,
		Id:          a.id,
		RemoteAddr:  a.remoteAddr,
		LastAckRcv:  a.LastAckRcvTime(),
		LastAckSend: time.Unix(0, a.lastAckSend.Load()),
	}
}

func (a *Agent) IsHealthy() bool {
	if !a.adminConn.IsValid() {
		return false
	}
	return time.Since(a.LastAckRcvTime()) < healthyTimeout
}

func (a *Agent) LastAckRcvTime() time.Time {
	return time.Unix(0, a.lastAckRcv.Load())
}

func (a *Agent) SendUDPData(sessionID uint32, localAddr string, data []byte) error {
	return a.adminConn.Send(model.NewUDPDataMsg(sessionID, localAddr, data))
}

func (a *Agent) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.adminConn != nil {
		_ = a.adminConn.Close()
	}
}

func newToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
