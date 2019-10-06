package main

import (
	"anywhere/agent/anywhereAgent"
	"anywhere/log"
	"time"
)

func main() {
	log.InitStdLogger()
	a := anywhereAgent.InitAnyWhereAgent("agent-id", "127.0.0.1", "1111")
	_ = a.SetCredentials("../credential/client.crt", "../credential/client.key", "../credential/ca.crt")
	a.Start()
	_ = a.SendProxyConfig("3333", "127.0.0.1", "3306")
	time.Sleep(100 * time.Second)
}
