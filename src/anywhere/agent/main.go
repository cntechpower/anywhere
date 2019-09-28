package main

import (
	"anywhere/agent/anywhereAgent"
	"anywhere/log"
	"anywhere/model"
)

func main() {
	a := anywhereAgent.InitAnyWhereAgent("agent-id", "127.0.0.1", "1111")
	_ = a.SetCredentials("../credential/client.crt", "../credential/client.key", "../credential/ca.crt")
	a.Start()
	log.InitStdLogger()
	pConfig, err := model.NewProxyConfigMsg("22", "172.16.100.1", "22")
	if err != nil {
		panic(err)
	}
	req := model.NewRequestMsg("0.0.1",
		model.PkgReqNewproxy,
		&pConfig)
	if err := a.AdminConn.Send(req); err != nil {
		log.Error("send message error: %v", err)
	}

	//
	//if err := a.AdminConn.SendProxyConfig("22", "172.16.100.1", "22", "0.0.1"); err != nil {
	//	panic(err)
	//}

	p, err := model.ParseProxyConfig(req.Message)
	if err != nil {
		panic(err)
	}
	log.Info("send pConfig: %v,%v", pConfig.RemoteAddr, pConfig.LocalAddr)
	log.Info("send req %v,%v", p.RemoteAddr, p.LocalAddr)
	var rsp *model.RequestMsg
	if err := a.AdminConn.Receive(rsp); err != nil {
		log.Error("receive message error: %v", err)
	}
	log.Info("got msg: %v", rsp)
	a.Stop()
}
