package main

import (
	"anywhere/agent/anywhereAgent"
	"anywhere/conn"
	"fmt"
)

func main() {
	a := anywhereAgent.InitAnyWhereAgent("agent-id", "127.0.0.1", "1111")
	_ = a.SetCredentials("../credential/client.crt", "../credential/client.key", "../credential/ca.crt")
	a.Start()
	if err := conn.SendRequest(a.AdminConn, "V1", "ROUTE", "route config"); err != nil {
		fmt.Printf("send request error: %v", err)
		return
	}
	rsp, err := conn.ReadResponse(a.AdminConn)
	if err != nil {
		fmt.Printf("read response error: %v", err)
	} else {
		fmt.Println(rsp)
	}
	if err := a.AdminConn.Close(); err != nil {
		fmt.Printf("close conn error: %v", err)
	}
}
