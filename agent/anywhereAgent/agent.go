package anywhereAgent

import (
	_tls "crypto/tls"
	"net"
	"time"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/tls"
	"github.com/cntechpower/anywhere/util"
)

type Agent struct {
	id              string
	user            string
	password        string
	addr            *net.TCPAddr
	credential      *_tls.Config
	adminConn       *conn.WrappedConn
	joinedConns     *conn.JoinedConnList
	version         string
	status          string
	lastAckSendTime time.Time
	lastAckRcvTime  time.Time
}

var agentInstance *Agent

func InitAnyWhereAgent(id, ip, user, password string, port int) *Agent {
	if agentInstance != nil {
		panic("agent already init")
	}
	addr, err := util.GetAddrByIpPort(ip, port)
	if err != nil {
		panic(err)
	}
	agentInstance = &Agent{
		id:          id,
		user:        user,
		password:    password,
		addr:        addr,
		joinedConns: conn.NewJoinedConnList(),
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

func (a *Agent) Start() {
	if a.status == "RUNNING" {
		panic("try to start a agent which is already started")
	}
	a.initControlConn(1)

	heartBeatExitChan := make(chan error, 0)
	go a.ControlConnHeartBeatSendLoop(1, heartBeatExitChan)
	go a.handleAdminConnection()
}

func (a *Agent) Stop() {
	h := log.NewHeader("agentMain")
	if a.adminConn != nil {
		a.adminConn.Close()
		log.Infof(h, "Agent Stopping...")
	}
	a.status = "STOPPED"
}

func (a *Agent) ListJoinedConns() []*conn.JoinedConnListItem {
	return a.joinedConns.List()
}

func (a *Agent) KillJoinedConnById(id int) error {
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
