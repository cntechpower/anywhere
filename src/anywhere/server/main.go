package main

import (
	"anywhere/server/anywhereServer"
	"anywhere/util"
	"fmt"
)

func main() {
	//tlsConfig, err := _tls.ParseTlsConfig("../credential/server.crt", "../credential/server.key", "../credential/ca.crt")
	//if err != nil {
	//	panic(err)
	//}
	//if err := conn.ListenAndServeTls(1111, tlsConfig); err != nil {
	//	panic(err)
	//}
	s := anywhereServer.InitServerInstance("server-id", "1111", true, true)
	if err := s.SetCredentials("../credential/server.crt", "../credential/server.key", "../credential/ca.crt"); err != nil {
		panic(err)
	}
	s.Start()
	serverExitChan := util.ListenKillSignal()

	select {
	case <-serverExitChan:
		fmt.Println("Exiting...")
		s.ListAgentInfo()
	case <-s.ExitChan:
		fmt.Println("server exit")

	}

}
