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
	AdminConn    *conn.AdminConn
	ProxyConfigs []model.ProxyConfig
	version      string
	status       string
}

var agentInstance *Agent

func InitAnyWhereAgent(id, ip, port string) *Agent {
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
	a.connectControlConn()
	go a.ControlConnHeartBeatLoop(1)
}

func (a *Agent) Stop() {
	if a.AdminConn != nil {
		a.AdminConn.Close()
		log.Info("Agent Stopping...")
	}
	a.status = "STOPPED"
}
