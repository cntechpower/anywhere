package util

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/cntechpower/anywhere/internal/trace"
)

// WhiteList decides whether a client IP may use a proxy config.
//
// Note: no CIDR is allowed implicitly. If the whitelist is enabled every
// allowed network (including private ranges) must be listed explicitly; the
// previous implementation silently allowed 10/8, 172.16/12, 192.168/16 and
// loopback, which defeated the purpose of enabling a whitelist.
type WhiteList struct {
	remotePort int
	agentId    string
	localAddr  string
	enable     bool
	cidrs      []*net.IPNet
	// any r/w to cidrs/enable should hold this mutex
	mutex sync.RWMutex
}

func (l *WhiteList) AddrInWhiteList(ctx context.Context, addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	return l.IpInWhiteList(ctx, host)
}

func (l *WhiteList) AddCidrToList(cidrString string, reset bool) error {
	_, cidr, err := net.ParseCIDR(cidrString)
	if err != nil {
		return err
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if reset {
		l.cidrs = make([]*net.IPNet, 0)
	}
	l.cidrs = append(l.cidrs, cidr)
	return nil
}

func (l *WhiteList) UpdateCidrs(cidrsString string) error {
	cidrs := strings.Split(cidrsString, ",")
	tmpCidrs := make([]*net.IPNet, 0, len(cidrs))
	for _, cidrString := range cidrs {
		cidrString = strings.TrimSpace(cidrString)
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
	return nil
}

func (l *WhiteList) SetEnable(enable bool) {
	l.mutex.Lock()
	l.enable = enable
	l.mutex.Unlock()
}

func NewWhiteList(remotePort int, agentId, localAddr, cidrList string, enable bool) (*WhiteList, error) {
	l := &WhiteList{
		remotePort: remotePort,
		agentId:    agentId,
		localAddr:  localAddr,
		enable:     enable,
		cidrs:      make([]*net.IPNet, 0),
	}
	if cidrList == "" {
		return l, nil
	}
	for _, cidrString := range strings.Split(cidrList, ",") {
		cidrString = strings.TrimSpace(cidrString)
		if cidrString == "" {
			continue
		}
		_, cidr, err := net.ParseCIDR(cidrString)
		if err != nil {
			return nil, err
		}
		l.cidrs = append(l.cidrs, cidr)
	}
	return l, nil
}

func (l *WhiteList) IpInWhiteList(ctx context.Context, ip string) bool {
	if ctx != nil {
		span, _ := trace.New(ctx, "WhiteList.IpInWhiteList")
		defer span.Finish()
	}

	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if !l.enable {
		return true
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, cidr := range l.cidrs {
		if cidr.Contains(parsed) {
			return true
		}
	}
	return false
}
