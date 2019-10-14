package main

import (
	"anywhere/agent/anywhereAgent"
	"anywhere/log"
	"fmt"
	"time"
)

func main() {
	log.InitStdLogger()
	a := anywhereAgent.InitAnyWhereAgent("agent-id", "127.0.0.1", "1111")
	_ = a.SetCredentials("../credential/client.crt", "../credential/client.key", "../credential/ca.crt")
	a.Start()
	fmt.Println(a.SendProxyConfig("3333", "127.0.0.1", "3306"))
	_ = a.SendProxyConfig("3334", "10.0.0.2", "80")
	time.Sleep(1000000 * time.Second)
}
