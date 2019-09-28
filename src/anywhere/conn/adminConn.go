package conn

import (
	"anywhere/model"
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

func (c *AdminConn) SendProxyConfig(remotePort, localIp, localPort, version string) error {
	p, err := model.NewProxyConfigMsg(remotePort, localIp, localPort)
	if err != nil {
		return err
	}
	msg := model.NewRequestMsg(version, model.PkgReqNewproxy, p)
	return c.Send(msg)
}
