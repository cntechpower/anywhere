package conn

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

type BaseConn struct {
	net.Conn
	statusMutex     sync.RWMutex
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
}

func (c *BaseConn) SetAck(sendTime, rcvTime time.Time) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.LastAckSendTime = sendTime
	c.LastAckRcvTime = rcvTime
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

func (c *BaseConn) GetLocalAddr() string {
	return c.LocalAddr().String()
}

func NewBaseConn(c net.Conn) *BaseConn {
	return &BaseConn{
		Conn:            c,
		statusMutex:     sync.RWMutex{},
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
	}
}
