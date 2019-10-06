package conn

import (
	"anywhere/log"
	"io"
	"net"
	"os"
	"sync"
)

func JoinConn(dst, src net.Conn) {
	var wg sync.WaitGroup
	join := func(dst, src net.Conn) {
		defer dst.Close()
		defer src.Close()
		defer wg.Done()
		dConn := io.MultiWriter(dst, os.Stdout)
		if _, err := io.Copy(dConn, src); err != nil {
			log.Error("io copy got error %v", err)
		}
	}
	wg.Add(2)
	go join(dst, src)
	go join(src, dst)
	log.Info("joined conn %v and %v", dst.RemoteAddr(), src.RemoteAddr())
	wg.Wait()
}
