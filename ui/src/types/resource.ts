export interface Resource {
  id: string
  host: string
  service_id: string
  org_id?: string
  site_id?: string
  status: 'active' | 'disabled'
  entrypoints: string
  tls_domains: string
  tcp_enabled: boolean
  tcp_entrypoints: string
  tcp_sni_rule: string
  custom_headers: string
  router_priority: number
  source_type: string
  mtls_enabled: boolean
  mtls_rules?: string
  mtls_request_headers?: string
  mtls_reject_message?: string
  mtls_reject_code?: number
  mtls_refresh_interval?: string
  mtls_external_data?: string
  middlewares: string
  created_at?: string
  updated_at?: string
}

export interface ResourceMiddleware {
  resource_id: string
  middleware_id: string
  priority: number
  created_at?: string
}

export interface ResourceService {
  resource_id: string
  service_id: string
  created_at?: string
}

export interface HTTPConfig {
  entrypoints: string  // comma-separated list
}

export interface TLSConfig {
  tls_domains: string  // comma-separated list
}

export interface TCPConfig {
  tcp_enabled?: boolean
  tcp_entrypoints: string  // comma-separated list
  tcp_sni_rule: string
}

export interface HeadersConfig {
  custom_headers: Record<string, string>
}

export interface MTLSWhitelistExternalData {
  url?: string
  skipTlsVerify?: boolean
  dataKey?: string
  headers?: Record<string, string>
  [key: string]: unknown
}

export interface MTLSWhitelistConfigRequest {
  rules?: unknown[]
  request_headers?: Record<string, string>
  reject_message?: string
  reject_code?: number
  refresh_interval?: string
  external_data?: MTLSWhitelistExternalData | Record<string, unknown>
}

export interface AssignMiddlewareRequest {
  middleware_id: string
  priority: number
}

export interface AssignServiceRequest {
  service_id: string
}
