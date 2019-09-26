package model

import "net"

const (
	NEWPROXY = "NEWPROXY"
)

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
