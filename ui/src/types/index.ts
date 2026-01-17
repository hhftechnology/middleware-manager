// Resource types
export type {
  Resource,
  ResourceMiddleware,
  ResourceService,
  HTTPConfig,
  TLSConfig,
  TCPConfig,
  HeadersConfig,
  AssignMiddlewareRequest,
  AssignServiceRequest,
  MTLSWhitelistConfigRequest,
  MTLSWhitelistExternalData,
} from './resource'

// Middleware types
export type {
  MiddlewareType,
  Middleware,
  MiddlewareTemplate,
  CreateMiddlewareRequest,
  UpdateMiddlewareRequest,
} from './middleware'
export { MIDDLEWARE_TYPE_LABELS } from './middleware'

// Service types
export type {
  ServiceType,
  Service,
  LoadBalancerConfig,
  WeightedConfig,
  MirroringConfig,
  FailoverConfig,
  CreateServiceRequest,
  UpdateServiceRequest,
} from './service'
export { SERVICE_TYPE_LABELS } from './service'

// Data source types
export type {
  DataSourceType,
  DataSourceConfig,
  DataSourceInfo,
  SetActiveDataSourceRequest,
  UpdateDataSourceRequest,
  TestConnectionResponse,
} from './datasource'
export { DATA_SOURCE_TYPE_LABELS } from './datasource'

// Plugin types
export type {
  Plugin,
  PluginUsage,
  PluginInstallRequest,
  PluginRemoveRequest,
  PluginConfigPathResponse,
  PluginConfigPathRequest,
  CataloguePlugin,
} from './plugin'

// mTLS types
export type {
  MTLSConfig,
  MTLSClient,
  CreateCARequest,
  CreateClientRequest,
  MTLSConfigRequest,
  PluginCheckResponse,
  MTLSMiddlewareConfig,
} from './mtls'

// Traefik API types
export type {
  TraefikOverview,
  TraefikVersion,
  TraefikFeatures,
  TraefikEntrypoint,
  ProtocolOverview,
  StatusCount,
  HTTPRouter,
  TCPRouter,
  UDPRouter,
  HTTPService,
  TCPService,
  UDPService,
  HTTPMiddleware,
  TCPMiddleware,
  FullTraefikData,
  AllRoutersResponse,
  AllServicesResponse,
  AllMiddlewaresResponse,
  ProtocolType,
} from './traefik'

// Common types
export interface ApiError {
  message: string
  status: number
  details?: unknown
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
}

export type Page =
  | 'dashboard'
  | 'resources'
  | 'resource-detail'
  | 'middlewares'
  | 'middleware-form'
  | 'services'
  | 'service-form'
  | 'plugin-hub'
  | 'security'
