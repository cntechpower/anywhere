package zone

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cntechpower/anywhere/conn"
	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
)

func TestMain(m *testing.M) {
	// Isolate the test database from the real proxy.db.
	_ = os.Setenv("ANYWHERE_DB_DSN", "file::memory:?cache=shared")
	dao.Init(model.GetPersistModels())
	code := m.Run()
	dao.Close()
	os.Exit(code)
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func freeUDPPort(t *testing.T) int {
	t.Helper()
	l, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.LocalAddr().(*net.UDPAddr).Port
}

func newTestZone(t *testing.T) IZone {
	t.Helper()
	z := NewZone("user", "zone", 0)
	t.Cleanup(z.Shutdown)
	return z
}

func TestAddProxyConfigBindsSynchronously(t *testing.T) {
	z := newTestZone(t)
	port := freeTCPPort(t)
	cfg := &model.ProxyConfig{
		UserName:   "user",
		ZoneName:   "zone",
		RemotePort: port,
		LocalAddr:  "127.0.0.1:9",
		ListenType: model.ListenTypeTCP,
	}
	if err := z.AddProxyConfig(cfg); err != nil {
		t.Fatalf("add proxy config: %v", err)
	}
	if got := z.GetProxyConfigCount(); got != 1 {
		t.Fatalf("config count = %d, want 1", got)
	}
	if err := z.AddProxyConfig(cfg); err == nil {
		t.Fatal("expected duplicate config to be rejected")
	}
	if err := z.RemoveProxyConfig(port, "127.0.0.1:9"); err != nil {
		t.Fatalf("remove proxy config: %v", err)
	}
	if got := z.GetProxyConfigCount(); got != 0 {
		t.Fatalf("config count = %d, want 0", got)
	}
}

func TestAddProxyConfigReportsPortConflict(t *testing.T) {
	z := newTestZone(t)
	// Bind on 0.0.0.0 to match the zone listener; SO_REUSEADDR on some
	// platforms still allows 0.0.0.0 after a 127.0.0.1 bind.
	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	cfg := &model.ProxyConfig{
		UserName:   "user",
		ZoneName:   "zone",
		RemotePort: port,
		LocalAddr:  "127.0.0.1:9",
		ListenType: model.ListenTypeTCP,
	}
	if err := z.AddProxyConfig(cfg); err == nil {
		t.Fatal("expected port conflict error instead of a silent success")
	}
	if got := z.GetProxyConfigCount(); got != 0 {
		t.Fatalf("failed config must not be registered, count = %d", got)
	}
}

func TestConfigCountIsPerZone(t *testing.T) {
	z1 := NewZone("u1", "z1", 0)
	defer z1.Shutdown()
	z2 := NewZone("u2", "z2", 0)
	defer z2.Shutdown()

	cfg := &model.ProxyConfig{
		UserName:   "u1",
		ZoneName:   "z1",
		RemotePort: freeTCPPort(t),
		LocalAddr:  "127.0.0.1:9",
		ListenType: model.ListenTypeTCP,
	}
	if err := z1.AddProxyConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if got := z2.GetCurrentConnectionCount(); got != 0 {
		t.Fatalf("other zone connection count = %d, want 0", got)
	}
}

// TestDataConnTokenBinding verifies that a data connection can only be added
// with a valid, single use token issued to that agent.
func TestDataConnTokenBinding(t *testing.T) {
	z := newTestZone(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	server, peer := net.Pipe()
	defer peer.Close()
	z.RegisterAgent("agent-1", server)

	msgs := make(chan *model.RequestMsg, 8)
	go func() {
		wc := conn.NewWrappedConn("test", peer)
		for {
			var m model.RequestMsg
			if err := wc.Receive(&m); err != nil {
				close(msgs)
				return
			}
			msgs <- &m
		}
	}()

	const localAddr = "127.0.0.1:8080"
	z.(*Zone).requestNewProxyConn(localAddr)

	var tunnel *model.TunnelBeginMsg
	select {
	case m := <-msgs:
		var err error
		tunnel, err = model.ParseTunnelBeginPkg(m.Message)
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("agent did not receive tunnel begin")
	}
	if tunnel.Token == "" {
		t.Fatal("expected a token in the tunnel begin message")
	}

	// Wrong token is rejected.
	bad, badPeer := net.Pipe()
	defer badPeer.Close()
	if err := z.PutProxyConnection(ctx, "agent-1", localAddr, "not-a-token", bad); err == nil {
		t.Fatal("expected invalid token to be rejected")
	}

	// Correct token is accepted once.
	data, dataPeer := net.Pipe()
	defer dataPeer.Close()
	if err := z.PutProxyConnection(ctx, "agent-1", localAddr, tunnel.Token, data); err != nil {
		t.Fatalf("put proxy connection: %v", err)
	}
	// The token must be single use.
	replay, replayPeer := net.Pipe()
	defer replayPeer.Close()
	if err := z.PutProxyConnection(ctx, "agent-1", localAddr, tunnel.Token, replay); err == nil {
		t.Fatal("expected token replay to be rejected")
	}

	c, err := z.(*Zone).connectionPool.Get(ctx, localAddr)
	if err != nil {
		t.Fatalf("get pooled conn: %v", err)
	}
	if c == nil {
		t.Fatal("expected a pooled connection")
	}
}

func TestUDPEchoRoundTrip(t *testing.T) {
	z := newTestZone(t)
	port := freeUDPPort(t)
	cfg := &model.ProxyConfig{
		UserName:   "user",
		ZoneName:   "zone",
		RemotePort: port,
		LocalAddr:  "127.0.0.1:9999",
		ListenType: model.ListenTypeUDP,
	}
	if err := z.AddProxyConfig(cfg); err != nil {
		t.Fatalf("add udp config: %v", err)
	}

	server, peer := net.Pipe()
	defer peer.Close()
	z.RegisterAgent("agent-udp", server)

	go func() {
		wc := conn.NewWrappedConn("test", peer)
		for {
			var m model.RequestMsg
			if err := wc.Receive(&m); err != nil {
				return
			}
			if m.ReqType != model.PkgUDPData {
				continue
			}
			msg, err := model.ParseUDPDataMsg(m.Message)
			if err != nil {
				continue
			}
			_ = wc.Send(model.NewUDPDataRespMsg(msg.SessionID, msg.LocalAddr, msg.Data))
		}
	}()

	client, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(3 * time.Second))

	if _, err := client.Write([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 32)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("read udp echo: %v", err)
	}
	if string(buf[:n]) != "ping" {
		t.Fatalf("got %q, want ping", string(buf[:n]))
	}
}

func TestConcurrentConfigAccess(t *testing.T) {
	z := newTestZone(t)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = z.ListProxyConfigs()
				_ = z.GetProxyConfigCount()
				_ = z.Info()
				_ = z.Infos()
			}
		}()
	}
	wg.Wait()
}

func TestProxyConfigStatsRestoresCounters(t *testing.T) {
	cfg := &model.ProxyConfig{
		UserName:                        "user",
		ZoneName:                        "zone",
		RemotePort:                      8080,
		LocalAddr:                       "127.0.0.1:80",
		NetworkFlowRemoteToLocalInBytes: 100,
		NetworkFlowLocalToRemoteInBytes: 200,
		ProxyConnectCount:               5,
		ProxyConnectRejectCount:         2,
	}
	stats := newProxyConfigStats(cfg, nil)
	stats.AddNetworkFlow(10, 20)
	stats.AddConnectCount(1)
	stats.AddConnectRejectedCount(1)

	snap := stats.snapshot()
	if snap.NetworkFlowRemoteToLocalInBytes != 110 || snap.NetworkFlowLocalToRemoteInBytes != 220 {
		t.Fatalf("flow counters = (%d, %d), want (110, 220)", snap.NetworkFlowRemoteToLocalInBytes, snap.NetworkFlowLocalToRemoteInBytes)
	}
	if snap.ProxyConnectCount != 6 || snap.ProxyConnectRejectCount != 3 {
		t.Fatalf("connect/reject = (%d, %d), want (6, 3)", snap.ProxyConnectCount, snap.ProxyConnectRejectCount)
	}
}

func TestUpdateWhiteListAtomicOnInvalidCIDR(t *testing.T) {
	z := newTestZone(t)
	port := freeTCPPort(t)
	cfg := &model.ProxyConfig{
		UserName:      "user",
		ZoneName:      "zone",
		RemotePort:    port,
		LocalAddr:     "127.0.0.1:9",
		ListenType:    model.ListenTypeTCP,
		IsWhiteListOn: false,
	}
	if err := z.AddProxyConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if err := z.UpdateProxyConfigWhiteListConfig(port, "127.0.0.1:9", "not-a-cidr", true); err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
	// Verify ACL was not accidentally enabled on error.
	p := z.(*Zone).proxyConfigs[z.(*Zone).getProxyConfigMapKey(port, "127.0.0.1:9")]
	if !p.acl.IpInWhiteList(context.Background(), "1.2.3.4") {
		t.Fatal("ACL must remain disabled when CIDR update fails")
	}
}

func TestMultiUDPSessionIsolationAndFlow(t *testing.T) {
	z := newTestZone(t)
	port1 := freeUDPPort(t)
	port2 := freeUDPPort(t)
	const sharedLocal = "127.0.0.1:9999"

	for _, port := range []int{port1, port2} {
		if err := z.AddProxyConfig(&model.ProxyConfig{
			UserName:   "user",
			ZoneName:   "zone",
			RemotePort: port,
			LocalAddr:  sharedLocal,
			ListenType: model.ListenTypeUDP,
		}); err != nil {
			t.Fatal(err)
		}
	}

	server, peer := net.Pipe()
	defer peer.Close()
	z.RegisterAgent("agent-udp", server)

	go func() {
		wc := conn.NewWrappedConn("test", peer)
		for {
			var m model.RequestMsg
			if err := wc.Receive(&m); err != nil {
				return
			}
			if m.ReqType != model.PkgUDPData {
				continue
			}
			msg, err := model.ParseUDPDataMsg(m.Message)
			if err != nil {
				continue
			}
			_ = wc.Send(model.NewUDPDataRespMsg(msg.SessionID, msg.LocalAddr, msg.Data))
		}
	}()

	for i, port := range []int{port1, port2} {
		c, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
		if err != nil {
			t.Fatal(err)
		}
		_ = c.SetDeadline(time.Now().Add(3 * time.Second))
		payload := []byte{byte('a' + i), 'b', 'c'}
		if _, err := c.Write(payload); err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 32)
		n, err := c.Read(buf)
		_ = c.Close()
		if err != nil {
			t.Fatalf("udp port %d read error: %v", port, err)
		}
		if string(buf[:n]) != string(payload) {
			t.Fatalf("udp port %d got %q, want %q", port, string(buf[:n]), string(payload))
		}
	}

	for _, cfg := range z.ListProxyConfigs() {
		if cfg.NetworkFlowLocalToRemoteInBytes == 0 || cfg.NetworkFlowRemoteToLocalInBytes == 0 {
			t.Fatalf("expected non-zero UDP flow stats for port %d, got in=%d out=%d",
				cfg.RemotePort, cfg.NetworkFlowRemoteToLocalInBytes, cfg.NetworkFlowLocalToRemoteInBytes)
		}
	}
}
