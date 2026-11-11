package conn

import (
	"io"
	"net"
	"sync"
	"time"
)

// deadlineConn refreshes the per-operation deadline before every read/write,
// which turns a blocking io.Copy into an idle-timeout aware copy.
type deadlineConn struct {
	net.Conn
	idle time.Duration
}

func (d *deadlineConn) Read(p []byte) (int, error) {
	if d.idle > 0 {
		_ = d.Conn.SetReadDeadline(time.Now().Add(d.idle))
	}
	return d.Conn.Read(p)
}

func (d *deadlineConn) Write(p []byte) (int, error) {
	if d.idle > 0 {
		_ = d.Conn.SetWriteDeadline(time.Now().Add(d.idle))
	}
	return d.Conn.Write(p)
}

// JoinConn bridges two connections and returns the number of bytes copied
// local->remote and remote->local. Both connections are closed when either
// direction finishes or fails.
func JoinConn(remote, local net.Conn) (localToRemoteBytes, remoteToLocalBytes uint64) {
	return JoinConnWithIdleTimeout(remote, local, 0)
}

// JoinConnWithIdleTimeout is JoinConn with an optional idle timeout: if no data
// flows in a direction for longer than idle, the connection is closed. A zero
// or negative idle disables the timeout (long lived tunnels such as SSH).
func JoinConnWithIdleTimeout(remote, local net.Conn, idle time.Duration) (localToRemoteBytes, remoteToLocalBytes uint64) {
	if remote == nil || local == nil {
		return 0, 0
	}
	remoteC, localC := remote, local
	if idle > 0 {
		remoteC = &deadlineConn{Conn: remote, idle: idle}
		localC = &deadlineConn{Conn: local, idle: idle}
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
