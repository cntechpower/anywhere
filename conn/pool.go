package conn

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/cntechpower/utils/tracing"

	"github.com/cntechpower/anywhere/constants"
	log "github.com/cntechpower/utils/log.v2"
)

var ErrConnectionPoolFull = fmt.Errorf("connection pool is full")

type ConnectionPool interface {
	Get(ctx context.Context, proxyAddr string) (*WrappedConn, error)
	Put(ctx context.Context, proxyAddr string, connection *WrappedConn) error
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
		pool:                  make(map[string] /*localAddr*/ chan *WrappedConn),
		idleTimeout:           constants.ProxyConnMaxIdleTimeout * time.Second,
		waitTimeout:           constants.ProxyConnGetRetryMilliseconds * time.Millisecond,
		newConnectionFn:       newConnectionFn,
		houseKeepLoopInterval: 30 * time.Second,
	}
	go p.houseKeeper()
	return p
}

func (p *connectionPool) Get(ctx context.Context, proxyAddr string) (c *WrappedConn, err error) {
	span, ctxNew := tracing.New(ctx, "connectionPool.Get")
	defer span.Finish()
	p.mu.Lock()
	if _, ok := p.pool[proxyAddr]; !ok {
		p.pool[proxyAddr] = make(chan *WrappedConn, constants.ProxyConnBufferForEachAgent)
	}
	p.mu.Unlock()
	for i := 0; i < constants.ProxyConnGetMaxRetryCount; i++ {
		_ = tracing.Do(ctxNew, fmt.Sprintf("connectionPool.Get.Wait-%v", i), func() error {
			//get connection first
			p.newConnectionFn(proxyAddr)
			select {
			case c = <-p.pool[proxyAddr]:
				return nil
			case <-time.After(p.waitTimeout):
			}
			return nil
		})
		if c != nil {
			return
		}
	}
	ext.HTTPStatusCode.Set(span, http.StatusRequestTimeout)
	ext.Error.Set(span, true)
	return nil, fmt.Errorf("timeout while waiting for proxy conn for %v", proxyAddr)
}

func (p *connectionPool) Put(ctx context.Context, proxyAddr string, connection *WrappedConn) error {
	span, _ := tracing.New(ctx, "connectionPool.Put")
	defer span.Finish()
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
	fields := map[string]interface{}{
		log.FieldNameBizName: "connectionPool.houseKeeper",
	}
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
				if c.CreateTime.Add(p.idleTimeout).Before(time.Now()) { //connection is exceeded idle timeout, closing it.
					log.Infof(fields, "connection for %v is exceed idle timeout, will close it.", proxyAddr)
					_ = c.Close()
				} else {
					p.pool[proxyAddr] <- c
				}
			default:
			}
		}
		p.mu.Unlock()
	}
}
