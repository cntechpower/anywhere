package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/dao/connlist"
	"github.com/cntechpower/anywhere/model"
	tlsutil "github.com/cntechpower/anywhere/tls"
)

const (
	statusInit       = "INIT"
	statusRunning    = "RUNNING"
	statusReconnect  = "RECONNECTING"
	statusStopped    = "STOPPED"
	reconnectMinWait = time.Second
	reconnectMaxWait = 30 * time.Second
)

type Agent struct {
	zone       string
	id         string
	user       string
	password   string
	addr       *net.TCPAddr
	credential *tls.Config

	adminConn  *conn.WrappedConn
	joinedConn *connlist.JoinedConnList
	udp        *udpManager

	version string

	statusMu sync.RWMutex
	status   string

	lastAckSend atomic.Int64 // unix nano
	lastAckRcv  atomic.Int64 // unix nano

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New builds an agent instance. There is no package level singleton: the
// caller owns the agent and decides when to start/stop it.
func New(zone, id, ip, user, password string, port int) (*Agent, error) {
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("invalid server address %s:%d: %w", ip, port, err)
	}
	return &Agent{
		zone:       zone,
		id:         id,
		user:       user,
		password:   password,
		addr:       addr,
		joinedConn: connlist.NewJoinedConnList(user, zone),
		version:    constants.AnywhereVersion,
		status:     statusInit,
	}, nil
}

func (a *Agent) SetCredentials(certFile, keyFile, caFile, serverName string, insecure bool) error {
	tlsConfig, err := tlsutil.ParseClientTlsConfig(certFile, keyFile, caFile, serverName, insecure)
	if err != nil {
		return err
	}
	a.credential = tlsConfig
	return nil
}

// Start connects to the server and keeps the control connection alive until ctx
// is cancelled. It returns immediately.
func (a *Agent) Start(ctx context.Context) {
	a.statusMu.Lock()
	if a.status == statusRunning {
		a.statusMu.Unlock()
		panic("try to start an agent which is already started")
	}
	a.ctx, a.cancel = context.WithCancel(ctx)
	a.adminConn = conn.NewWrappedConn("server", nil)
	a.udp = newUDPManager()
	a.udp.send = func(msg *model.RequestMsg) error { return a.adminConn.Send(msg) }
	a.status = statusInit
	a.statusMu.Unlock()

	a.wg.Add(1)
	go a.controlLoop()
}

func (a *Agent) Stop() {
	a.statusMu.Lock()
	a.status = statusStopped
	a.statusMu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
	if a.adminConn != nil {
		_ = a.adminConn.Close()
	}
	if a.udp != nil {
		a.udp.CloseAll()
	}
	a.joinedConn.Flush()
	a.wg.Wait()
	log.Infof(log.NewHeader("agentMain"), "agent %s stopped", a.id)
}

func (a *Agent) ListJoinedConn() ([]*model.JoinedConnListItem, error) {
	return a.joinedConn.List(), nil
}

func (a *Agent) KillJoinedConnById(id uint) error {
	return a.joinedConn.KillById(id)
}

func (a *Agent) FlushJoinedConn() {
	a.joinedConn.Flush()
}

func (a *Agent) GetStatus() model.AgentInfoInAgent {
	a.statusMu.RLock()
	status := a.status
	a.statusMu.RUnlock()
	if status == "" {
		status = statusInit
	}
	return model.AgentInfoInAgent{
		Id:          a.id,
		LocalAddr:   a.adminConn.GetLocalAddr(),
		ServerAddr:  a.adminConn.GetRemoteAddr(),
		LastAckSend: time.Unix(0, a.lastAckSend.Load()).Format(constants.DefaultTimeFormat),
		LastAckRcv:  time.Unix(0, a.lastAckRcv.Load()).Format(constants.DefaultTimeFormat),
	}
}
