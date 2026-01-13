import { create } from 'zustand'
import { mtlsApi } from '@/services/api'
import type { MTLSConfig, MTLSClient, CreateCARequest, CreateClientRequest } from '@/types'

interface MTLSState {
  // Data
  config: MTLSConfig | null
  clients: MTLSClient[]

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
  clearError: () => void
}

export const useMTLSStore = create<MTLSState>((set, get) => ({
  // Initial state
  config: null,
  clients: [],
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

  // Clear error
  clearError: () => set({ error: null }),
}))
