export interface RouteApp {
  id: string
  name: string
  rule: string
  service_name: string
  target: string
  middlewares: string[]
  entryPoints: string[]
  protocol: 'http' | 'tcp' | 'udp'
  tls: boolean
  enabled: boolean
  passHostHeader?: boolean
  certResolver?: string
  insecureSkipVerify?: boolean
  configFile: string
  provider: string
}

export interface MiddlewareEntry {
  name: string
  yaml: string
  type: string
  configFile: string
}

export interface RouteResponse {
  apps: RouteApp[]
  middlewares: MiddlewareEntry[]
}

export interface ConfigFileEntry {
  label: string
  path: string
}

export interface ConfigListResponse {
  files: ConfigFileEntry[]
  configDirSet: boolean
}

export interface SelfRoute {
  domain: string
  service_url: string
  router_name?: string
}

export interface Settings {
  domains: string[]
  cert_resolver: string
  traefik_api_url: string
  visible_tabs: Record<string, boolean>
  disabled_routes: Record<string, unknown>
  self_route: SelfRoute
  acme_json_path: string
  access_log_path: string
  static_config_path: string
  auth_enabled: boolean
  has_password: boolean
  config_dir_set: boolean
}

export interface BackupEntry {
  name: string
  size: number
  modified: string
}

export interface CertificateInfo {
  resolver: string
  main: string
  sans: string[]
  not_after: string
  certFile?: string
}

export interface CertificatesResponse {
  certs: CertificateInfo[]
}

export interface LogsResponse {
  lines: string[]
  error?: string
}

export interface PluginInfo {
  name: string
  moduleName: string
  version: string
  settings?: Record<string, unknown>
}

export interface PluginsResponse {
  plugins: PluginInfo[]
  error?: string
}

export interface DashboardConfig {
  custom_groups: Array<Record<string, unknown>>
  route_overrides: Record<string, unknown>
}

export interface TraefikPing {
  ok: boolean
  latency_ms: number | null
}

export interface ManagerVersion {
  version: string
  repo: string
}

export interface RouteRequest {
  protocol: 'http' | 'tcp' | 'udp'
  configFile?: string
  serviceName: string
  domains: string[]
  subdomain: string
  rule: string
  target: string
  targetPort: string
  scheme: string
  middlewares: string[]
  entryPoints: string[]
  certResolver: string
  passHostHeader?: boolean
  insecureSkipVerify: boolean
}

export interface MiddlewareRequest {
  name: string
  configFile?: string
  yaml: string
  originalName?: string
}

export interface ApiErrorShape {
  code?: number
  message?: string
  details?: string
  error?: string
}

export interface ProviderFieldSchema {
  key: string
  label: string
  placeholder: string
  required?: boolean
}

export interface ProviderConfigDraft {
  enabled: boolean
  fields: Record<string, string>
}

export interface ServiceSummary {
  id: string
  name: string
  protocol: 'http' | 'tcp' | 'udp'
  target: string
  routeCount: number
  middlewares: string[]
  enabled: boolean
}

export interface RouteMapNode {
  id: string
  label: string
  kind: 'route' | 'service' | 'middleware'
}

export interface RouteMapEdge {
  id: string
  from: string
  to: string
}

export interface RouteMapGraph {
  nodes: RouteMapNode[]
  edges: RouteMapEdge[]
}
