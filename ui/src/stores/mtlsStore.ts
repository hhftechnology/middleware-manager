import { create } from 'zustand'
import { mtlsApi } from '@/services/api'
import type { MTLSConfig, MTLSClient, CreateCARequest, CreateClientRequest, PluginCheckResponse, MTLSMiddlewareConfig } from '@/types'

interface MTLSState {
  // Data
  config: MTLSConfig | null
  clients: MTLSClient[]
  pluginStatus: PluginCheckResponse | null
  middlewareConfig: MTLSMiddlewareConfig | null

  // Loading states
  loading: boolean
  loadingClients: boolean

  // Error state
  error: string | null

  // Actions
  fetchConfig: () => Promise<void>
  enableMTLS: () => Promise<boolean>
  disableMTLS: () => Promise<boolean>
  createCA: (data: CreateCARequest) => Promise<boolean>
  deleteCA: () => Promise<boolean>
  fetchClients: () => Promise<void>
  createClient: (data: CreateClientRequest) => Promise<MTLSClient | null>
  revokeClient: (id: string) => Promise<boolean>
  deleteClient: (id: string) => Promise<boolean>
  checkPlugin: () => Promise<boolean>
  fetchMiddlewareConfig: () => Promise<void>
  updateMiddlewareConfig: (config: MTLSMiddlewareConfig) => Promise<boolean>
  clearError: () => void
}

export const useMTLSStore = create<MTLSState>((set, get) => ({
  // Initial state
  config: null,
  clients: [],
  pluginStatus: null,
  middlewareConfig: null,
  loading: false,
  loadingClients: false,
  error: null,

  // Fetch mTLS configuration
  fetchConfig: async () => {
    set({ loading: true, error: null })
    try {
      const config = await mtlsApi.getConfig()
      set({ config, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load mTLS configuration',
        loading: false,
      })
    }
  },

  // Enable mTLS globally
  enableMTLS: async () => {
    set({ error: null })
    try {
      await mtlsApi.enableMTLS()
      await get().fetchConfig()
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to enable mTLS',
      })
      return false
    }
  },

  // Disable mTLS globally
  disableMTLS: async () => {
    set({ error: null })
    try {
      await mtlsApi.disableMTLS()
      await get().fetchConfig()
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to disable mTLS',
      })
      return false
    }
  },

  // Create Certificate Authority
  createCA: async (data) => {
    set({ loading: true, error: null })
    try {
      await mtlsApi.createCA(data)
      await get().fetchConfig()
      set({ loading: false })
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create CA',
        loading: false,
      })
      return false
    }
  },

  // Delete Certificate Authority
  deleteCA: async () => {
    set({ loading: true, error: null })
    try {
      await mtlsApi.deleteCA()
      set({ clients: [] })
      await get().fetchConfig()
      set({ loading: false })
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete CA',
        loading: false,
      })
      return false
    }
  },

  // Fetch all client certificates
  fetchClients: async () => {
    set({ loadingClients: true, error: null })
    try {
      const clients = await mtlsApi.getClients()
      set({ clients, loadingClients: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load clients',
        loadingClients: false,
      })
    }
  },

  // Create a new client certificate
  createClient: async (data) => {
    set({ error: null })
    try {
      const client = await mtlsApi.createClient(data)
      await get().fetchClients()
      await get().fetchConfig() // Update client count
      return client
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create client certificate',
      })
      return null
    }
  },

  // Revoke a client certificate
  revokeClient: async (id) => {
    set({ error: null })
    try {
      await mtlsApi.revokeClient(id)
      // Optimistically mark revoked locally to avoid stale UI when fetch may fail
      set((state) => ({
        clients: state.clients.map((c) =>
          c.id === id ? { ...c, revoked: true, revoked_at: new Date().toISOString() } : c
        ),
      }))
      await get().fetchClients()
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to revoke client certificate',
      })
      return false
    }
  },

  // Delete a client certificate
  deleteClient: async (id) => {
    set({ error: null })
    try {
      await mtlsApi.deleteClient(id)
      set((state) => ({
        clients: state.clients.filter((c) => c.id !== id),
      }))
      await get().fetchConfig() // Update client count
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete client certificate',
      })
      return false
    }
  },

  // Check plugin status
  checkPlugin: async () => {
    try {
      const status = await mtlsApi.checkPlugin()
      // Normalize version display; default to recommended v0.0.4 if missing
      const version = status.version || 'v0.0.4'
      set({ pluginStatus: { ...status, version } })
      return status.installed
    } catch (err) {
      console.error('Failed to check plugin status:', err)
      set({ pluginStatus: { installed: false, plugin_name: 'mtlswhitelist', version: 'v0.0.4' } })
      return false
    }
  },

  // Fetch middleware configuration
  fetchMiddlewareConfig: async () => {
    try {
      const config = await mtlsApi.getMiddlewareConfig()
      set({ middlewareConfig: config })
    } catch (err) {
      console.error('Failed to fetch middleware config:', err)
    }
  },

  // Update middleware configuration
  updateMiddlewareConfig: async (config) => {
    set({ error: null })
    try {
      await mtlsApi.updateMiddlewareConfig(config)
      set({ middlewareConfig: config })
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update middleware configuration',
      })
      return false
    }
  },

  // Clear error
  clearError: () => set({ error: null }),
}))
