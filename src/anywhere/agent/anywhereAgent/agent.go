package anywhereAgent

import (
	. "anywhere/model"
	"anywhere/tls"
	_tls "crypto/tls"
	"fmt"
	"net"
)

type Agent struct {
	Id           string
	ServerId     string
	Addr         *net.TCPAddr
	credential   *_tls.Config
	AdminConn    net.Conn
	ProxyConfigs []ProxyConfig
}

func InitAnyWhereAgent(id, ip, port string) *Agent {
	addrString := fmt.Sprintf("%v:%v", ip, port)
	addr, _ := net.ResolveTCPAddr("tcp", addrString)
	return &Agent{
		Id:           id,
		ServerId:     "",
		Addr:         addr,
		ProxyConfigs: nil,
	}
}

func (a *Agent) SetCredentials(certFile, keyFile, caFile string) error {
	tlsConfig, err := tls.ParseTlsConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}
	a.credential = tlsConfig
	return nil
}

func (a *Agent) Start() {
	c, err := tls.DialTlsServer(a.Addr.IP.String(), a.Addr.Port, a.credential)
	if err != nil {
		panic(err)
	}
	a.AdminConn = c
}

func (a *Agent) Send() {

}
