package conn

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/cntechpower/anywhere/internal/trace"
	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
)

var (
	ErrConnectionPoolFull   = errors.New("connection pool is full")
	ErrConnectionPoolClosed = errors.New("connection pool is closed")
)

type ConnectionPool interface {
	Get(ctx context.Context, proxyAddr string) (*WrappedConn, error)
	Put(ctx context.Context, proxyAddr string, connection *WrappedConn) error
	Close()
}

type connectionPool struct {
	mu sync.Mutex
	// pool maps a proxy local address to a buffered channel of ready connections.
	pool map[string]chan *WrappedConn

	idleTimeout           time.Duration
	waitTimeout           time.Duration
	newConnectionFn       func(proxyAddr string)
	houseKeepLoopInterval time.Duration

	closed    bool
	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewConnectionPool(newConnectionFn func(proxyAddr string)) ConnectionPool {
	p := &connectionPool{
		pool:                  make(map[string]chan *WrappedConn),
		idleTimeout:           constants.ProxyConnMaxIdleTimeout * time.Second,
		waitTimeout:           constants.ProxyConnGetRetryMilliseconds * time.Millisecond,
		newConnectionFn:       newConnectionFn,
		houseKeepLoopInterval: 30 * time.Second,
		closeCh:               make(chan struct{}),
	}
	go p.houseKeeper()
	return p
}

// channel returns the buffered channel for proxyAddr, creating it if needed.
// It returns false once the pool is closed so callers cannot resurrect it.
func (p *connectionPool) channel(proxyAddr string) (chan *WrappedConn, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, false
	}
	ch, ok := p.pool[proxyAddr]
	if !ok {
		ch = make(chan *WrappedConn, constants.ProxyConnBufferForEachAgent)
		p.pool[proxyAddr] = ch
	}
	return ch, true
}

func (p *connectionPool) isClosed() bool {
	select {
	case <-p.closeCh:
		return true
	default:
		return false
	}
}

func (p *connectionPool) Get(ctx context.Context, proxyAddr string) (c *WrappedConn, err error) {
	span, ctxNew := trace.New(ctx, "connectionPool.Get")
	defer span.Finish()

	if p.isClosed() {
		return nil, ErrConnectionPoolClosed
	}
	ch, ok := p.channel(proxyAddr)
	if !ok {
		return nil, ErrConnectionPoolClosed
	}

	for i := 0; i < constants.ProxyConnGetMaxRetryCount; i++ {
		// Ask the agent for a fresh tunnel connection, then wait for it to be
		// delivered by Put. Ask first so a prefetched connection still wins the
		// race without wasting a request.
		_ = trace.Do(ctxNew, fmt.Sprintf("connectionPool.Get.Wait-%d", i), func() error {
			p.newConnectionFn(proxyAddr)
			timer := time.NewTimer(p.waitTimeout)
			defer timer.Stop()
			select {
			case c = <-ch:
				return nil
			case <-timer.C:
			case <-ctxNew.Done():
				err = ctxNew.Err()
			case <-p.closeCh:
				err = ErrConnectionPoolClosed
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if c != nil {
			return c, nil
		}
	}
	ext.HTTPStatusCode.Set(span, http.StatusRequestTimeout)
	ext.Error.Set(span, true)
	return nil, fmt.Errorf("timeout while waiting for proxy conn for %s", proxyAddr)
}

func (p *connectionPool) Put(_ context.Context, proxyAddr string, connection *WrappedConn) error {
	if connection == nil {
		return ErrNilConn
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		_ = connection.Close()
		return ErrConnectionPoolClosed
	}
	ch, ok := p.pool[proxyAddr]
	if !ok {
		ch = make(chan *WrappedConn, constants.ProxyConnBufferForEachAgent)
		p.pool[proxyAddr] = ch
	}
	select {
	case ch <- connection:
		p.mu.Unlock()
		return nil
	default:
		// Never block the caller (the conn reader) on a full pool.
		p.mu.Unlock()
		_ = connection.Close()
		return ErrConnectionPoolFull
	}
}

func (p *connectionPool) houseKeeper() {
	h := log.NewHeader("connection_pool_housekeeper")
	ticker := time.NewTicker(p.houseKeepLoopInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.closeCh:
			return
		case <-ticker.C:
			p.evictIdle(h)
		}
	}
}

// evictIdle scans every per-address channel once and closes connections that
// exceeded the idle timeout. All connections are inspected, not just the head.
func (p *connectionPool) evictIdle(h *log.Header) {
	p.mu.Lock()
	channels := make(map[string]chan *WrappedConn, len(p.pool))
	for addr, ch := range p.pool {
		channels[addr] = ch
	}
	p.mu.Unlock()

	now := time.Now()
	for addr, ch := range channels {
		n := len(ch)
	drain:
		for i := 0; i < n; i++ {
			var c *WrappedConn
			select {
			case c = <-ch:
			default:
				break drain
			}
			if c.CreateTime.Add(p.idleTimeout).Before(now) {
				log.Infof(h, "connection for %s exceeded idle timeout, closing it", addr)
				_ = c.Close()
				continue
			}
			select {
			case ch <- c:
			default:
				_ = c.Close()
			}
		}
	}
}

func (p *connectionPool) Close() {
	p.closeOnce.Do(func() {
		close(p.closeCh)
		p.mu.Lock()
		p.closed = true
		defer p.mu.Unlock()
		for addr, ch := range p.pool {
			for {
				select {
				case c := <-ch:
					if c != nil {
						_ = c.Close()
					}
					continue
				default:
				}
				break
			}
			delete(p.pool, addr)
		}
	})
}
