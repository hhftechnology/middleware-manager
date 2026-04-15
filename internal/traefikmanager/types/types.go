package types

import "time"

type RuntimeConfig struct {
	Port              string
	UIPath            string
	SettingsPath      string
	BackupDir         string
	ConfigDir         string
	ConfigPath        string
	ConfigPaths       []string
	TraefikAPIURL     string
	AcmeJSONPath      string
	AccessLogPath     string
	StaticConfigPath  string
	AllowCORS         bool
	CORSOrigin        string
	Debug             bool
	GitHubRepo        string
	SettingsDir       string
	GroupsConfigFile  string
	GroupsCacheDir    string
	HTTPClientTimeout time.Duration
}

type Settings struct {
	Domains        []string                 `json:"domains" yaml:"domains"`
	CertResolver   string                   `json:"cert_resolver" yaml:"cert_resolver"`
	TraefikAPIURL  string                   `json:"traefik_api_url" yaml:"traefik_api_url"`
	VisibleTabs    map[string]bool          `json:"visible_tabs" yaml:"visible_tabs"`
	DisabledRoutes map[string]DisabledRoute `json:"disabled_routes" yaml:"disabled_routes"`
	SelfRoute      SelfRoute                `json:"self_route" yaml:"self_route"`
	AcmeJSONPath   string                   `json:"acme_json_path" yaml:"acme_json_path"`
	AccessLogPath  string                   `json:"access_log_path" yaml:"access_log_path"`
	StaticConfig   string                   `json:"static_config_path" yaml:"static_config_path"`
}

type DisabledRoute struct {
	Protocol   string         `json:"protocol" yaml:"protocol"`
	Router     map[string]any `json:"router" yaml:"router"`
	Service    map[string]any `json:"service" yaml:"service"`
	ConfigFile string         `json:"configFile" yaml:"configFile"`
}

type SelfRoute struct {
	Domain     string `json:"domain" yaml:"domain"`
	ServiceURL string `json:"service_url" yaml:"service_url"`
	RouterName string `json:"router_name,omitempty" yaml:"router_name,omitempty"`
}

type App struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Rule               string   `json:"rule"`
	ServiceName        string   `json:"service_name"`
	Target             string   `json:"target"`
	Middlewares        []string `json:"middlewares"`
	EntryPoints        []string `json:"entryPoints"`
	Protocol           string   `json:"protocol"`
	TLS                bool     `json:"tls"`
	Enabled            bool     `json:"enabled"`
	PassHostHeader     bool     `json:"passHostHeader,omitempty"`
	CertResolver       string   `json:"certResolver,omitempty"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify,omitempty"`
	ConfigFile         string   `json:"configFile"`
	Provider           string   `json:"provider"`
}

type MiddlewareEntry struct {
	Name       string `json:"name"`
	YAML       string `json:"yaml"`
	Type       string `json:"type"`
	ConfigFile string `json:"configFile"`
}

type BackupEntry struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type DashboardConfig struct {
	CustomGroups   []map[string]any `json:"custom_groups" yaml:"custom_groups"`
	RouteOverrides map[string]any   `json:"route_overrides" yaml:"route_overrides"`
}

type RouteRequest struct {
	Protocol           string   `json:"protocol"`
	ConfigFile         string   `json:"configFile"`
	ServiceName        string   `json:"serviceName"`
	Domains            []string `json:"domains"`
	Subdomain          string   `json:"subdomain"`
	Rule               string   `json:"rule"`
	Target             string   `json:"target"`
	TargetPort         string   `json:"targetPort"`
	Scheme             string   `json:"scheme"`
	Middlewares        []string `json:"middlewares"`
	EntryPoints        []string `json:"entryPoints"`
	CertResolver       string   `json:"certResolver"`
	PassHostHeader     *bool    `json:"passHostHeader"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify"`
}

type MiddlewareRequest struct {
	Name         string `json:"name"`
	ConfigFile   string `json:"configFile"`
	YAML         string `json:"yaml"`
	OriginalName string `json:"originalName,omitempty"`
}

type ToggleRouteRequest struct {
	Enable bool `json:"enable"`
}

type SettingsRequest struct {
	Domains          []string        `json:"domains"`
	CertResolver     string          `json:"cert_resolver"`
	TraefikAPIURL    string          `json:"traefik_api_url"`
	VisibleTabs      map[string]bool `json:"visible_tabs"`
	AcmeJSONPath     string          `json:"acme_json_path"`
	AccessLogPath    string          `json:"access_log_path"`
	StaticConfigPath string          `json:"static_config_path"`
	SelfRoute        *SelfRoute      `json:"self_route,omitempty"`
}

type SelfRouteRequest struct {
	Domain     string `json:"domain"`
	ServiceURL string `json:"service_url"`
	RouterName string `json:"router_name"`
}

type TestConnectionRequest struct {
	URL string `json:"url"`
}

type CertificateInfo struct {
	Resolver string   `json:"resolver"`
	Main     string   `json:"main"`
	Sans     []string `json:"sans"`
	NotAfter string   `json:"not_after"`
	CertFile string   `json:"certFile,omitempty"`
}

type PluginInfo struct {
	Name       string         `json:"name"`
	ModuleName string         `json:"moduleName"`
	Version    string         `json:"version"`
	Settings   map[string]any `json:"settings,omitempty"`
}

var OptionalTabs = []string{
	"dashboard",
	"routemap",
	"docker",
	"kubernetes",
	"swarm",
	"nomad",
	"ecs",
	"consulcatalog",
	"redis",
	"etcd",
	"consul",
	"zookeeper",
	"http_provider",
	"file_external",
	"certs",
	"plugins",
	"logs",
}
