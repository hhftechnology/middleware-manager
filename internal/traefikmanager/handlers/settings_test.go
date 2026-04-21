package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSettingsHandlerTestConnectionOKViaVersion(t *testing.T) {
	env := newTestEnv(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/version" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()

	handler := NewSettingsHandler(env.settings, env.files, upstream.Client())
	router := gin.New()
	router.POST("/test", handler.TestConnection)

	payload, _ := json.Marshal(map[string]string{"url": upstream.URL})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if ok, _ := body["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got %v", body)
	}
}

func TestSettingsHandlerTestConnectionFallbackPing(t *testing.T) {
	env := newTestEnv(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			w.WriteHeader(http.StatusNotFound)
		case "/ping":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer upstream.Close()

	handler := NewSettingsHandler(env.settings, env.files, upstream.Client())
	router := gin.New()
	router.POST("/test", handler.TestConnection)

	payload, _ := json.Marshal(map[string]string{"url": upstream.URL})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if ok, _ := body["ok"].(bool); !ok {
		t.Fatalf("expected ok=true with ping fallback, got %v", body)
	}
}
