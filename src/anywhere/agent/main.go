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
	req := model.NewProxyConfigMsg("22", "172.16.100.1", "22")
	if err := a.AdminConn.Send(req); err != nil {
		log.Error("send message error: %v", err)
	}
	var rsp *model.RequestMsg
	err := a.AdminConn.Receive(rsp)
	if err != nil {
		log.Error("receive message error: %v", err)
	}
	log.Info("got msg: %v", rsp)
	a.Stop()
}
