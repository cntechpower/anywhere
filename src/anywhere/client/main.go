package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"
)

func main() {
	clientCert, err := tls.LoadX509KeyPair("../credential/client.pem", "../credential/client.key")
	if err != nil {
		panic(err)
	}
	clientCertBytes, err := ioutil.ReadFile("../credential/client.pem")
	if err != nil {
		panic(err)
	}
	clientCertPool := x509.NewCertPool()
	if ok := clientCertPool.AppendCertsFromPEM(clientCertBytes); !ok {
		panic(err)
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: true,
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:1111", conf)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	n, err := conn.Write([]byte("hello\n"))
	if err != nil {
		fmt.Printf("n is %v, err is %v/n", n, err)
	}
	fmt.Println("after conn")
	buf := make([]byte, 100)
	if err := conn.SetDeadline(<-time.After(5 * time.Second)); err != nil {
		fmt.Printf("set readtimeout error: %v", err)
	}
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println(n, err)
		return
	}
	println(string(buf[:n]))
}
