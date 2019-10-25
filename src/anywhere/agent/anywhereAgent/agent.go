package anywhereAgent

import (
	"anywhere/conn"
	"anywhere/log"
	"anywhere/model"
	"anywhere/tls"
	"anywhere/util"
	_tls "crypto/tls"
	"net"
)

type Agent struct {
	Id           string
	ServerId     string
	Addr         *net.TCPAddr
	credential   *_tls.Config
	AdminConn    *conn.BaseConn
	DataConn     []*conn.BaseConn
	ProxyConfigs []model.ProxyConfig
	version      string
	status       string
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
		Id:           id,
		ServerId:     "",
		Addr:         addr,
		ProxyConfigs: nil,
		version:      "0.0.1",
		status:       "INIT",
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
		panic("agent already started")
	}
	a.initControlConn(1)

	heartBeatExitChan := make(chan error, 0)
	go a.ControlConnHeartBeatSendLoop(1, heartBeatExitChan)
	go a.handleAdminConnection()
}

func (a *Agent) Stop() {
	if a.AdminConn != nil {
		a.AdminConn.Close()
		log.Info("Agent Stopping...")
	}
	a.status = "STOPPED"
}
