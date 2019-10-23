package util

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func genErrInvalidIp(ip string) error {
	return fmt.Errorf("IP FORMAT INVALID: %v", ip)
}

func genErrInvalidPort(port string) error {
	return fmt.Errorf("PORT FORMAT INVALID: %v", port)
}

func ListenKillSignal() chan os.Signal {
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2 /*graceful-shutdown*/)
	return quitChan
}

func GetAddrByIpPort(ip string, port string) (*net.TCPAddr, error) {
	if i := net.ParseIP(ip); i == nil || i.String() != ip {
		return nil, genErrInvalidIp(ip)
	}
	if portInt, err := strconv.Atoi(port); err != nil || portInt > 65535 || portInt < 1 {
		return nil, genErrInvalidPort(port)
	}
	addrString := fmt.Sprintf("%v:%v", ip, port)
	return net.ResolveTCPAddr("tcp", addrString)
}

func ListenTcp(addr string) (*net.TCPListener, error) {
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp", rAddr)
	if err != nil {
		return nil, err
	}
	return ln, err
}
