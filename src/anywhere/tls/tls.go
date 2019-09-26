package tls

import (
	"crypto/tls"
	_tls "crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

func getDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:  5 * time.Second,
		Deadline: time.Time{},
	}
}
func ParseTlsConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	tlsCert, err := _tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(ca)
	if !ok {
		return nil, fmt.Errorf("error while add ca to certPool")
	}
	return &_tls.Config{
		Certificates:       []_tls.Certificate{tlsCert},
		ClientAuth:         _tls.RequireAndVerifyClientCert,
		ClientCAs:          certPool,
		InsecureSkipVerify: true,
	}, nil
}

func DialTlsServer(ip string, port int, config *_tls.Config) (c *_tls.Conn, err error) {
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("wrong format of ip :%v", ip)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("wrong format of port: %v", port)
	}
	addr := fmt.Sprintf("%v:%v", ip, port)
	c, err = _tls.DialWithDialer(getDialer(), "tcp", addr, config)
	if err != nil {
		return nil, err
	}
	fmt.Printf("connected to %v\n", c.RemoteAddr())
	return c, nil
}
