package util

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cntechpower/utils/tracing"
)

const (
	connCountRefreshInterval = time.Minute
	connCountRejectCount     = 1000
)

type WhiteList struct {
	remotePort int
	agentId    string
	localAddr  string
	enable     bool
	cidrs      []*net.IPNet
	// any r/w to cidrs should hold mutex by caller
	mutex sync.RWMutex

	connCountMu sync.Mutex
	connCount   map[string]int64
}

func getPrivateCidrs() []*net.IPNet {
	_, ipNetLocalHost, _ := net.ParseCIDR("127.0.0.1/32")
	_, ipNetA, _ := net.ParseCIDR("10.0.0.0/8")
	_, ipNetB, _ := net.ParseCIDR("127.16.0.0/12")
	_, ipNetC, _ := net.ParseCIDR("192.168.0.0/16")
	return []*net.IPNet{ipNetLocalHost, ipNetA, ipNetB, ipNetC}

}

func (l *WhiteList) AddrInWhiteList(ctx context.Context, addr string) bool {
	ip := strings.Split(addr, ":")[0]
	return l.IpInWhiteList(ctx, ip)
}

func (l *WhiteList) AddCidrToList(cidrString string, reset bool) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if reset {
		l.cidrs = make([]*net.IPNet, 0)
	}
	_, cidr, err := net.ParseCIDR(cidrString)
	if err != nil {
		return err
	}
	l.cidrs = append(l.cidrs, cidr)
	return nil
}

func (l *WhiteList) UpdateCidrs(cidrsString string) (err error) {
	cidrs := strings.Split(cidrsString, ",")
	tmpCidrs := make([]*net.IPNet, 0, len(cidrs))
	for _, cidrString := range cidrs {
		if cidrString == "" {
			continue
		}
		_, cidr, err := net.ParseCIDR(cidrString)
		if err != nil {
			return err
		}
		tmpCidrs = append(tmpCidrs, cidr)
	}

	l.mutex.Lock()
	l.cidrs = tmpCidrs
	l.mutex.Unlock()

	return
}

func (l *WhiteList) SetEnable(enable bool) {
	l.mutex.Lock()
	l.enable = enable
	l.mutex.Unlock()
}

func NewWhiteList(remotePort int, agentId, localAddr, cidrList string, enable bool) (*WhiteList, error) {
	l := &WhiteList{
		remotePort:  remotePort,
		agentId:     agentId,
		localAddr:   localAddr,
		enable:      enable,
		cidrs:       make([]*net.IPNet, 0),
		mutex:       sync.RWMutex{},
		connCountMu: sync.Mutex{},
		connCount:   make(map[string]int64, 0),
	}
	go func() {
		for {
			time.Sleep(connCountRefreshInterval)
			l.connCountMu.Lock()
			l.connCount = make(map[string]int64, 0)
			l.connCountMu.Unlock()
		}
	}()

	// add private to start of cidrs
	l.cidrs = append(l.cidrs, getPrivateCidrs()...)
	if cidrList == "" {
		return l, nil
	}
	cidrs := strings.Split(cidrList, ",")
	for _, cidrString := range cidrs {
		_, cidr, err := net.ParseCIDR(cidrString)
		if err != nil {
			return nil, err
		}
		l.cidrs = append(l.cidrs, cidr)
	}
	return l, nil
}

func (l *WhiteList) IpInWhiteList(ctx context.Context, ip string) (res bool) {
	if ctx != nil {
		span, _ := tracing.New(ctx, "WhiteList.IpInWhiteList")
		defer span.Finish()
	}
	l.connCountMu.Lock()
	defer l.connCountMu.Unlock()
	if l.connCount[ip] > connCountRejectCount {
		return false
	}

	l.mutex.RLock()
	defer func() {
		l.mutex.RUnlock()
		l.connCount[ip]++
	}()
	if !l.enable {
		res = true
		return
	}
	for _, cidr := range l.cidrs {
		if cidr.Contains(net.ParseIP(ip)) {
			res = true
			return
		}
	}
	return res
}
