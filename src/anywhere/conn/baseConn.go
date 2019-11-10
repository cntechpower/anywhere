package conn

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

type BaseConn struct {
	net.Conn
	status          CStatus
	statusMutex     sync.RWMutex
	failReason      string
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
}

func (c *BaseConn) SetHealthy() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusHealthy
	c.LastAckSendTime = time.Now()
}

func (c *BaseConn) SetBad(reason string) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusBad
	c.failReason = reason

}

func (c *BaseConn) GetStatus() CStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
}

func (c *BaseConn) GetFailReason() string {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.failReason
}

func (c *BaseConn) Send(m interface{}) error {
	p, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := c.Write(p); err != nil {
		return err
	}
	return nil
}

func (c *BaseConn) Receive(rsp interface{}) error {
	d := json.NewDecoder(c)
	if err := d.Decode(&rsp); err != nil {
		return err
	}
	return nil
}

func (c *BaseConn) GetRemoteAddr() string {
	return c.RemoteAddr().String()
}

func NewBaseConn(c net.Conn) *BaseConn {
	return &BaseConn{
		Conn:            c,
		status:          CStatusInit,
		statusMutex:     sync.RWMutex{},
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
		failReason:      "",
	}
}
