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

export interface AssignMiddlewareRequest {
  middleware_id: string
  priority: number
}

export interface AssignServiceRequest {
  service_id: string
}
