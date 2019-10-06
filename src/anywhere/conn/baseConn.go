package conn

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

type BaseConn struct {
	conn            net.Conn
	connMutex       sync.Mutex
	status          CStatus
	statusMutex     sync.RWMutex
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
	failReason      string
	failCount       int
}

func (c *BaseConn) SetHealthy() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusHealthy
	c.LastAckSendTime = time.Now()
	c.failCount = 0
}

func (c *BaseConn) SetBad(reason string) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusBad
	c.failCount++
	c.failReason = reason

}

func (c *BaseConn) GetFailCount() int {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.failCount
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
	if _, err := c.conn.Write(p); err != nil {
		return err
	}
	return nil
}

func (c *BaseConn) Receive(rsp interface{}) error {
	d := json.NewDecoder(c.conn)
	if err := d.Decode(&rsp); err != nil {
		return err
	}
	return nil
}

func (c *BaseConn) GetRemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *BaseConn) Close() {
	_ = c.conn.Close()
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusClosed
}

func (c *BaseConn) GetRawConn() net.Conn {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	return c.conn
}

func NewBaseConn(c net.Conn) *BaseConn {
	return &BaseConn{
		conn:            c,
		status:          CStatusInit,
		statusMutex:     sync.RWMutex{},
		LastAckSendTime: time.Time{},
		LastAckRcvTime:  time.Time{},
		failReason:      "",
		failCount:       0,
	}
}
