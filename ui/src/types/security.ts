// Security configuration types

export interface SecureHeadersConfig {
  x_content_type_options: string
  x_frame_options: string
  x_xss_protection: string
  hsts: string
  referrer_policy: string
  csp: string
  permissions_policy: string
}

export interface SecurityConfig {
  id: number
  tls_hardening_enabled: boolean
  secure_headers_enabled: boolean
  secure_headers: SecureHeadersConfig
  created_at?: string
  updated_at?: string
}

export interface DuplicateCheckResult {
  has_duplicates: boolean
  duplicates: Duplicate[]
  api_available: boolean
  warning_message?: string
}

export interface Duplicate {
  name: string
  provider: string
  type: string
}

export interface DuplicateCheckRequest {
  name: string
  plugin_name?: string
}

export interface UpdateResourceSecurityRequest {
  enabled: boolean
}

// Default secure headers configuration
export const defaultSecureHeaders: SecureHeadersConfig = {
  x_content_type_options: 'nosniff',
  x_frame_options: 'SAMEORIGIN',
  x_xss_protection: '1; mode=block',
  hsts: 'max-age=31536000; includeSubDomains',
  referrer_policy: 'strict-origin-when-cross-origin',
  csp: '',
  permissions_policy: '',
}
