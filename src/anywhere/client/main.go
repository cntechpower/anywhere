package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

func main() {
	clientCert, err := tls.LoadX509KeyPair("../credential/client.crt", "../credential/client.key")
	if err != nil {
		panic(err)
	}
	caCertBytes, err := ioutil.ReadFile("../credential/ca.crt")
	if err != nil {
		panic(err)
	}
	clientCertPool := x509.NewCertPool()
	if ok := clientCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		panic(err)
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: true,
	}
	tlsDialer := &net.Dialer{
		Timeout:  5 * time.Second,
		Deadline: time.Time{},
	}
	conn, err := tls.DialWithDialer(tlsDialer, "tcp", "127.0.0.1:1111", conf)
	if err != nil {
		panic(err)
	}
	fmt.Printf("connected to %v\n", conn.RemoteAddr())
	n, err := conn.Write([]byte("hello\n"))
	if err != nil {
		fmt.Printf("n is %v, err is %v\n", n, err)
	}
	buf := make([]byte, 100)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println(n, err)
		return
	}
	fmt.Println(string(buf[:n]))
	if err := conn.Close(); err != nil {
		fmt.Printf("close conn error: %v", err)
	}
}
