package conn

import (
	"fmt"
	"sync"
	"time"

	"github.com/cntechpower/anywhere/constants"
	"github.com/cntechpower/anywhere/log"
)

var ErrConnectionPoolFull = fmt.Errorf("connection pool is full")

type ConnectionPool interface {
	Get(proxyAddr string) (*WrappedConn, error)
	Put(proxyAddr string, connection *WrappedConn) error
}

type connectionPool struct {
	mu                    sync.Mutex
	pool                  map[string] /*localAddr*/ chan *WrappedConn
	idleTimeout           time.Duration
	waitTimeout           time.Duration
	newConnectionFn       func(proxyAddr string)
	houseKeepLoopInterval time.Duration
}

func NewConnectionPool(newConnectionFn func(proxyAddr string)) ConnectionPool {
	p := &connectionPool{
		mu:                    sync.Mutex{},
		pool:                  make(map[string] /*localAddr*/ chan *WrappedConn, 0),
		idleTimeout:           constants.ProxyConnMaxIdleTimeout * time.Second,
		waitTimeout:           constants.ProxyConnGetRetryMilliseconds * time.Millisecond,
		newConnectionFn:       newConnectionFn,
		houseKeepLoopInterval: 30 * time.Second,
	}
	go p.houseKeeper()
	return p
}

func (p *connectionPool) Get(proxyAddr string) (*WrappedConn, error) {
	if _, ok := p.pool[proxyAddr]; !ok {
		p.pool[proxyAddr] = make(chan *WrappedConn, constants.ProxyConnBufferForEachAgent)
	}
	for i := 0; i < constants.ProxyConnGetMaxRetryCount; i++ {
		select {
		case c := <-p.pool[proxyAddr]:
			return c, nil
		case <-time.After(p.waitTimeout):
		}
		//get connection timeout, try to request a new connection.
		p.newConnectionFn(proxyAddr)
	}

	return nil, fmt.Errorf("timeout while waiting for proxy conn for %v", proxyAddr)
}

func (p *connectionPool) Put(proxyAddr string, connection *WrappedConn) error {
	if _, ok := p.pool[proxyAddr]; !ok {
		p.pool[proxyAddr] = make(chan *WrappedConn, constants.ProxyConnBufferForEachAgent)
	}
	if len(p.pool[proxyAddr]) >= constants.ProxyConnBufferForEachAgent {
		_ = connection.Close()
		return ErrConnectionPoolFull
	}
	p.pool[proxyAddr] <- connection
	return nil
}

func (p *connectionPool) houseKeeper() {
	h := log.NewHeader("connection_pool_housekeeper")
	for range time.NewTicker(p.houseKeepLoopInterval).C {
		p.mu.Lock()
		for proxyAddr, pool := range p.pool {
			if len(pool) == 0 {
				continue
			}
			checkedMap := make(map[*WrappedConn]struct{}, len(pool))
			select {
			case c := <-p.pool[proxyAddr]:
				//this connection is already checked
				//because channel is FIFO, so that means all connection in channel has been checked.
				if _, ok := checkedMap[c]; ok {
					break
				}
				checkedMap[c] = struct{}{}
				if c.CreateTime.Add(p.idleTimeout).Before(time.Now()) { //connection is exceed idle timeout, closing it.
					log.Infof(h, "connection for %v is exceed idle timeout, will close it.", proxyAddr)
					_ = c.Close()
				} else {
					p.pool[proxyAddr] <- c
				}
			default:
			}
			p.mu.Unlock()
		}
	}
}
