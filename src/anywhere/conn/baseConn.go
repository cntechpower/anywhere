package conn

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

type baseConn struct {
	conn            net.Conn
	status          CStatus
	statusMutex     sync.RWMutex
	LastAckSendTime time.Time
	LastAckRcvTime  time.Time
	failReason      string
	failCount       int
}

func (c *baseConn) SetHealthy() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusHealthy
	c.LastAckSendTime = time.Now()
	c.failCount = 0
}

func (c *baseConn) SetBad(reason string) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusBad
	c.failCount++
	c.failReason = reason

}

func (c *baseConn) GetFailCount() int {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.failCount
}

func (c *baseConn) GetStatus() CStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
}

func (c *baseConn) GetFailReason() string {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.failReason
}

func (c *baseConn) Send(m interface{}) error {
	p, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(p); err != nil {
		return err
	}
	return nil
}

func (c *baseConn) Receive(rsp interface{}) error {
	d := json.NewDecoder(c.conn)
	if err := d.Decode(&rsp); err != nil {
		return err
	}
	return nil
}

func (c *baseConn) GetRemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *baseConn) Close() {
	_ = c.conn.Close()
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusClosed
}

func (c *baseConn) GetRawConn() net.Conn {
	return c.conn
}
