export type ServiceType = 'loadBalancer' | 'weighted' | 'mirroring' | 'failover'

export interface Service {
  id: string
  name: string
  type: ServiceType
  config: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

export interface LoadBalancerConfig {
  servers: Array<{
    url: string
    weight?: number
  }>
  healthCheck?: {
    path?: string
    interval?: string
    timeout?: string
  }
  sticky?: {
    cookie?: {
      name?: string
      secure?: boolean
      httpOnly?: boolean
    }
  }
  passHostHeader?: boolean
  serversTransport?: string
}

export interface WeightedConfig {
  services: Array<{
    name: string
    weight: number
  }>
  sticky?: {
    cookie?: {
      name?: string
    }
  }
}

export interface MirroringConfig {
  service: string
  mirrors: Array<{
    name: string
    percent: number
  }>
  maxBodySize?: number
}

export interface FailoverConfig {
  service: string
  fallback: string
  healthCheck?: {
    path?: string
    interval?: string
  }
}

export interface CreateServiceRequest {
  name: string
  type: ServiceType
  config: Record<string, unknown>
}

export interface UpdateServiceRequest {
  name?: string
  type?: ServiceType
  config?: Record<string, unknown>
}

// Service type display names
export const SERVICE_TYPE_LABELS: Record<ServiceType, string> = {
  loadBalancer: 'Load Balancer',
  weighted: 'Weighted',
  mirroring: 'Mirroring',
  failover: 'Failover',
}
