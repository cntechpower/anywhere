package conf

import (
	"testing"

	"github.com/cntechpower/anywhere/model"
)

func TestNormalizeGeneratesJwtSecret(t *testing.T) {
	cfg := &model.SystemConfig{
		UiConfig: &model.UiConfig{},
		User:     &model.UserConfig{Users: []*model.User{{UserName: "admin", UserPass: "admin"}}},
	}
	normalize(cfg)
	if len(cfg.UiConfig.JwtSecret) < 32 {
		t.Fatalf("expected a long random secret, got %q", cfg.UiConfig.JwtSecret)
	}
	first := cfg.UiConfig.JwtSecret

	cfg2 := &model.SystemConfig{
		UiConfig: &model.UiConfig{},
		User:     &model.UserConfig{Users: []*model.User{{UserName: "admin", UserPass: "admin"}}},
	}
	normalize(cfg2)
	if cfg2.UiConfig.JwtSecret == first {
		t.Fatal("secrets must be random, two runs generated the same value")
	}
}

func TestNormalizeKeepsConfiguredSecret(t *testing.T) {
	cfg := &model.SystemConfig{
		UiConfig: &model.UiConfig{JwtSecret: "a-configured-secret-value"},
		User:     &model.UserConfig{Users: []*model.User{{UserName: "admin", UserPass: "admin"}}},
	}
	normalize(cfg)
	if cfg.UiConfig.JwtSecret != "a-configured-secret-value" {
		t.Fatal("configured secret must not be overwritten")
	}
}

func TestSessionKeyDerivesFromJwtSecret(t *testing.T) {
	Conf = &model.SystemConfig{UiConfig: &model.UiConfig{JwtSecret: "0123456789abcdef0123456789abcdef"}}
	k1 := SessionKey()
	if len(k1) != 32 {
		t.Fatalf("session key length = %d, want 32", len(k1))
	}
	Conf.UiConfig.JwtSecret = "another-secret-value-0123456789"
	k2 := SessionKey()
	if string(k1) == string(k2) {
		t.Fatal("session key must change when the jwt secret changes")
	}
}
