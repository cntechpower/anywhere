package conn

import (
	"net"
)

type CStatus string

func (s CStatus) String() string {
	return string(s)
}

const (
	CStatusHealthy CStatus = "CStatusHealthy"
	CStatusBad     CStatus = "CStatusBad"
	CStatusInit    CStatus = "Init"
)

type Conn interface {
	setHealthy()
	setBad()
	GetStatus() CStatus
	GetRemoteAddr() string
	HeartBeatLoop(f func(c net.Conn) error)
	Send(interface{}) error
	Receive(interface{}) error
	Close()
}
