package util

import (
	"strings"
)

type whiteList struct {
	list map[string]bool
}

func isPrivateIp(s string) bool {
	if strings.HasPrefix(s, "172.16") ||
		strings.HasPrefix(s, "10.0") ||
		strings.HasPrefix(s, "192.168") ||
		strings.HasPrefix(s, "127.0.0.1") {
		return true
	}
	return false
}

func (l *whiteList) AddrInWhiteList(addr string) bool {
	ip := strings.Split(addr, ":")[0]
	return l.IpInWhiteList(ip)
}

func (l *whiteList) AddIpToList(ip string) {
	l.list[ip] = true
}

func NewWhiteList(ips string) *whiteList {
	l := &whiteList{list: make(map[string]bool, 0)}
	if ips == "" {
		return l
	}
	ipList := strings.Split(ips, ",")
	for _, ip := range ipList {
		l.list[ip] = true
	}
	return l
}

func (l *whiteList) IpInWhiteList(ip string) bool {
	if isPrivateIp(ip) {
		return true
	}
	val, ok := l.list[ip]
	if ok == val == true {
		return true
	}
	return false
}
