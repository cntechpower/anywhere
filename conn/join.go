package conn

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// activityConn updates a shared last-active timestamp on every successful read
// or write, so unidirectional traffic keeps the tunnel alive.
type activityConn struct {
	net.Conn
	lastActive *atomic.Int64
}

func (a *activityConn) Read(p []byte) (int, error) {
	n, err := a.Conn.Read(p)
	if n > 0 {
		a.lastActive.Store(time.Now().UnixNano())
	}
	return n, err
}

func (a *activityConn) Write(p []byte) (int, error) {
	n, err := a.Conn.Write(p)
	if n > 0 {
		a.lastActive.Store(time.Now().UnixNano())
	}
	return n, err
}

// JoinConn bridges two connections and returns the number of bytes copied
// local->remote and remote->local. Both connections are closed when either
// direction finishes or fails.
func JoinConn(remote, local net.Conn) (localToRemoteBytes, remoteToLocalBytes uint64) {
	return JoinConnWithIdleTimeout(remote, local, 0)
}

// JoinConnWithIdleTimeout is JoinConn with an optional idle timeout: if no data
// flows in either direction for longer than idle, both connections are closed.
// A zero or negative idle disables the timeout (long lived tunnels such as SSH).
func JoinConnWithIdleTimeout(remote, local net.Conn, idle time.Duration) (localToRemoteBytes, remoteToLocalBytes uint64) {
	if remote == nil || local == nil {
		return 0, 0
	}
	remoteC, localC := remote, local
	var done chan struct{}
	if idle > 0 {
		var lastActive atomic.Int64
		lastActive.Store(time.Now().UnixNano())
		remoteC = &activityConn{Conn: remote, lastActive: &lastActive}
		localC = &activityConn{Conn: local, lastActive: &lastActive}
		done = make(chan struct{})
		interval := idle / 4
		if interval < 5*time.Millisecond {
			interval = 5 * time.Millisecond
		} else if interval > time.Second {
			interval = time.Second
		}
		go func() {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if time.Since(time.Unix(0, lastActive.Load())) >= idle {
						_ = remote.Close()
						_ = local.Close()
						return
					}
				}
			}
		}()
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var localToRemote, remoteToLocal int64

	copyDir := func(dst, src net.Conn, counter *int64) {
		defer wg.Done()
		n, _ := io.Copy(dst, src)
		mu.Lock()
		*counter = n
		mu.Unlock()
		// Closing both sides unblocks the peer direction.
		_ = src.Close()
		_ = dst.Close()
	}

	wg.Add(2)
	go copyDir(remoteC, localC, &localToRemote)
	go copyDir(localC, remoteC, &remoteToLocal)
	wg.Wait()
	if done != nil {
		close(done)
	}

	mu.Lock()
	defer mu.Unlock()
	if localToRemote > 0 {
		localToRemoteBytes = uint64(localToRemote)
	}
	if remoteToLocal > 0 {
		remoteToLocalBytes = uint64(remoteToLocal)
	}
	return localToRemoteBytes, remoteToLocalBytes
}
