package main

import (
	"anywhere/conn"
	"anywhere/tls"
	"fmt"
)

func main() {
	tlsConfig, err := tls.ParseTlsConfig("../credential/client.crt", "../credential/client.key", "../credential/ca.crt")
	if err != nil {
		panic(err)
	}
	c, err := tls.DialTlsServer("127.0.0.1", 1111, tlsConfig)
	if err != nil {
		panic(err)
	}
	if err := conn.SendRequest(c, "V1", "ROUTE", "route config"); err != nil {
		fmt.Printf("send request error: %v", err)
		return
	}
	rsp, err := conn.ReadResponse(c)
	if err != nil {
		fmt.Printf("read response error: %v", err)
	} else {
		fmt.Println(rsp)
	}
	if err := c.Close(); err != nil {
		fmt.Printf("close conn error: %v", err)
	}
}
