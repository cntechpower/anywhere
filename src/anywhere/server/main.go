package main

import (
	"anywhere/conn"
	_tls "anywhere/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func ListenKillSignal() chan os.Signal {
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2 /*graceful-shutdown*/)
	return quitChan
}

func main() {
	tlsConfig, err := _tls.ParseTlsConfig("../credential/server.crt", "../credential/server.key", "../credential/ca.crt")
	if err != nil {
		panic(err)
	}
	if err := conn.ListenAndServeTls(1111, tlsConfig); err != nil {
		panic(err)
	}
	serverExitChan := ListenKillSignal()

	select {
	case <-serverExitChan:
		fmt.Println("Exiting...")
	}

}
