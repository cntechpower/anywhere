package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("ANYWHERE_DB_DSN", "file::memory:?cache=shared")
	dao.Init(model.GetPersistModels())
	code := m.Run()
	dao.Close()
	os.Exit(code)
}

func shutdownServer(t *testing.T, s *Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	s, err := New("test-server", 1111, &model.UserConfig{Users: []*model.User{
		{UserName: "admin", UserPass: "admin"},
	}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestConcurrentZoneAccess is the regression test for the "concurrent map read
// and map write" panic: RegisterAgent mutated the zones map without a lock
// while list/summary calls iterated it.
func TestConcurrentZoneAccess(t *testing.T) {
	s := newTestServer(t)
	defer shutdownServer(t, s)

	peers := make([]net.Conn, 0, 32)
	defer func() {
		for _, p := range peers {
			_ = p.Close()
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		serverConn, peer := net.Pipe()
		peers = append(peers, peer)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.RegisterAgent("admin", fmt.Sprintf("zone-%d", i%4), fmt.Sprintf("agent-%d", i), serverConn)
		}(i)
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				s.ListZones()
				s.ListAgentInfo()
				s.ListProxyConfigs()
				s.GetSummary()
			}
		}()
	}
	wg.Wait()

	if got := len(s.ListZones()); got != 4 {
		t.Fatalf("zone count = %d, want 4", got)
	}
	if got := len(s.ListAgentInfo()); got != 32 {
		t.Fatalf("agent count = %d, want 32", got)
	}
}

func TestGetSummaryReturnsCopy(t *testing.T) {
	s := newTestServer(t)
	defer shutdownServer(t, s)

	s.statusRwMutex.Lock()
	s.statusCache.ProxyNetworkFlowTop10 = []*model.ProxyConfig{{LocalAddr: "a"}}
	s.statusRwMutex.Unlock()

	got := s.GetSummary()
	got.ProxyNetworkFlowTop10[0] = &model.ProxyConfig{LocalAddr: "mutated"}

	again := s.GetSummary()
	if again.ProxyNetworkFlowTop10[0].LocalAddr != "a" {
		t.Fatal("GetSummary must return a copy of the top lists")
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	s := newTestServer(t)
	shutdownServer(t, s)
	shutdownServer(t, s)
}
