package models

import (
	"testing"
)

func TestHTTPRouter_ScanValue(t *testing.T) {
	original := HTTPRouter{
		Name:    "test-router",
		Rule:    "Host(`example.com`)",
		Service: "test-svc",
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned HTTPRouter
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "test-router" {
		t.Errorf("Name = %q, want test-router", scanned.Name)
	}

	t.Run("nil", func(t *testing.T) {
		var r HTTPRouter
		if err := r.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}

func TestHTTPRouter_ToDynamic(t *testing.T) {
	t.Run("non-nil", func(t *testing.T) {
		r := &HTTPRouter{Name: "test"}
		dyn := r.ToDynamic()
		if dyn == nil {
			t.Fatal("ToDynamic() returned nil")
		}
		if dyn.Name != "test" {
			t.Errorf("Name = %q, want test", dyn.Name)
		}
	})

	t.Run("nil", func(t *testing.T) {
		var r *HTTPRouter
		dyn := r.ToDynamic()
		if dyn != nil {
			t.Error("ToDynamic() on nil should return nil")
		}
	})
}

func TestWrapRouter(t *testing.T) {
	t.Run("non-nil", func(t *testing.T) {
		r := &TraefikRouter{Name: "test"}
		wrapped := WrapRouter(r)
		if wrapped == nil {
			t.Fatal("WrapRouter() returned nil")
		}
		if wrapped.Name != "test" {
			t.Errorf("Name = %q, want test", wrapped.Name)
		}
	})

	t.Run("nil", func(t *testing.T) {
		wrapped := WrapRouter(nil)
		if wrapped != nil {
			t.Error("WrapRouter(nil) should return nil")
		}
	})
}

func TestHTTPMiddleware_ScanValue(t *testing.T) {
	original := HTTPMiddleware{
		Name: "test-mw",
		Type: "headers",
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned HTTPMiddleware
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "test-mw" {
		t.Errorf("Name = %q, want test-mw", scanned.Name)
	}
}

func TestHTTPMiddleware_ToDynamic(t *testing.T) {
	m := &HTTPMiddleware{Name: "test"}
	dyn := m.ToDynamic()
	if dyn == nil || dyn.Name != "test" {
		t.Errorf("ToDynamic() failed")
	}

	var nilM *HTTPMiddleware
	if nilM.ToDynamic() != nil {
		t.Error("nil ToDynamic should return nil")
	}
}

func TestWrapMiddleware(t *testing.T) {
	m := &TraefikMiddleware{Name: "test"}
	wrapped := WrapMiddleware(m)
	if wrapped == nil || wrapped.Name != "test" {
		t.Error("WrapMiddleware failed")
	}

	if WrapMiddleware(nil) != nil {
		t.Error("WrapMiddleware(nil) should return nil")
	}
}

func TestHTTPService_ScanValue(t *testing.T) {
	original := HTTPService{
		Name: "test-svc",
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned HTTPService
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "test-svc" {
		t.Errorf("Name = %q, want test-svc", scanned.Name)
	}

	t.Run("nil", func(t *testing.T) {
		var s HTTPService
		if err := s.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}

func TestHTTPService_ToDynamic(t *testing.T) {
	s := &HTTPService{Name: "test"}
	dyn := s.ToDynamic()
	if dyn == nil || dyn.Name != "test" {
		t.Error("ToDynamic failed")
	}

	var nilS *HTTPService
	if nilS.ToDynamic() != nil {
		t.Error("nil ToDynamic should return nil")
	}
}

func TestWrapService(t *testing.T) {
	s := &TraefikService{Name: "test"}
	wrapped := WrapService(s)
	if wrapped == nil || wrapped.Name != "test" {
		t.Error("WrapService failed")
	}

	if WrapService(nil) != nil {
		t.Error("WrapService(nil) should return nil")
	}
}

func TestTCPRouter_ScanValue(t *testing.T) {
	original := TCPRouter{Name: "tcp-router", Rule: "HostSNI(`*`)"}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned TCPRouter
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "tcp-router" {
		t.Errorf("Name = %q, want tcp-router", scanned.Name)
	}

	t.Run("nil", func(t *testing.T) {
		var r TCPRouter
		if err := r.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}

func TestUDPRouter_ScanValue(t *testing.T) {
	original := UDPRouter{Name: "udp-router"}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned UDPRouter
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "udp-router" {
		t.Errorf("Name = %q, want udp-router", scanned.Name)
	}

	t.Run("nil", func(t *testing.T) {
		var r UDPRouter
		if err := r.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}

func TestTCPService_ScanValue(t *testing.T) {
	original := TCPService{Name: "tcp-svc"}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned TCPService
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "tcp-svc" {
		t.Errorf("Name = %q, want tcp-svc", scanned.Name)
	}
}

func TestUDPService_ScanValue(t *testing.T) {
	original := UDPService{Name: "udp-svc"}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned UDPService
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "udp-svc" {
		t.Errorf("Name = %q, want udp-svc", scanned.Name)
	}
}

func TestTCPMiddleware_ScanValue(t *testing.T) {
	original := TCPMiddleware{Name: "tcp-mw", Type: "inFlightConn"}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned TCPMiddleware
	if err := scanned.Scan(data); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if scanned.Name != "tcp-mw" {
		t.Errorf("Name = %q, want tcp-mw", scanned.Name)
	}

	t.Run("nil", func(t *testing.T) {
		var m TCPMiddleware
		if err := m.Scan(nil); err != nil {
			t.Errorf("Scan(nil) error = %v", err)
		}
	})
}
