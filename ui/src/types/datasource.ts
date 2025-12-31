export type DataSourceType = 'pangolin' | 'traefik'

export interface DataSourceConfig {
  name: string
  type: DataSourceType
  url: string
  basicAuth?: {
    username: string
    password: string
  }
  isActive?: boolean
}

export interface DataSourceInfo {
  name: string
  type: DataSourceType
  url: string
  isActive: boolean
  status?: 'connected' | 'disconnected' | 'error'
  lastSync?: string
}

export interface SetActiveDataSourceRequest {
  name: string
}

export interface UpdateDataSourceRequest {
  url?: string
  basicAuth?: {
    username: string
    password: string
  }
}

export interface TestConnectionResponse {
  success: boolean
  message?: string
  error?: string
}

// Data source type display names
export const DATA_SOURCE_TYPE_LABELS: Record<DataSourceType, string> = {
  pangolin: 'Pangolin',
  traefik: 'Traefik API',
}
