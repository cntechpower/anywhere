package conn

import (
	"io"
	"net"
	"sync"

	"github.com/cntechpower/anywhere/log"
)

func JoinConn(remote, local net.Conn) (uint64, uint64) {
	h := log.NewHeader("JoinConn")
	var wg sync.WaitGroup
	joinWithClose := func(dst, src net.Conn, bytesCopied *int64) {
		defer wg.Done()
		defer src.Close()
		defer dst.Close()

		var err error
		*bytesCopied, err = io.Copy(dst, src)
		if err != nil {
			return
		}
	}
	wg.Add(2)
	var localToRemoteBytes, remoteToLocalBytes int64
	go joinWithClose(remote, local, &localToRemoteBytes)
	go joinWithClose(local, remote, &remoteToLocalBytes)
	log.Infof(h, "joined conn %v and %v", remote.LocalAddr(), local.RemoteAddr())
	wg.Wait()
	return uint64(localToRemoteBytes), uint64(remoteToLocalBytes)
}
