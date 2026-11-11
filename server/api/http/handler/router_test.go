package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cntechpower/anywhere/model"
	"github.com/cntechpower/anywhere/server/api/auth"
	"github.com/cntechpower/anywhere/server/server"
)

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
