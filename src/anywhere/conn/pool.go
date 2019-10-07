package conn

import (
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
		go func(connChan chan *baseConnItem) {
			for c := range connChan {
				if err := checkFunc(c.conn); err != nil {
					continue
				}
				connChan <- c
			}
		}(p.conns[agent])

	}
}

var pool *connPool

func InitConnPool(size int) *connPool {
	pool = &connPool{
		connMu:        sync.Mutex{},
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
