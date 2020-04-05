package conn

import (
	"anywhere/log"
	"io"
	"net"
	"sync"
)

func JoinConn(remote, local net.Conn) {
	var wg sync.WaitGroup
	joinWithClose := func(dst, src net.Conn) {
		defer wg.Done()
		defer src.Close()
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return
		}
	}
	wg.Add(2)
	go joinWithClose(remote, local)
	go joinWithClose(local, remote)
	log.GetDefaultLogger().Infof("joined conn %v and %v", remote.LocalAddr(), local.RemoteAddr())
	wg.Wait()
}
