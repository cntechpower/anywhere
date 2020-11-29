package conn

import (
	"net"
)

type CStatus string

func (s CStatus) String() string {
	return string(s)
}

type Conn interface {
	SetHealthy()
	SetBad(string)
	GetFailCount() int
	GetStatus() CStatus
	GetFailReason() string
	GetRemoteAddr() string
	Send(interface{}) error
	Receive(interface{}) error
	Close() error
	GetRawConn() net.Conn
}
