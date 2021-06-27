package inst

import "github.com/cntechpower/anywhere/server/server"

var s *server.Server

func GetServerInst() *server.Server {
	if s == nil {
		panic("ServerInst is not init")
	}
	return s
}

func Init(si *server.Server) {
	s = si
}
