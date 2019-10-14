package anywhereServer

import (
	"anywhere/log"
	"net"
)

func (s *anyWhereServer) listenPort(addr string) *net.TCPListener {
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Error("parse proxy port error: %v", err)
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		log.Error("listen proxy port error: %v", err)
	}
	return ln
}
