package handler

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/cntechpower/anywhere/dao"
	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/server/server"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("ANYWHERE_DB_DSN", "file::memory:?cache=shared")
	dao.Init(model.GetPersistModels())
	code := m.Run()
	dao.Close()
	os.Exit(code)
}

func newTestAPI(t *testing.T) *API {
	t.Helper()
	users := &model.UserConfig{Users: []*model.User{{UserName: "admin", UserPass: "admin"}}}
	s, err := server.New("test-server", 1111, users, 0)
	if err != nil {
		t.Fatal(err)
	}
	jwt, err := auth.NewJwtValidator([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return New(s, auth.NewUserValidator(users), jwt)
}

// TestNewRESTHandlerWiresOperations guards the dependency injected Swagger
// wiring: every generated operation must be registered before api.Serve runs.
func TestNewRESTHandlerWiresOperations(t *testing.T) {
	h := newTestAPI(t)
	restHandler, err := NewRESTHandler(h)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/v1/summary", "/api/v1/summary"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		restHandler.ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			var body map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("summary response is not json: %v", err)
			}
			return
		}
	}
	t.Fatalf("summary endpoint was not routed")
}

func TestMultiUserScoping(t *testing.T) {
	users := &model.UserConfig{Users: []*model.User{
		{UserName: "admin", UserPass: "admin", IsAdmin: true},
		{UserName: "alice", UserPass: "alice-pass", IsAdmin: false},
		{UserName: "bob", UserPass: "bob-pass", IsAdmin: false},
	}}
	s, err := server.New("test-server", 1111, users, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })

	c1, p1 := net.Pipe()
	c2, p2 := net.Pipe()
	defer p1.Close()
	defer p2.Close()
	s.RegisterAgent("alice", "z-alice", "agent-a", c1)
	s.RegisterAgent("bob", "z-bob", "agent-b", c2)

	lnA, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	portA := lnA.Addr().(*net.TCPAddr).Port
	_ = lnA.Close()
	lnB, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	portB := lnB.Addr().(*net.TCPAddr).Port
	_ = lnB.Close()

	if err := s.AddProxyConfig("alice", "z-alice", portA, "127.0.0.1:8080", false, "", model.ListenTypeTCP); err != nil {
		t.Fatal(err)
	}
	if err := s.AddProxyConfig("bob", "z-bob", portB, "127.0.0.1:8081", false, "", model.ListenTypeTCP); err != nil {
		t.Fatal(err)
	}

	jwt, err := auth.NewJwtValidator([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	h := New(s, auth.NewUserValidator(users), jwt)
	restHandler, err := NewRESTHandler(h)
	if err != nil {
		t.Fatal(err)
	}

	// Alice only sees her own proxy config.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/proxy/list", nil)
	req = req.WithContext(withAuthUser(req.Context(), "alice"))
	rec := httptest.NewRecorder()
	restHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("alice proxy list status = %d, want 200", rec.Code)
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0]["user_name"] != "alice" {
		t.Fatalf("alice saw unexpected configs: %+v", list)
	}

	// Alice cannot add a proxy config for Bob.
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/proxy/add",
		strings.NewReader("user_name=bob&zone_name=z-bob&remote_port=9090&local_addr=127.0.0.1:9090&white_list_enable=false"))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addReq = addReq.WithContext(withAuthUser(addReq.Context(), "alice"))
	addRec := httptest.NewRecorder()
	restHandler.ServeHTTP(addRec, addReq)
	if addRec.Code != http.StatusForbidden {
		t.Fatalf("cross-user proxy add status = %d, want 403", addRec.Code)
	}

	// Admin sees both Alice and Bob.
	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/proxy/list", nil)
	adminReq = adminReq.WithContext(withAuthUser(adminReq.Context(), "admin"))
	adminRec := httptest.NewRecorder()
	restHandler.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("admin proxy list status = %d, want 200", adminRec.Code)
	}
	var adminList []map[string]interface{}
	if err := json.Unmarshal(adminRec.Body.Bytes(), &adminList); err != nil {
		t.Fatal(err)
	}
	if len(adminList) != 2 {
		t.Fatalf("admin saw %d configs, want 2", len(adminList))
	}
}

func TestLoginRateLimitIgnoresSpoofedXFF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestAPI(t)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.POST("/user_login", h.UserLogin)

	loginFailures.Lock()
	loginFailures.m = make(map[string]*failureRecord)
	loginFailures.Unlock()

	for i := 0; i < loginMaxFailures; i++ {
		req := httptest.NewRequest(http.MethodPost, "/user_login", strings.NewReader("username=admin&password=wrong"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Forwarded-For", "10.0.0."+string(rune('1'+i)))
		req.RemoteAddr = "192.0.2.10:12345"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want 401", i, rec.Code)
		}
	}

	// Next attempt from the same RemoteAddr with a different XFF must be rate-limited.
	req := httptest.NewRequest(http.MethodPost, "/user_login", strings.NewReader("username=admin&password=wrong"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	req.RemoteAddr = "192.0.2.10:12345"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("rate-limited attempt status = %d, want 429", rec.Code)
	}
}
