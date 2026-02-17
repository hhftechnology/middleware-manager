package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) APIError {
	t.Helper()
	var resp APIError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return resp
}

func TestHandleAPIError_WithError(t *testing.T) {
	c, w := newTestContext()
	HandleAPIError(c, http.StatusBadRequest, "bad input", errors.New("field missing"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	resp := parseResponse(t, w)
	if resp.Code != http.StatusBadRequest {
		t.Errorf("resp.Code = %d, want %d", resp.Code, http.StatusBadRequest)
	}
	if resp.Message != "bad input" {
		t.Errorf("resp.Message = %q, want %q", resp.Message, "bad input")
	}
	if resp.Details != "field missing" {
		t.Errorf("resp.Details = %q, want %q", resp.Details, "field missing")
	}
}

func TestHandleAPIError_WithoutError(t *testing.T) {
	c, w := newTestContext()
	HandleAPIError(c, http.StatusNotFound, "not found", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	resp := parseResponse(t, w)
	if resp.Details != "" {
		t.Errorf("resp.Details = %q, want empty", resp.Details)
	}
}

func TestNotFound(t *testing.T) {
	t.Run("with ID", func(t *testing.T) {
		c, w := newTestContext()
		NotFound(c, "Middleware", "abc123")

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
		resp := parseResponse(t, w)
		if resp.Message != "Middleware not found: abc123" {
			t.Errorf("message = %q", resp.Message)
		}
	})

	t.Run("without ID", func(t *testing.T) {
		c, w := newTestContext()
		NotFound(c, "Service", "")

		resp := parseResponse(t, w)
		if resp.Message != "Service not found" {
			t.Errorf("message = %q", resp.Message)
		}
	})
}

func TestBadRequest(t *testing.T) {
	c, w := newTestContext()
	BadRequest(c, "invalid input", errors.New("parse error"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestServerError(t *testing.T) {
	c, w := newTestContext()
	ServerError(c, "internal failure", errors.New("db down"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUnauthorized(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		c, w := newTestContext()
		Unauthorized(c, "")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
		resp := parseResponse(t, w)
		if resp.Message != "Unauthorized access" {
			t.Errorf("message = %q, want %q", resp.Message, "Unauthorized access")
		}
	})

	t.Run("custom message", func(t *testing.T) {
		c, w := newTestContext()
		Unauthorized(c, "token expired")

		resp := parseResponse(t, w)
		if resp.Message != "token expired" {
			t.Errorf("message = %q, want %q", resp.Message, "token expired")
		}
	})
}

func TestForbidden(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		c, w := newTestContext()
		Forbidden(c, "")

		if w.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
		resp := parseResponse(t, w)
		if resp.Message != "Access forbidden" {
			t.Errorf("message = %q, want %q", resp.Message, "Access forbidden")
		}
	})

	t.Run("custom message", func(t *testing.T) {
		c, w := newTestContext()
		Forbidden(c, "admin only")

		resp := parseResponse(t, w)
		if resp.Message != "admin only" {
			t.Errorf("message = %q, want %q", resp.Message, "admin only")
		}
	})
}

func TestConflict(t *testing.T) {
	c, w := newTestContext()
	Conflict(c, "duplicate name", errors.New("unique constraint"))

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestUnprocessableEntity(t *testing.T) {
	c, w := newTestContext()
	UnprocessableEntity(c, "validation failed", errors.New("invalid config"))

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}
}

func TestServiceUnavailable(t *testing.T) {
	c, w := newTestContext()
	ServiceUnavailable(c, "traefik unreachable", errors.New("connection refused"))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}
