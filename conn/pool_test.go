package conn

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cntechpower/utils/log"

	"github.com/cntechpower/anywhere/constants"
)

func newTestPool(newConnFn func(string)) *connectionPool {
	if newConnFn == nil {
		newConnFn = func(string) {}
	}
	return &connectionPool{
		pool:                  make(map[string]chan *WrappedConn),
		idleTimeout:           time.Minute,
		waitTimeout:           20 * time.Millisecond,
		newConnectionFn:       newConnFn,
		houseKeepLoopInterval: time.Hour,
		closeCh:               make(chan struct{}),
	}
}

func newTestWrappedConn(t *testing.T) (*WrappedConn, net.Conn) {
	t.Helper()
	a, b := net.Pipe()
	return NewWrappedConn("test", a), b
}

func TestPoolPutGet(t *testing.T) {
	p := newTestPool(nil)
	defer p.Close()

	c, peer := newTestWrappedConn(t)
	defer peer.Close()
	if err := p.Put(context.Background(), "127.0.0.1:80", c); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := p.Get(context.Background(), "127.0.0.1:80")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != c {
		t.Fatal("get returned a different connection")
	}
}

func TestPoolGetRequestsNewConn(t *testing.T) {
	p := newTestPool(nil)
	defer p.Close()

	var once sync.Once
	p.newConnectionFn = func(addr string) {
		once.Do(func() {
			c, peer := newTestWrappedConn(t)
			t.Cleanup(func() { _ = peer.Close() })
			go func() {
				_ = p.Put(context.Background(), addr, c)
			}()
		})
	}

	got, err := p.Get(context.Background(), "127.0.0.1:81")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected a connection")
	}
}

func TestPoolPutFullDoesNotBlock(t *testing.T) {
	p := newTestPool(nil)
	defer p.Close()

	addr := "127.0.0.1:82"
	for i := 0; i < constants.ProxyConnBufferForEachAgent; i++ {
		c, peer := newTestWrappedConn(t)
		defer peer.Close()
		if err := p.Put(context.Background(), addr, c); err != nil {
			t.Fatalf("put #%d: %v", i, err)
		}
	}

	c, peer := newTestWrappedConn(t)
	defer peer.Close()
	done := make(chan error, 1)
	go func() { done <- p.Put(context.Background(), addr, c) }()
	select {
	case err := <-done:
		if err != ErrConnectionPoolFull {
			t.Fatalf("got %v, want ErrConnectionPoolFull", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Put blocked on a full pool")
	}
}

func TestPoolCloseUnblocksGet(t *testing.T) {
	p := newTestPool(nil)
	errCh := make(chan error, 1)
	go func() {
		_, err := p.Get(context.Background(), "127.0.0.1:83")
		errCh <- err
	}()
	time.Sleep(10 * time.Millisecond)
	p.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error from closed pool")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Get did not unblock after Close")
	}
}

func TestPoolEvictsAllIdleConnections(t *testing.T) {
	p := newTestPool(nil)
	defer p.Close()

	c, peer := newTestWrappedConn(t)
	defer peer.Close()
	c.CreateTime = time.Now().Add(-2 * time.Minute)
	if err := p.Put(context.Background(), "127.0.0.1:84", c); err != nil {
		t.Fatal(err)
	}
	p.evictIdle(log.NewHeader("test"))
	if c.IsValid() {
		t.Fatal("expected idle connection to be closed")
	}
}

func TestPoolConcurrentPutGet(t *testing.T) {
	p := newTestPool(nil)
	defer p.Close()

	addr := "127.0.0.1:85"
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, peer := newTestWrappedConn(t)
			defer peer.Close()
			if err := p.Put(context.Background(), addr, c); err != nil {
				return
			}
			_, _ = p.Get(context.Background(), addr)
		}()
	}
	wg.Wait()
}
