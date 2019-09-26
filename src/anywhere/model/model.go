package model

import "net"

type Agent struct {
	Id           string
	ServerId     string
	Addr         net.Addr
	ProxyConfigs []ProxyConfig
}

type ProxyConfig struct {
	RemoteAddr net.Addr
	LocalAddr  net.Addr
}

type RequestMsg struct {
	Version string
	ReqType string
	Message string
}

type ResponseMsg struct {
	Code    int
	Message string
}
