package util

import (
	"anywhere/log"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func genErrInvalidIp(ip string) error {
	return fmt.Errorf("IP FORMAT INVALID: %v", ip)
}

func genErrInvalidPort(port string) error {
	return fmt.Errorf("PORT FORMAT INVALID: %v", port)
}

var ipWhiteList map[string]bool

func isPrivateIp(s string) bool {
	if strings.HasPrefix(s, "172.16") ||
		strings.HasPrefix(s, "10.0") ||
		strings.HasPrefix(s, "192.168") ||
		strings.HasPrefix(s, "127.0.0.1") {
		return true
	}
	return false
}

func InitIpWhiteList(ips string) error {
	if ips == "" {
		return nil
	}
	ipList := strings.Split(ips, ",")
	for _, ip := range ipList {
		ipWhiteList[ip] = true
	}
	return nil
}

func IpInWhiteList(ip string) bool {
	if isPrivateIp(ip) {
		return true
	}
	val, ok := ipWhiteList[ip]
	if ok == true && val == true {
		return true
	}
	return false
}

func AddrInWhiteList(addr string) bool {
	ip := strings.Split(addr, ":")[0]
	return IpInWhiteList(ip)
}

func ListenKillSignal() chan os.Signal {
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2 /*graceful-shutdown*/)
	return quitChan
}

func ListenTTINSignalLoop() {
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.Signal(0x15))
	l := log.GetCustomLogger("ttin_listener")
	ttinChan := make(chan os.Signal, 10)
	signal.Notify(ttinChan, syscall.Signal(0x15))
	profileToCapture := []string{"cpu", "heap", "goroutine"}
	l.Infof("ttin listener started")
	for {
		sig := <-ttinChan
		switch sig {
		case syscall.Signal(0x15):
			dumpPath := "./dump_" + FormatTimestampForFileName()
			if err := MkdirIfNotExist(dumpPath); err != nil {
				l.Errorf("mkdir error: %v", err)
			}
			for _, name := range profileToCapture {
				err := CaptureProfile(name, dumpPath, 2)
				l.Infof("capture profile for %v, err: %v", name, err)
			}
		default:
			l.Fatalf("got unexpected signal: %v ", sig.String())
		}
	}
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

func FormatTimestampForFileName() string {
	return time.Now().Format("2006_01_02_15_04")
}

func MkdirIfNotExist(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(path, 0755)
		}
	}
	if s.IsDir() {
		return nil
	} else {
		return fmt.Errorf("%s exists, but is not a directory", path)
	}
}

func CaptureProfile(name, dumpPath string, extraInfo int) error {
	f, err := os.OpenFile(dumpPath+"/"+name+".out", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0640)
	if nil != err {
		return fmt.Errorf("write dump error(%v)", err)
	}
	defer f.Close()
	switch name {
	case "cpu":
		if extraInfo <= 0 {
			extraInfo = 30
		}
		if err := pprof.StartCPUProfile(f); nil != err {
			return err
		}
		time.Sleep(time.Duration(extraInfo) * time.Second)
		pprof.StopCPUProfile()
	case "heap":
		return pprof.Lookup("heap").WriteTo(f, 1)
	case "mutex":
		runtime.SetMutexProfileFraction(extraInfo)
		return pprof.Lookup("mutex").WriteTo(f, 1)
	case "block":
		runtime.SetBlockProfileRate(extraInfo)
		return pprof.Lookup("block").WriteTo(f, 1)
	case "goroutine":
		return pprof.Lookup("goroutine").WriteTo(f, 1)
	case "threadcreate":
		return pprof.Lookup("threadcreate").WriteTo(f, 1)
	default:
		return fmt.Errorf("not support profile %v", name)
	}
	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandId() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 3)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
