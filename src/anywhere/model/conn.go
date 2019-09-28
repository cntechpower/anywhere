package model

import (
	"net"
	"sync"
	"time"
)

type ConnStatus string

const (
	ConnHealthy ConnStatus = "CStatusHealthy"
	ConnBad     ConnStatus = "CStatusBad"
	ConnInit    ConnStatus = "Init"
)

type Conn interface {
	SetHealthy()
	SetBad()
	GetStatus() ConnStatus
	HeartBeatLoop()
}

type baseConn struct {
	conn        net.Conn
	status      ConnStatus
	statusMutex sync.Mutex
	lastAck     time.Time
	failCount   int
}

func (c *baseConn) SetHealthy() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = ConnHealthy
}

func (c *baseConn) SetBad() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = ConnBad
}

//func (c *baseConn) HeartBeatLoop() {
//	conn.HeartBeatLoop(c.conn)
//}

type AdminConn struct {
	baseConn
}

func NewAdminConn(c net.Conn) *AdminConn {
	return &AdminConn{baseConn{
		conn:   c,
		status: ConnInit,
	}}
}

type DataConnList struct {
	connList []*baseConn
}

//func (c *DataConnList) HeartBeatLoop() {
//	for _, c1 := range c.connList {
//		conn.HeartBeatLoop(c1.conn)
//	}
//}
