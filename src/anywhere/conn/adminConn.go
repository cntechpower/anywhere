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
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
		failReason:      "",
		failCount:       0,
	}}
}
