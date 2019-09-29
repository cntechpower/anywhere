package conn

import (
	"net"
)

type CStatus string

func (s CStatus) String() string {
	return string(s)
}

const (
	CStatusHealthy CStatus = "Healthy"
	CStatusBad     CStatus = "Bad"
	CStatusInit    CStatus = "Init"
	CStatusClosed  CStatus = "Closed"
)

type Conn interface {
	SetHealthy()
	SetBad(string)
	GetFailCount() int
	GetStatus() CStatus
	GetFailReason() string
	GetRemoteAddr() string
	Send(interface{}) error
	Receive(interface{}) error
	Close()
	GetRawConn() net.Conn
}
