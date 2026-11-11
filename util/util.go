package util

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func genErrInvalidIp(ip string) error {
	return fmt.Errorf("IP FORMAT INVALID: %+v", ip)
}

func genErrInvalidPort(port string) error {
	return fmt.Errorf("PORT FORMAT INVALID: %+v", port)
}

func CheckAddrValid(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	return CheckPortValid(tcpAddr.Port)
}

func CheckPortValid(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port %+v", port)
	}
	return nil
}

func GetAddrByIpPort(ip string, port int) (*net.TCPAddr, error) {
	if i := net.ParseIP(ip); i == nil || i.String() != ip {
		return nil, genErrInvalidIp(ip)
	}
	if port > 65535 || port < 1 {
		return nil, genErrInvalidPort(strconv.Itoa(port))
	}
	addrString := fmt.Sprintf("%+v:%+v", ip, port)
	return net.ResolveTCPAddr("tcp", addrString)
}

// IsLoopbackAddr reports whether addr is bound to a loopback IP.
func IsLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
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

func ListenUdp(addr string) (*net.UDPConn, error) {
	rAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenUDP("udp", rAddr)
	if err != nil {
		return nil, err
	}
	return ln, err
}

func CheckPathExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func StringNvl(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func BoolNvl(s *bool) bool {
	if s == nil {
		return false
	}
	return *s
}

func Int64Nvl(s *int64) int64 {
	if s == nil {
		return 0
	}
	return *s
}

func BoolToString(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}
