package conn

import (
	"io"
	"net"
	"sync"

	log "github.com/cntechpower/utils/log.v2"
)

func JoinConn(remote, local net.Conn) (uint64, uint64) {
	fields := map[string]interface{}{
		log.FieldNameBizName: "conn.JoinConn",
	}
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
	log.Infof(fields, "joined conn %v and %v", remote.LocalAddr(), local.RemoteAddr())
	wg.Wait()
	return uint64(localToRemoteBytes), uint64(remoteToLocalBytes)
}
