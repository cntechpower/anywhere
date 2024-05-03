package agent

import (
	"fmt"
	"net"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
)

type IAgent interface {
	ResetAdminConn(c net.Conn)
	AskProxyConn(proxyAddr string) error
	// status
	Info() *model.AgentInfoInServer
	IsHealthy() bool
	LastAckRcvTime() time.Time
	SendUDPData(localAddr string, data []byte) (err error)
}

type Agent struct {
	zone       string
	id         string
	userName   string
	version    string
	RemoteAddr net.Addr
	adminConn  *conn.WrappedConn
	errChan    chan error
	CloseChan  chan struct{}
}

func NewAgentInfo(userName, zoneName, agentId string, c net.Conn, errChan chan error) *Agent {
	a := &Agent{
		zone:       zoneName,
		id:         agentId,
		userName:   userName,
		version:    constants.AnywhereVersion,
		RemoteAddr: c.RemoteAddr(),
		adminConn:  conn.NewWrappedConn(agentId, c),
		errChan:    errChan,
		CloseChan:  make(chan struct{}),
	}
	go a.handleAdminConnection()
	return a
}

func (a *Agent) Info() *model.AgentInfoInServer {
	return &model.AgentInfoInServer{
		UserName:    a.userName,
		ZoneName:    a.zone,
		Id:          a.id,
		RemoteAddr:  a.RemoteAddr.String(),
		LastAckRcv:  a.adminConn.LastAckRcvTime,
		LastAckSend: a.adminConn.LastAckSendTime,
	}
}

func (a *Agent) ResetAdminConn(c net.Conn) {
	a.adminConn.ResetConn(c)
	go a.handleAdminConnection()
}

func (a *Agent) AskProxyConn(proxyAddr string) error {
	h := log.NewHeader("agent.requestNewProxyConn")
	if err := a.adminConn.Send(model.NewTunnelBeginMsg(a.userName, a.zone, a.id, proxyAddr)); err != nil {
		errMsg := fmt.Errorf("agent %v request for new proxy conn error %v", a.id, err)
		log.Errorf(h, "%v", err)
		return errMsg
	}
	return nil
}

func (a *Agent) handleAdminConnection() {
	h := log.NewHeader("handleAdminConnection")
	if !a.adminConn.IsValid() {
		log.Errorf(h, "agent %v admin connection is invalid, skip handle loop", a.id)
		return
	}
	defer func() {
		// handleAdminConnection will not exit in normal
		// when handleAdminConnection there is always error happen.
		// so we need close adminConn and wait client reconnect.
		log.Warnf(h, "handleAdminConnection for %v closed", a.id)
		_ = a.adminConn.Close()
	}()
	msg := &model.RequestMsg{}
	for {
		if err := a.adminConn.Receive(&msg); err != nil {
			if err == conn.ErrNilConn {
				log.Errorf(h, "receive from agent %v admin conn error: %v, wait client reconnecting", a.id, err)
			} else {
				log.Errorf(h, "receive from agent %v admin conn error: %v, will close this connection.", a.id, err)
				_ = a.adminConn.Close()
			}
			// TODO: make this configurable
			time.Sleep(5 * time.Second)
		}
		switch msg.ReqType {
		case model.PkgReqHeartBeatPing:
			m, err := model.ParseHeartBeatPkg(msg.Message)
			if err != nil {
				log.Errorf(h, "got corrupted heartbeat ping packet from agent %v admin conn, will close it", a.id)
				return
			}
			if err := a.adminConn.Send(model.NewHeartBeatPongMsg(a.adminConn.GetLocalAddr(), a.adminConn.GetRemoteAddr(), a.zone, a.id)); err != nil {
				log.Errorf(h, "send pong msg to %v admin conn error, will close it", a.id)
				return
			} else {
				a.adminConn.SetAck(m.SendTime, time.Now())
			}

		default:
			log.Errorf(h, "got unknown ReqType: %v ,body: %v, will close admin conn", msg.ReqType, msg.Message)
			return
		}
	}
}

func (a *Agent) IsHealthy() bool {
	return a.adminConn.IsValid()
}

func (a *Agent) LastAckRcvTime() time.Time {
	return a.adminConn.LastAckRcvTime
}

func (a *Agent) SendUDPData(localAddr string, data []byte) (err error) {
	return a.adminConn.Send(model.NewUDPDataMsg(localAddr, a.zone, a.id, data))
}
