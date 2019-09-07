package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/siddontang/go-log/log"
)

func demoHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	_, err := w.Write([]byte("Hello from anywhere"))
	if err != nil {
		fmt.Printf("write response error: %v\n", err)
	}
}

func main() {
	http.HandleFunc("/", demoHandler)
	tlsCert, err := tls.LoadX509KeyPair("../credential/server.pem", "../credential/server.key")
	if err != nil {
		panic(err)
	}
	ca, err := ioutil.ReadFile("../credential/ca.pem")
	if err != nil {
		panic(err)
	}
	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(ca)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		//ClientAuth:   tls.RequireAndVerifyClientCert,
		//ClientCAs:    clientCertPool,
	}
	ln, err := tls.Listen("tcp", ":1111", tlsConfig)
	defer func() {
		err := ln.Close()
		if err != nil {
			log.Fatal("close error: %v", err)
		}
	}()
	for {
		conn, err := ln.Accept()
		if err := conn.SetDeadline(<-time.After(5 * time.Second)); err != nil {
			fmt.Printf("set readtimeout error: %v", err)
		}
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	buffer := make([]byte, 0)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("read from conn error: %v\n", err)
	} else {
		fmt.Printf("got message: %v", string(buffer[:n]))
	}
	_, err = conn.Write([]byte("Got it"))
	if err != nil {
		fmt.Printf("send message error : %v", err)
	}
}
