package agent

import (
	"context"
	_tls "crypto/tls"
	"net"
	"time"

	"github.com/cntechpower/anywhere/dao/connlist"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/tls"
	"github.com/cntechpower/anywhere/util"
	log "github.com/cntechpower/utils/log.v2"
)

type Agent struct {
	zone            string
	id              string
	user            string
	password        string
	addr            *net.TCPAddr
	credential      *_tls.Config
	adminConn       *conn.WrappedConn
	joinedConns     *connlist.JoinedConnList
	version         string
	status          string
	lastAckSendTime time.Time
	lastAckRcvTime  time.Time
}

var agentInstance *Agent

func InitAnyWhereAgent(zone, id, ip, user, password string, port int) *Agent {
	if agentInstance != nil {
		panic("agent already init")
	}
	addr, err := util.GetAddrByIpPort(ip, port)
	if err != nil {
		panic(err)
	}
	agentInstance = &Agent{
		zone:        zone,
		id:          id,
		user:        user,
		password:    password,
		addr:        addr,
		joinedConns: connlist.NewJoinedConnList(user, zone),
		version:     constants.AnywhereVersion,
		status:      "INIT",
	}
	return agentInstance
}

func (a *Agent) SetCredentials(certFile, keyFile, caFile string) error {
	tlsConfig, err := tls.ParseTlsConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}
	a.credential = tlsConfig
	return nil
}

func (a *Agent) Start(ctx context.Context) {
	if a.status == "RUNNING" {
		panic("try to start a agent which is already started")
	}
	a.initControlConn(1)

	go a.ControlConnHeartBeatSendLoop(1, ctx)
	go a.handleAdminConnection(ctx)
}

func (a *Agent) Stop() {
	fields := map[string]interface{}{
		log.FieldNameBizName: "Agent.Stop",
		"agent_id":           a.id,
	}
	if a.adminConn != nil {
		_ = a.adminConn.Close()
		log.Infof(fields, "Agent Stopping...")
	}
	a.status = "STOPPED"
}

func (a *Agent) ListJoinedConns() ([]*model.JoinedConnListItem, error) {
	return a.joinedConns.List()
}

func (a *Agent) KillJoinedConnById(id uint) error {
	return a.joinedConns.KillById(id)
}

func (a *Agent) FlushJoinedConns() {
	a.joinedConns.Flush()
}

func (a *Agent) GetStatus() model.AgentInfoInAgent {
	return model.AgentInfoInAgent{
		Id:          a.id,
		LocalAddr:   a.adminConn.GetLocalAddr(),
		ServerAddr:  a.adminConn.GetRemoteAddr(),
		LastAckSend: a.lastAckSendTime.Format(constants.DefaultTimeFormat),
		LastAckRcv:  a.lastAckRcvTime.Format(constants.DefaultTimeFormat),
	}
}
