package util

import (
	"testing"
)

func TestWhiteListDisabledAllowsEverything(t *testing.T) {
	l, err := NewWhiteList(1, "agent", "127.0.0.1:80", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if !l.IpInWhiteList(nil, "203.0.113.9") {
		t.Fatal("disabled whitelist must allow every IP")
	}
}

func TestWhiteListAllowsOnlyConfiguredCidrs(t *testing.T) {
	l, err := NewWhiteList(1, "agent", "127.0.0.1:80", "10.0.0.0/8,192.168.1.0/24", true)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		ip   string
		want bool
	}{
		{"10.1.2.3", true},
		{"192.168.1.10", true},
		{"192.168.2.10", false},
		{"203.0.113.9", false},
		{"127.0.0.1", false}, // no implicit loopback
		{"not-an-ip", false},
	}
	for _, c := range cases {
		if got := l.IpInWhiteList(nil, c.ip); got != c.want {
			t.Errorf("ip %s: got %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestWhiteListUpdateAndToggle(t *testing.T) {
	l, err := NewWhiteList(1, "agent", "127.0.0.1:80", "10.0.0.0/8", true)
	if err != nil {
		t.Fatal(err)
	}
	if l.IpInWhiteList(nil, "192.168.1.1") {
		t.Fatal("unexpected allow before update")
	}
	if err := l.UpdateCidrs("192.168.0.0/16"); err != nil {
		t.Fatal(err)
	}
	if !l.IpInWhiteList(nil, "192.168.1.1") {
		t.Fatal("expected updated cidr to allow")
	}
	l.SetEnable(false)
	if !l.IpInWhiteList(nil, "203.0.113.9") {
		t.Fatal("expected disabled whitelist to allow")
	}
}

func TestWhiteListInvalidCidr(t *testing.T) {
	if _, err := NewWhiteList(1, "agent", "127.0.0.1:80", "not-a-cidr", true); err == nil {
		t.Fatal("expected error for invalid cidr")
	}
}

func TestAddrInWhiteList(t *testing.T) {
	l, err := NewWhiteList(1, "agent", "127.0.0.1:80", "127.0.0.0/8", true)
	if err != nil {
		t.Fatal(err)
	}
	if !l.AddrInWhiteList(nil, "127.0.0.1:5555") {
		t.Fatal("expected host:port form to be accepted")
	}
	if l.AddrInWhiteList(nil, "[::1]:5555") {
		t.Fatal("ipv6 ::1 must not match 127.0.0.0/8")
	}
}
