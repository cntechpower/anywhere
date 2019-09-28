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

	if err := a.SendProxyConfig("22", "172.16.100.1", "22"); err != nil {
		panic(err)
	}
	var rsp model.ResponseMsg
	if err := a.AdminConn.Receive(&rsp); err != nil {
		log.Error("receive message error: %v", err)
	} else {
		log.Info("got msg: %v", rsp)
	}

	a.Stop()
}
