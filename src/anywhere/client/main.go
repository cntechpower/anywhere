package main

import (
	_conn "anywhere/conn"
	"anywhere/tls"
	"encoding/json"
	"fmt"
)

func main() {
	tlsConfig, err := tls.ParseTlsConfig("../credential/client.crt", "../credential/client.key", "../credential/ca.crt")
	if err != nil {
		panic(err)
	}
	conn, err := tls.DialTlsServer("127.0.0.1", 1111, tlsConfig)
	if err != nil {
		panic(err)
	}
	p, err := json.Marshal(&_conn.Package{
		Version: "3.19.09.0",
		Type:    "admin",
		Message: "HelloWorldFromClient",
	})
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write(p); err != nil {
		fmt.Printf("send message error: %v", err)
	}

	buf := make([]byte, 2000)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err, string(buf[:n]))
		return
	}
	fmt.Println(string(buf[:n]))
	if err := conn.Close(); err != nil {
		fmt.Printf("close conn error: %v", err)
	}
}
