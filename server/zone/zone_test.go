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
