package conn

import (
	"net"
	"sync"
	"time"
)

type AdminConn struct {
	baseConn
}

func NewAdminConn(c net.Conn) *AdminConn {
	return &AdminConn{baseConn{
		conn:            c,
		status:          CStatusInit,
		statusMutex:     sync.RWMutex{},
		lastAckSendTime: time.Time{},
		lastAckRcvTime:  time.Time{},
		failReason:      "",
		failCount:       0,
	}}
}
