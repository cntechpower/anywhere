package anywhereAgent

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/tls"
	"anywhere/util"
	_tls "crypto/tls"
	"net"
)

type Agent struct {
	id          string
	addr        *net.TCPAddr
	credential  *_tls.Config
	adminConn   *conn.BaseConn
	joinedConns *conn.JoinedConnList
	version     string
	status      string
}

var agentInstance *Agent

func InitAnyWhereAgent(id, ip string, port int) *Agent {
	if agentInstance != nil {
		panic("agent already init")
	}
	addr, err := util.GetAddrByIpPort(ip, port)
	if err != nil {
		panic(err)
	}
	agentInstance = &Agent{
		id:          id,
		addr:        addr,
		joinedConns: conn.NewJoinedConnList(),
		version:     "0.0.1",
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
