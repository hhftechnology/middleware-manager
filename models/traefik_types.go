package models

import (
	"database/sql/driver"
	"encoding/json"
)

// TCPTLSConfig represents TLS configuration for TCP routers
type TCPTLSConfig struct {
	Passthrough  bool     `json:"passthrough,omitempty"`
	CertResolver string   `json:"certResolver,omitempty"`
	Domains      []string `json:"domains,omitempty"`
	Options      string   `json:"options,omitempty"`
}

// TCPRouter represents a TCP router configuration from Traefik API
type TCPRouter struct {
	Name        string        `json:"name"`
	Rule        string        `json:"rule"`
	Service     string        `json:"service"`
	EntryPoints []string      `json:"entryPoints"`
	Middlewares []string      `json:"middlewares,omitempty"`
	TLS         *TCPTLSConfig `json:"tls,omitempty"`
	Priority    int           `json:"priority"`
	Provider    string        `json:"provider"`
	Status      string        `json:"status"`
}

// UDPRouter represents a UDP router configuration from Traefik API
type UDPRouter struct {
	Name        string   `json:"name"`
	Service     string   `json:"service"`
	EntryPoints []string `json:"entryPoints"`
	Provider    string   `json:"provider"`
	Status      string   `json:"status"`
}

// TCPService represents a TCP service configuration from Traefik API
type TCPService struct {
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	LoadBalancer *struct {
		Servers []struct {
			Address string `json:"address"`
			Weight  *int   `json:"weight,omitempty"`
		} `json:"servers,omitempty"`
		TerminationDelay *int `json:"terminationDelay,omitempty"`
	} `json:"loadBalancer,omitempty"`
	Weighted *struct {
		Services []struct {
			Name   string `json:"name"`
			Weight int    `json:"weight"`
		} `json:"services,omitempty"`
	} `json:"weighted,omitempty"`
}

// UDPService represents a UDP service configuration from Traefik API
type UDPService struct {
	Name         string `json:"name"`
	Provider     string `json:"provider"`
	LoadBalancer *struct {
		Servers []struct {
			Address string `json:"address"`
		} `json:"servers,omitempty"`
	} `json:"loadBalancer,omitempty"`
	Weighted *struct {
		Services []struct {
			Name   string `json:"name"`
			Weight int    `json:"weight"`
		} `json:"services,omitempty"`
	} `json:"weighted,omitempty"`
}

// TCPMiddleware represents a TCP middleware configuration from Traefik API
type TCPMiddleware struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type,omitempty"`
	Provider string                 `json:"provider,omitempty"`
	Status   string                 `json:"status,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
	// TCP middleware specific types
	InFlightConn *struct {
		Amount int `json:"amount"`
	} `json:"inFlightConn,omitempty"`
	IPAllowList *struct {
		SourceRange []string `json:"sourceRange"`
	} `json:"ipAllowList,omitempty"`
	IPWhiteList *struct {
		SourceRange []string `json:"sourceRange"`
	} `json:"ipWhiteList,omitempty"`
}

// HTTPRouter is a type alias for TraefikRouter with SQL driver support
type HTTPRouter TraefikRouter

// Scan implements sql.Scanner for HTTPRouter (JSON deserialization from DB)
func (r *HTTPRouter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), r)
	}
	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer for HTTPRouter (JSON serialization to DB)
func (r HTTPRouter) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// ToDynamic converts HTTPRouter to TraefikRouter
func (r *HTTPRouter) ToDynamic() *TraefikRouter {
	if r == nil {
		return nil
	}
	router := TraefikRouter(*r)
	return &router
}

// WrapRouter wraps a TraefikRouter into HTTPRouter
func WrapRouter(r *TraefikRouter) *HTTPRouter {
	if r == nil {
		return nil
	}
	router := HTTPRouter(*r)
	return &router
}

// Scan implements sql.Scanner for TCPRouter
func (r *TCPRouter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), r)
	}
	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer for TCPRouter
func (r TCPRouter) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements sql.Scanner for UDPRouter
func (r *UDPRouter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), r)
	}
	return json.Unmarshal(bytes, r)
}

// Value implements driver.Valuer for UDPRouter
func (r UDPRouter) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// HTTPMiddleware is a type alias for TraefikMiddleware with SQL driver support
type HTTPMiddleware TraefikMiddleware

// Scan implements sql.Scanner for HTTPMiddleware
func (m *HTTPMiddleware) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), m)
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for HTTPMiddleware
func (m HTTPMiddleware) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// ToDynamic converts HTTPMiddleware to TraefikMiddleware
func (m *HTTPMiddleware) ToDynamic() *TraefikMiddleware {
	if m == nil {
		return nil
	}
	middleware := TraefikMiddleware(*m)
	return &middleware
}

// WrapMiddleware wraps a TraefikMiddleware into HTTPMiddleware
func WrapMiddleware(m *TraefikMiddleware) *HTTPMiddleware {
	if m == nil {
		return nil
	}
	middleware := HTTPMiddleware(*m)
	return &middleware
}

// Scan implements sql.Scanner for TCPMiddleware
func (m *TCPMiddleware) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), m)
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for TCPMiddleware
func (m TCPMiddleware) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// HTTPService is a type alias for TraefikService with SQL driver support
type HTTPService TraefikService

// Scan implements sql.Scanner for HTTPService
func (s *HTTPService) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), s)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for HTTPService
func (s HTTPService) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// ToDynamic converts HTTPService to TraefikService
func (s *HTTPService) ToDynamic() *TraefikService {
	if s == nil {
		return nil
	}
	service := TraefikService(*s)
	return &service
}

// WrapService wraps a TraefikService into HTTPService
func WrapService(s *TraefikService) *HTTPService {
	if s == nil {
		return nil
	}
	service := HTTPService(*s)
	return &service
}

// Scan implements sql.Scanner for TCPService
func (s *TCPService) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), s)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for TCPService
func (s TCPService) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for UDPService
func (s *UDPService) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), s)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for UDPService
func (s UDPService) Value() (driver.Value, error) {
	return json.Marshal(s)
}
