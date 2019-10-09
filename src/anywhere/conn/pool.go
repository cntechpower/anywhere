package conn

import (
	"anywhere/log"
	"anywhere/model"
	"fmt"
	"net"
	"sync"
	"time"
)

type baseConnItem struct {
	conn       *BaseConn
	accessTime time.Time
}

func NewBaseConnItem(c net.Conn) *baseConnItem {
	return &baseConnItem{
		conn:       NewBaseConn(c),
		accessTime: time.Now(),
	}
}

type connPool struct {
	connMu        sync.Mutex
	agents        chan string
	conns         map[string]chan *baseConnItem
	connsPoolSize int
}

func (p *connPool) putConnToPool(id string, c net.Conn) error {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	if p.conns[id] == nil {
		p.conns[id] = make(chan *baseConnItem, p.connsPoolSize)
		p.agents <- id
	}
	select {
	case p.conns[id] <- NewBaseConnItem(c):
		return nil
	default:
		_ = c.Close()
		return fmt.Errorf("conn pool for %v is full", id)
	}
}

func (p *connPool) getConnFromPool(id string) (*BaseConn, error) {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	if p.conns[id] == nil {
		return nil, fmt.Errorf("conn pool for %v is not exist", id)
	}
	select {
	case c := <-p.conns[id]:
		return c.conn, nil
	default:
		return nil, fmt.Errorf("conn pool for %v is empty", id)
	}
}

func (p *connPool) healthCheckLoop(checkFunc func(conn *BaseConn) error) {
	for agent := range p.agents {
		log.Info("start health check loop for agent %v", agent)
		go func(connChan chan *baseConnItem) {
			for {
				for c := range connChan {
					if err := checkFunc(c.conn); err != nil {
						log.Error("check conn %v error, drop it", c.conn.RemoteAddr())
						continue
					}
					connChan <- c
				}
			}
		}(p.conns[agent])

	}
}

var pool *connPool

func InitConnPool(size int) *connPool {
	pool = &connPool{
		connMu:        sync.Mutex{},
		agents:        make(chan string, size),
		conns:         make(map[string]chan *baseConnItem, 0),
		connsPoolSize: size,
	}
	return pool
}

func getConnPool() *connPool {
	if pool != nil {
		return pool
	}
	panic("conn pool not init")
}

func PutToPool(id string, c net.Conn) error {
	return getConnPool().putConnToPool(id, c)
}

func GetFromPool(id string) (*BaseConn, error) {
	return getConnPool().getConnFromPool(id)
}

func HealthyCheck(checkFunc func(conn *BaseConn) error) {
	getConnPool().healthCheckLoop(checkFunc)
}

func HeartBeatCheckFunc(c *BaseConn) error {
	msg := &model.RequestMsg{}
	if err := c.Receive(&msg); err != nil {
		log.Error("receive from data conn %v  error: %v, close this data conn", c.RemoteAddr(), err)
		_ = c.Close()
		return err

	}
	switch msg.ReqType {
	case model.PkgReqHeartBeat:
		m, _ := model.ParseHeartBeatPkg(msg.Message)
		c.LastAckSendTime = m.SendTime
		c.LastAckRcvTime = time.Now()
		c.SetHealthy()
	default:
		log.Error("got unknown ReqType: %v from %v", msg.ReqType, c.RemoteAddr())
		_ = c.Close()
		return fmt.Errorf("unknown ReqType: %v", msg.ReqType)
	}
	return nil
}
