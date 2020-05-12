package util

import (
	"net"
	"strings"
	"sync"
)

type WhiteList struct {
	enable bool
	cidrs  []*net.IPNet
	//any r/w to cidrs should hold mutex by caller
	mutex sync.RWMutex
}

func getPrivateCidrs() []*net.IPNet {
	_, ipNetLocalHost, _ := net.ParseCIDR("127.0.0.1/32")
	_, ipNetA, _ := net.ParseCIDR("10.0.0.0/8")
	_, ipNetB, _ := net.ParseCIDR("127.16.0.0/12")
	_, ipNetC, _ := net.ParseCIDR("192.168.0.0/16")
	return []*net.IPNet{ipNetLocalHost, ipNetA, ipNetB, ipNetC}

}

func (l *WhiteList) AddrInWhiteList(addr string) bool {
	ip := strings.Split(addr, ":")[0]
	return l.IpInWhiteList(ip)
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

func (l *WhiteList) SetEnable(enable bool) {
	l.mutex.Lock()
	l.enable = enable
	l.mutex.Unlock()
}

func NewWhiteList(cidrList string, enable bool) (*WhiteList, error) {
	l := &WhiteList{
		enable: enable,
		cidrs:  make([]*net.IPNet, 0),
		mutex:  sync.RWMutex{},
	}

	//add private to start of cidrs
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

func (l *WhiteList) IpInWhiteList(ip string) bool {
	l.mutex.RLock()
	if !l.enable {
		return true
	}
	defer l.mutex.RUnlock()
	for _, cidr := range l.cidrs {
		if cidr.Contains(net.ParseIP(ip)) {
			return true
		}
	}
	return false
}
