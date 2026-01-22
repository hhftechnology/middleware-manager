package models

import (
	"database/sql/driver"
	"encoding/json"
)

// TraefikOverview represents the /api/overview response from Traefik API
type TraefikOverview struct {
	HTTP      ProtocolOverview `json:"http"`
	TCP       ProtocolOverview `json:"tcp"`
	UDP       ProtocolOverview `json:"udp"`
	Features  TraefikFeatures  `json:"features"`
	Providers []string         `json:"providers"`
}

// ProtocolOverview represents statistics for a specific protocol
type ProtocolOverview struct {
	Routers     StatusCount `json:"routers"`
	Services    StatusCount `json:"services"`
	Middlewares StatusCount `json:"middlewares"`
}

// StatusCount represents counts with status breakdown
type StatusCount struct {
	Total    int `json:"total"`
	Warnings int `json:"warnings"`
	Errors   int `json:"errors"`
}

// TraefikFeatures represents enabled features in Traefik
type TraefikFeatures struct {
	Tracing   string `json:"tracing,omitempty"`
	Metrics   string `json:"metrics,omitempty"`
	AccessLog bool   `json:"accessLog"`
}

// TraefikEntrypoint represents an entrypoint from Traefik API
type TraefikEntrypoint struct {
	Name      string                 `json:"name"`
	Address   string                 `json:"address"`
	Transport TraefikTransportConfig `json:"transport,omitempty"`
	HTTP      *EntrypointHTTPConfig  `json:"http,omitempty"`
	HTTP2     *EntrypointHTTP2Config `json:"http2,omitempty"`
	HTTP3     *EntrypointHTTP3Config `json:"http3,omitempty"`
	UDP       *EntrypointUDPConfig   `json:"udp,omitempty"`
}

// TraefikTransportConfig represents transport configuration
type TraefikTransportConfig struct {
	LifeCycle          *LifeCycleConfig  `json:"lifeCycle,omitempty"`
	RespondingTimeouts *TimeoutsConfig   `json:"respondingTimeouts,omitempty"`
	ProxyProtocol      *ProxyProtocolConfig `json:"proxyProtocol,omitempty"`
}

// LifeCycleConfig represents lifecycle configuration
type LifeCycleConfig struct {
	RequestAcceptGraceTimeout string `json:"requestAcceptGraceTimeout,omitempty"`
	GraceTimeOut              string `json:"graceTimeOut,omitempty"`
}

// TimeoutsConfig represents timeout configuration
type TimeoutsConfig struct {
	ReadTimeout  string `json:"readTimeout,omitempty"`
	WriteTimeout string `json:"writeTimeout,omitempty"`
	IdleTimeout  string `json:"idleTimeout,omitempty"`
}

// ProxyProtocolConfig represents proxy protocol configuration
type ProxyProtocolConfig struct {
	Insecure   bool     `json:"insecure,omitempty"`
	TrustedIPs []string `json:"trustedIPs,omitempty"`
}

// EntrypointHTTPConfig represents HTTP-specific configuration for entrypoints
type EntrypointHTTPConfig struct {
	Redirections *EntrypointRedirections `json:"redirections,omitempty"`
	Middlewares  []string                `json:"middlewares,omitempty"`
	TLS          *EntrypointTLSConfig    `json:"tls,omitempty"`
}

// EntrypointRedirections represents HTTP redirections configuration
type EntrypointRedirections struct {
	EntryPoint *struct {
		To        string `json:"to,omitempty"`
		Scheme    string `json:"scheme,omitempty"`
		Permanent bool   `json:"permanent,omitempty"`
		Priority  int    `json:"priority,omitempty"`
	} `json:"entryPoint,omitempty"`
}

// EntrypointTLSConfig represents TLS configuration for entrypoints
type EntrypointTLSConfig struct {
	Options      string        `json:"options,omitempty"`
	CertResolver string        `json:"certResolver,omitempty"`
	Domains      []TLSDomain   `json:"domains,omitempty"`
}

// TLSDomain represents a TLS domain configuration
type TLSDomain struct {
	Main string   `json:"main"`
	Sans []string `json:"sans,omitempty"`
}

// EntrypointHTTP2Config represents HTTP/2 configuration
type EntrypointHTTP2Config struct {
	MaxConcurrentStreams int32 `json:"maxConcurrentStreams,omitempty"`
}

// EntrypointHTTP3Config represents HTTP/3 configuration
type EntrypointHTTP3Config struct {
	AdvertisedPort int `json:"advertisedPort,omitempty"`
}

// EntrypointUDPConfig represents UDP-specific configuration
type EntrypointUDPConfig struct {
	Timeout string `json:"timeout,omitempty"`
}

// TraefikVersion represents version info from Traefik API
type TraefikVersion struct {
	Version   string `json:"version"`
	Codename  string `json:"codename"`
	StartDate string `json:"startDate,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
}

// TraefikRawConfig represents the raw dynamic configuration from Traefik
// This matches the /api/rawdata endpoint response
type TraefikRawConfig struct {
	Routers     map[string]RawRouterInfo     `json:"routers,omitempty"`
	Middlewares map[string]RawMiddlewareInfo `json:"middlewares,omitempty"`
	Services    map[string]RawServiceInfo    `json:"services,omitempty"`
}

// RawRouterInfo represents router info from raw config
type RawRouterInfo struct {
	EntryPoints []string               `json:"entryPoints"`
	Middlewares []string               `json:"middlewares,omitempty"`
	Service     string                 `json:"service"`
	Rule        string                 `json:"rule"`
	Priority    int                    `json:"priority,omitempty"`
	TLS         map[string]interface{} `json:"tls,omitempty"`
	Status      string                 `json:"status"`
	Using       []string               `json:"using,omitempty"`
	Name        string                 `json:"name"`
	Provider    string                 `json:"provider"`
}

// RawMiddlewareInfo represents middleware info from raw config
type RawMiddlewareInfo struct {
	Status   string `json:"status"`
	Using    []string `json:"using,omitempty"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Type     string `json:"type,omitempty"`
}

// RawServiceInfo represents service info from raw config
type RawServiceInfo struct {
	LoadBalancer map[string]interface{} `json:"loadBalancer,omitempty"`
	Status       string                 `json:"status"`
	Using        []string               `json:"using,omitempty"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Type         string                 `json:"type,omitempty"`
	ServerStatus map[string]string      `json:"serverStatus,omitempty"`
}

// Scan implements sql.Scanner for TraefikOverview
func (o *TraefikOverview) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), o)
	}
	return json.Unmarshal(bytes, o)
}

// Value implements driver.Valuer for TraefikOverview
func (o TraefikOverview) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// Scan implements sql.Scanner for TraefikEntrypoint
func (e *TraefikEntrypoint) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), e)
	}
	return json.Unmarshal(bytes, e)
}

// Value implements driver.Valuer for TraefikEntrypoint
func (e TraefikEntrypoint) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements sql.Scanner for TraefikVersion
func (v *TraefikVersion) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), v)
	}
	return json.Unmarshal(bytes, v)
}

// Value implements driver.Valuer for TraefikVersion
func (v TraefikVersion) Value() (driver.Value, error) {
	return json.Marshal(v)
}

// FullTraefikData represents all data fetched from Traefik API
// This is used for concurrent fetching of all endpoints
type FullTraefikData struct {
	// HTTP Protocol
	HTTPRouters     []TraefikRouter     `json:"httpRouters"`
	HTTPServices    []TraefikService    `json:"httpServices"`
	HTTPMiddlewares []TraefikMiddleware `json:"httpMiddlewares"`

	// TCP Protocol
	TCPRouters     []TCPRouter     `json:"tcpRouters"`
	TCPServices    []TCPService    `json:"tcpServices"`
	TCPMiddlewares []TCPMiddleware `json:"tcpMiddlewares"`

	// UDP Protocol
	UDPRouters  []UDPRouter  `json:"udpRouters"`
	UDPServices []UDPService `json:"udpServices"`

	// Metadata
	Overview    *TraefikOverview    `json:"overview,omitempty"`
	Version     *TraefikVersion     `json:"version,omitempty"`
	Entrypoints []TraefikEntrypoint `json:"entrypoints,omitempty"`
}

// GetHTTPRouterCount returns the count of HTTP routers
func (d *FullTraefikData) GetHTTPRouterCount() int {
	return len(d.HTTPRouters)
}

// GetTCPRouterCount returns the count of TCP routers
func (d *FullTraefikData) GetTCPRouterCount() int {
	return len(d.TCPRouters)
}

// GetUDPRouterCount returns the count of UDP routers
func (d *FullTraefikData) GetUDPRouterCount() int {
	return len(d.UDPRouters)
}

// GetTotalRouterCount returns the total count of all routers
func (d *FullTraefikData) GetTotalRouterCount() int {
	return d.GetHTTPRouterCount() + d.GetTCPRouterCount() + d.GetUDPRouterCount()
}

// GetHTTPServiceCount returns the count of HTTP services
func (d *FullTraefikData) GetHTTPServiceCount() int {
	return len(d.HTTPServices)
}

// GetTCPServiceCount returns the count of TCP services
func (d *FullTraefikData) GetTCPServiceCount() int {
	return len(d.TCPServices)
}

// GetUDPServiceCount returns the count of UDP services
func (d *FullTraefikData) GetUDPServiceCount() int {
	return len(d.UDPServices)
}

// GetTotalServiceCount returns the total count of all services
func (d *FullTraefikData) GetTotalServiceCount() int {
	return d.GetHTTPServiceCount() + d.GetTCPServiceCount() + d.GetUDPServiceCount()
}

// GetHTTPMiddlewareCount returns the count of HTTP middlewares
func (d *FullTraefikData) GetHTTPMiddlewareCount() int {
	return len(d.HTTPMiddlewares)
}

// GetTCPMiddlewareCount returns the count of TCP middlewares
func (d *FullTraefikData) GetTCPMiddlewareCount() int {
	return len(d.TCPMiddlewares)
}

// GetTotalMiddlewareCount returns the total count of all middlewares
func (d *FullTraefikData) GetTotalMiddlewareCount() int {
	return d.GetHTTPMiddlewareCount() + d.GetTCPMiddlewareCount()
}
