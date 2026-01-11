// Traefik API Types - Following Mantrae patterns

// Protocol Overview statistics
export interface ProtocolOverview {
  routers: StatusCount
  services: StatusCount
  middlewares: StatusCount
}

export interface StatusCount {
  total: number
  warnings: number
  errors: number
}

// Traefik Features
export interface TraefikFeatures {
  tracing?: string
  metrics?: string
  accessLog: boolean
}

// Overview response from /api/overview
export interface TraefikOverview {
  http: ProtocolOverview
  tcp: ProtocolOverview
  udp: ProtocolOverview
  features: TraefikFeatures
  providers: string[]
}

// Version response from /api/version
export interface TraefikVersion {
  version: string
  codename: string
  startDate?: string
  goVersion?: string
}

// TLS Configuration
export interface TLSConfig {
  certResolver?: string
  domains?: TLSDomain[]
}

export interface TLSDomain {
  main: string
  sans?: string[]
}

// TCP TLS Configuration
export interface TCPTLSConfig {
  passthrough?: boolean
  certResolver?: string
  domains?: string[]
  options?: string
}

// HTTP Router
export interface HTTPRouter {
  name: string
  entryPoints: string[]
  middlewares?: string[]
  service: string
  rule: string
  priority: number
  tls: TLSConfig
  status: string
  provider: string
}

// TCP Router
export interface TCPRouter {
  name: string
  rule: string
  service: string
  entryPoints: string[]
  middlewares?: string[]
  tls?: TCPTLSConfig
  priority: number
  provider: string
  status: string
}

// UDP Router
export interface UDPRouter {
  name: string
  service: string
  entryPoints: string[]
  provider: string
  status: string
}

// Load Balancer Server
export interface LoadBalancerServer {
  url?: string
  address?: string
  weight?: number
}

// HTTP Service
export interface HTTPService {
  name: string
  provider: string
  loadBalancer?: {
    servers?: LoadBalancerServer[]
    passHostHeader?: boolean
    sticky?: unknown
    healthCheck?: unknown
  }
  weighted?: {
    services?: Array<{ name: string; weight: number }>
    sticky?: unknown
    healthCheck?: unknown
  }
  mirroring?: {
    service: string
    mirrors?: Array<{ name: string; percent: number }>
    maxBodySize?: number
    mirrorBody?: boolean
    healthCheck?: unknown
  }
  failover?: {
    service: string
    fallback: string
    healthCheck?: unknown
  }
}

// TCP Service
export interface TCPService {
  name: string
  provider: string
  loadBalancer?: {
    servers?: Array<{ address: string; weight?: number }>
    terminationDelay?: number
  }
  weighted?: {
    services?: Array<{ name: string; weight: number }>
  }
}

// UDP Service
export interface UDPService {
  name: string
  provider: string
  loadBalancer?: {
    servers?: Array<{ address: string }>
  }
  weighted?: {
    services?: Array<{ name: string; weight: number }>
  }
}

// HTTP Middleware
export interface HTTPMiddleware {
  name: string
  type?: string
  provider?: string
  status?: string
  config?: Record<string, unknown>
}

// TCP Middleware
export interface TCPMiddleware {
  name: string
  type?: string
  provider?: string
  status?: string
  config?: Record<string, unknown>
  inFlightConn?: {
    amount: number
  }
  ipAllowList?: {
    sourceRange: string[]
  }
  ipWhiteList?: {
    sourceRange: string[]
  }
}

// Entrypoint Configuration
export interface TraefikEntrypoint {
  name: string
  address: string
  transport?: TransportConfig
  http?: EntrypointHTTPConfig
  http2?: EntrypointHTTP2Config
  http3?: EntrypointHTTP3Config
  udp?: EntrypointUDPConfig
}

export interface TransportConfig {
  lifeCycle?: {
    requestAcceptGraceTimeout?: string
    graceTimeOut?: string
  }
  respondingTimeouts?: {
    readTimeout?: string
    writeTimeout?: string
    idleTimeout?: string
  }
  proxyProtocol?: {
    insecure?: boolean
    trustedIPs?: string[]
  }
}

export interface EntrypointHTTPConfig {
  redirections?: {
    entryPoint?: {
      to?: string
      scheme?: string
      permanent?: boolean
      priority?: number
    }
  }
  middlewares?: string[]
  tls?: {
    options?: string
    certResolver?: string
    domains?: TLSDomain[]
  }
}

export interface EntrypointHTTP2Config {
  maxConcurrentStreams?: number
}

export interface EntrypointHTTP3Config {
  advertisedPort?: number
}

export interface EntrypointUDPConfig {
  timeout?: string
}

// Full Traefik Data response
export interface FullTraefikData {
  // HTTP Protocol
  httpRouters: HTTPRouter[]
  httpServices: HTTPService[]
  httpMiddlewares: HTTPMiddleware[]

  // TCP Protocol
  tcpRouters: TCPRouter[]
  tcpServices: TCPService[]
  tcpMiddlewares: TCPMiddleware[]

  // UDP Protocol
  udpRouters: UDPRouter[]
  udpServices: UDPService[]

  // Metadata
  overview?: TraefikOverview
  version?: TraefikVersion
  entrypoints?: TraefikEntrypoint[]
}

// Aggregated routers response
export interface AllRoutersResponse {
  http: HTTPRouter[]
  tcp: TCPRouter[]
  udp: UDPRouter[]
  total: {
    http: number
    tcp: number
    udp: number
  }
}

// Aggregated services response
export interface AllServicesResponse {
  http: HTTPService[]
  tcp: TCPService[]
  udp: UDPService[]
  total: {
    http: number
    tcp: number
    udp: number
  }
}

// Aggregated middlewares response
export interface AllMiddlewaresResponse {
  http: HTTPMiddleware[]
  tcp: TCPMiddleware[]
  total: {
    http: number
    tcp: number
  }
}

// Protocol type filter
export type ProtocolType = 'http' | 'tcp' | 'udp' | 'all'
