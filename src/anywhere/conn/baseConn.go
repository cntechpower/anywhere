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
	lastAckSendTime time.Time
	lastAckRcvTime  time.Time
	failReason      string
	failCount       int
}

func (c *baseConn) setHealthy() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = CStatusHealthy
	c.failCount = 0
}

func (c *baseConn) setBad() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.failCount++
	if c.failCount >= 3 {
		c.status = CStatusBad
	}
}

func (c *baseConn) GetStatus() CStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
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

func (c *baseConn) Close() {
	_ = c.conn.Close()
}

func (c *baseConn) HeartBeatLoop(f func(c net.Conn) error) {
	go func() {
		for {
			if err := f(c.conn); err != nil {
				c.setBad()
			} else {
				c.setHealthy()
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
