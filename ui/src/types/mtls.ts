// mTLS Configuration types

export interface MTLSConfig {
  id: number
  enabled: boolean
  ca_cert: string
  ca_cert_path: string
  ca_subject: string
  ca_expiry: string | null
  certs_base_path: string
  has_ca: boolean
  client_count: number
  created_at: string
  updated_at: string
}

export interface MTLSClient {
  id: string
  name: string
  cert: string
  p12_password_hint: string
  subject: string
  expiry: string | null
  revoked: boolean
  revoked_at: string | null
  created_at: string
}

export interface CreateCARequest {
  common_name: string
  organization?: string
  country?: string
  validity_days?: number
}

export interface CreateClientRequest {
  name: string
  validity_days?: number
  p12_password: string
}

export interface MTLSConfigRequest {
  mtls_enabled: boolean
}

export interface PluginCheckResponse {
  installed: boolean
  plugin_name: string
  version: string
}

export interface MTLSMiddlewareConfig {
  rules: string
  request_headers: string
  reject_message: string
  refresh_interval: number
}
