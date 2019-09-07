package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

func main() {
	tlsCert, err := tls.LoadX509KeyPair("../credential/server.crt", "../credential/server.key")
	if err != nil {
		panic(err)
	}
	ca, err := ioutil.ReadFile("../credential/ca.crt")
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(ca)
	if !ok {
		panic("error while add ca to certPool")
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	ln, err := tls.Listen("tcp", ":1111", tlsConfig)
	defer func() {
		err := ln.Close()
		if err != nil {
			fmt.Printf("close error: %v", err)
		}
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("accept conn error: %v", err)
			continue
		}
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			fmt.Printf("set readtimeout error: %v", err)
		}
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("got message : %v\n", msg)

		n, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Println(n, err)
			return
		}
	}
}
