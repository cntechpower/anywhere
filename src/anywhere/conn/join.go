package conn

import (
	"anywhere/log"
	"io"
	"net"
	"sync"
)

func JoinConn(dst, src net.Conn) {
	var wg sync.WaitGroup
	joinWithClose := func(dst, src net.Conn) {
		defer wg.Done()
		defer src.Close()
		defer dst.Close()

		//dConn := io.MultiWriter(dst, os.Stdout)
		if _, err := io.Copy(dst, src); err != nil {
			log.Error("io copy got error %v", err)
		}
	}
	wg.Add(2)
	go joinWithClose(dst, src)
	go joinWithClose(src, dst)
	log.Info("joined conn %v and %v", dst.LocalAddr(), src.RemoteAddr())
	wg.Wait()
}
