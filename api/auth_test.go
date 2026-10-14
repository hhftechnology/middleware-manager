package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hhftechnology/middleware-manager/internal/testutil"
)

func TestAuthConfigFromEnv(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		t.Setenv("MM_AUTH_ENABLED", "")
		t.Setenv("MM_AUTH_USER", "")
		t.Setenv("MM_AUTH_TOKEN", "")
		cfg, err := AuthConfigFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Enabled {
			t.Fatal("expected auth to be disabled")
		}
	})

	t.Run("enabled requires user and token", func(t *testing.T) {
		t.Setenv("MM_AUTH_ENABLED", "true")
		t.Setenv("MM_AUTH_USER", "admin")
		t.Setenv("MM_AUTH_TOKEN", "")
		if _, err := AuthConfigFromEnv(); err == nil {
			t.Fatal("expected error when token is missing")
		}
	})

	t.Run("enabled with credentials", func(t *testing.T) {
		t.Setenv("MM_AUTH_ENABLED", "1")
		t.Setenv("MM_AUTH_USER", "admin")
		t.Setenv("MM_AUTH_TOKEN", "test-token")
		cfg, err := AuthConfigFromEnv()
		if err != nil {
			t.Fatal(err)
		}
		if !cfg.Enabled || cfg.Username != "admin" || cfg.Token != "test-token" {
			t.Fatalf("unexpected config: %+v", cfg)
		}
	})
}

func TestOptionalBasicAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testutil.NewTempDB(t)
	cm := testutil.NewTestConfigManager(t)
	traefikPath := filepath.Join(t.TempDir(), "traefik.yml")

	srv := NewServer(db, ServerConfig{
		Port:   "0",
		UIPath: filepath.Join(t.TempDir(), "missing-ui"),
		Auth: AuthConfig{
			Enabled:  true,
			Username: "admin",
			Token:    "test-token",
		},
	}, cm, traefikPath)

	health := httptest.NewRecorder()
	srv.router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health without credentials: expected 200, got %d", health.Code)
	}

	unauth := httptest.NewRecorder()
	srv.router.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, "/api/datasource/active", nil))
	if unauth.Code != http.StatusUnauthorized {
		t.Fatalf("api without credentials: expected 401, got %d", unauth.Code)
	}

	authedReq := httptest.NewRequest(http.MethodGet, "/api/v1/traefik-config/status", nil)
	authedReq.SetBasicAuth("admin", "test-token")
	authed := httptest.NewRecorder()
	srv.router.ServeHTTP(authed, authedReq)
	if authed.Code == http.StatusUnauthorized {
		t.Fatalf("api with credentials was rejected: %d %s", authed.Code, authed.Body.String())
	}

	wrong := httptest.NewRequest(http.MethodGet, "/api/datasource/active", nil)
	wrong.SetBasicAuth("admin", "nope")
	wrongRec := httptest.NewRecorder()
	srv.router.ServeHTTP(wrongRec, wrong)
	if wrongRec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token: expected 401, got %d", wrongRec.Code)
	}
}
