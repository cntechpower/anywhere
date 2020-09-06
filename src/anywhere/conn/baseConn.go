package conn

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

var ErrNilConn = fmt.Errorf("empty net.Conn")

type WrappedConn struct {
	conn            net.Conn
	statusMutex     sync.RWMutex
	CreateTime      time.Time
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
}

func (c *WrappedConn) SetAck(sendTime, rcvTime time.Time) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.LastAckSendTime = sendTime
	c.LastAckRcvTime = rcvTime
}

func (c *WrappedConn) Send(m interface{}) error {
	if c.conn == nil {
		return ErrNilConn
	}
	p, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(p); err != nil {
		return err
	}
	return nil
}

func (c *WrappedConn) Receive(rsp interface{}) error {
	if c.conn == nil {
		return ErrNilConn
	}
	d := json.NewDecoder(c.conn)
	if err := d.Decode(&rsp); err != nil {
		return err
	}
	return nil
}

func (c *WrappedConn) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()

	// set conn to nil because net.Conn do not have a isClose flag.
	// we used conn == nil to validate conn
	c.conn = nil
	return err
}

func (c *WrappedConn) GetRemoteAddr() string {
	if c.conn == nil {
		return ""
	}
	return c.conn.RemoteAddr().String()
}

func (c *WrappedConn) GetLocalAddr() string {
	if c.conn == nil {
		return ""
	}
	return c.conn.LocalAddr().String()
}

func (c *WrappedConn) IsValid() bool {
	return c.conn != nil
}

func (c *WrappedConn) ResetConn(conn net.Conn) {
	//close old connection if exist, let old goroutine stop.
	c.Close()

	c.conn = conn
}

func (c *WrappedConn) GetConn() net.Conn {
	return c.conn
}

func NewWrappedConn(c net.Conn) *WrappedConn {
	return &WrappedConn{
		conn:            c,
		statusMutex:     sync.RWMutex{},
		CreateTime:      time.Now(),
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
	}
}
