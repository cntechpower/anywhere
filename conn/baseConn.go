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
	RemoteName      string
	connRwMu        sync.RWMutex
	Conn            net.Conn
	statusRwMutex   sync.RWMutex
	CreateTime      time.Time
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
}

func (c *WrappedConn) SetAck(sendTime, rcvTime time.Time) {
	c.statusRwMutex.Lock()
	defer c.statusRwMutex.Unlock()
	c.LastAckSendTime = sendTime
	c.LastAckRcvTime = rcvTime
}

func (c *WrappedConn) Send(m interface{}) error {
	c.connRwMu.RLock()
	defer c.connRwMu.RUnlock()
	if c.Conn == nil {
		return ErrNilConn
	}
	p, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := c.Conn.Write(p); err != nil {
		return err
	}
	return nil
}

func (c *WrappedConn) Receive(rsp interface{}) error {
	c.connRwMu.RLock()
	defer c.connRwMu.RUnlock()
	if c.Conn == nil {
		return ErrNilConn
	}
	d := json.NewDecoder(c.Conn)
	if err := d.Decode(&rsp); err != nil {
		return err
	}
	return nil
}

func (c *WrappedConn) Close() error {
	c.connRwMu.Lock()
	defer c.connRwMu.Unlock()
	if c.Conn == nil {
		return nil
	}
	err := c.Conn.Close()

	// set conn to nil because net.Conn do not have a isClose flag.
	// we used conn == nil to validate conn
	c.Conn = nil
	return err
}

func (c *WrappedConn) GetRemoteAddr() string {
	if c.Conn == nil {
		return ""
	}
	return c.Conn.RemoteAddr().String()
}

func (c *WrappedConn) GetLocalAddr() string {
	if c.Conn == nil {
		return ""
	}
	return c.Conn.LocalAddr().String()
}

func (c *WrappedConn) IsValid() bool {
	return c.Conn != nil
}

func (c *WrappedConn) ResetConn(conn net.Conn) {
	//close old connection if exist, let old goroutine stop.
	c.Close()

	c.Conn = conn
}

func (c *WrappedConn) GetConn() net.Conn {
	return c.Conn
}

func NewWrappedConn(remoteName string, c net.Conn) *WrappedConn {
	return &WrappedConn{
		RemoteName:      remoteName,
		Conn:            c,
		statusRwMutex:   sync.RWMutex{},
		CreateTime:      time.Now(),
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
	}
}
