import { create } from 'zustand'
import { resourceApi } from '@/services/api'
import type {
  Resource,
  AssignMiddlewareRequest,
  HTTPConfig,
  TLSConfig,
  TCPConfig,
  HeadersConfig,
} from '@/types'

interface ResourceState {
  // Data
  resources: Resource[]
  selectedResource: Resource | null

  // Loading states
  loading: boolean
  loadingResource: boolean

  // Error state
  error: string | null

  // Actions
  fetchResources: () => Promise<void>
  fetchResource: (id: string) => Promise<void>
  deleteResource: (id: string) => Promise<boolean>
  assignMiddleware: (resourceId: string, data: AssignMiddlewareRequest) => Promise<boolean>
  removeMiddleware: (resourceId: string, middlewareId: string) => Promise<boolean>
  assignService: (resourceId: string, serviceId: string) => Promise<boolean>
  removeService: (resourceId: string) => Promise<boolean>
  updateHTTPConfig: (resourceId: string, config: HTTPConfig) => Promise<boolean>
  updateTLSConfig: (resourceId: string, config: TLSConfig) => Promise<boolean>
  updateTCPConfig: (resourceId: string, config: TCPConfig) => Promise<boolean>
  updateHeadersConfig: (resourceId: string, config: HeadersConfig) => Promise<boolean>
  updateRouterPriority: (resourceId: string, priority: number) => Promise<boolean>
  updateMTLSConfig: (resourceId: string, mtlsEnabled: boolean) => Promise<boolean>
  clearError: () => void
  clearSelectedResource: () => void
}

export const useResourceStore = create<ResourceState>((set, get) => ({
  // Initial state
  resources: [],
  selectedResource: null,
  loading: false,
  loadingResource: false,
  error: null,

  // Fetch all resources
  fetchResources: async () => {
    set({ loading: true, error: null })
    try {
      const resources = await resourceApi.getAll()
      set({ resources, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load resources',
        loading: false,
      })
    }
  },

  // Fetch single resource
  fetchResource: async (id) => {
    set({ loadingResource: true, error: null })
    try {
      const resource = await resourceApi.getById(id)
      set({ selectedResource: resource, loadingResource: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load resource',
        loadingResource: false,
      })
    }
  },

  // Delete resource
  deleteResource: async (id) => {
    set({ loading: true, error: null })
    try {
      await resourceApi.delete(id)
      set((state) => ({
        resources: state.resources.filter((r) => r.id !== id),
        selectedResource: state.selectedResource?.id === id ? null : state.selectedResource,
        loading: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete resource',
        loading: false,
      })
      return false
    }
  },

  // Assign middleware to resource
  assignMiddleware: async (resourceId, data) => {
    set({ error: null })
    try {
      await resourceApi.assignMiddleware(resourceId, data)
      // Refresh the resource to get updated middleware list
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to assign middleware',
      })
      return false
    }
  },

  // Remove middleware from resource
  removeMiddleware: async (resourceId, middlewareId) => {
    set({ error: null })
    try {
      await resourceApi.removeMiddleware(resourceId, middlewareId)
      // Refresh the resource
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to remove middleware',
      })
      return false
    }
  },

  // Assign service to resource
  assignService: async (resourceId, serviceId) => {
    set({ error: null })
    try {
      await resourceApi.assignService(resourceId, { service_id: serviceId })
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to assign service',
      })
      return false
    }
  },

  // Remove service from resource
  removeService: async (resourceId) => {
    set({ error: null })
    try {
      await resourceApi.removeService(resourceId)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to remove service',
      })
      return false
    }
  },

  // Update HTTP config
  updateHTTPConfig: async (resourceId, config) => {
    set({ error: null })
    try {
      await resourceApi.updateHTTPConfig(resourceId, config)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update HTTP config',
      })
      return false
    }
  },

  // Update TLS config
  updateTLSConfig: async (resourceId, config) => {
    set({ error: null })
    try {
      await resourceApi.updateTLSConfig(resourceId, config)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update TLS config',
      })
      return false
    }
  },

  // Update TCP config
  updateTCPConfig: async (resourceId, config) => {
    set({ error: null })
    try {
      await resourceApi.updateTCPConfig(resourceId, config)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update TCP config',
      })
      return false
    }
  },

  // Update headers config
  updateHeadersConfig: async (resourceId, config) => {
    set({ error: null })
    try {
      await resourceApi.updateHeadersConfig(resourceId, config)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update headers config',
      })
      return false
    }
  },

  // Update router priority
  updateRouterPriority: async (resourceId, priority) => {
    set({ error: null })
    try {
      await resourceApi.updateRouterPriority(resourceId, priority)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update router priority',
      })
      return false
    }
  },

  // Update mTLS config
  updateMTLSConfig: async (resourceId, mtlsEnabled) => {
    set({ error: null })
    try {
      await resourceApi.updateMTLSConfig(resourceId, mtlsEnabled)
      await get().fetchResource(resourceId)
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update mTLS config',
      })
      return false
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Clear selected resource
  clearSelectedResource: () => set({ selectedResource: null }),
}))
