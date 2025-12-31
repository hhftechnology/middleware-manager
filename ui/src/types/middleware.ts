export type MiddlewareType =
  | 'basicAuth'
  | 'digestAuth'
  | 'forwardAuth'
  | 'ipWhiteList'
  | 'ipAllowList'
  | 'rateLimit'
  | 'headers'
  | 'stripPrefix'
  | 'addPrefix'
  | 'replacePath'
  | 'replacePathRegex'
  | 'stripPrefixRegex'
  | 'redirectRegex'
  | 'redirectScheme'
  | 'chain'
  | 'plugin'
  | 'buffering'
  | 'circuitBreaker'
  | 'compress'
  | 'contentType'
  | 'errors'
  | 'grpcWeb'
  | 'inFlightReq'
  | 'passTLSClientCert'
  | 'retry'

export interface Middleware {
  id: string
  name: string
  type: MiddlewareType
  config: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

export interface MiddlewareTemplate {
  name: string
  type: MiddlewareType
  config: Record<string, unknown>
  description?: string
}

export interface CreateMiddlewareRequest {
  name: string
  type: MiddlewareType
  config: Record<string, unknown>
}

export interface UpdateMiddlewareRequest {
  name?: string
  type?: MiddlewareType
  config?: Record<string, unknown>
}

// Middleware type display names
export const MIDDLEWARE_TYPE_LABELS: Record<MiddlewareType, string> = {
  basicAuth: 'Basic Auth',
  digestAuth: 'Digest Auth',
  forwardAuth: 'Forward Auth',
  ipWhiteList: 'IP Whitelist',
  ipAllowList: 'IP Allowlist',
  rateLimit: 'Rate Limit',
  headers: 'Headers',
  stripPrefix: 'Strip Prefix',
  addPrefix: 'Add Prefix',
  replacePath: 'Replace Path',
  replacePathRegex: 'Replace Path (Regex)',
  stripPrefixRegex: 'Strip Prefix (Regex)',
  redirectRegex: 'Redirect (Regex)',
  redirectScheme: 'Redirect Scheme',
  chain: 'Chain',
  plugin: 'Plugin',
  buffering: 'Buffering',
  circuitBreaker: 'Circuit Breaker',
  compress: 'Compress',
  contentType: 'Content Type',
  errors: 'Error Pages',
  grpcWeb: 'gRPC Web',
  inFlightReq: 'In-Flight Requests',
  passTLSClientCert: 'Pass TLS Client Cert',
  retry: 'Retry',
}
