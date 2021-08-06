package util

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func genErrInvalidIp(ip string) error {
	return fmt.Errorf("IP FORMAT INVALID: %v", ip)
}

func genErrInvalidPort(port string) error {
	return fmt.Errorf("PORT FORMAT INVALID: %v", port)
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
		return fmt.Errorf("invalid port %v", port)
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
	addrString := fmt.Sprintf("%v:%v", ip, port)
	return net.ResolveTCPAddr("tcp", addrString)
}

func GetIpPortByAddr(addr string) (string, int, error) {
	rAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", 0, err
	}
	return rAddr.IP.String(), rAddr.Port, nil
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

func SendUDP(addr string, data []byte) error {
	rAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()
	_, err = c.Write(data)
	return err
}

func FormatTimestampForFileName() string {
	return time.Now().Format("2006_01_02_15_04")
}

func CheckPathExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func MkdirIfNotExist(path string) error {
	if s, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0755)
		}
		return err
	} else if s.IsDir() {
		return nil
	} else {
		return fmt.Errorf("%s exists, but is not a directory", path)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandInt(n int) int {
	return rand.Intn(n)
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
