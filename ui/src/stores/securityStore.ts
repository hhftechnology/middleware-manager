import { create } from 'zustand'
import type { SecurityConfig, SecureHeadersConfig, DuplicateCheckResult } from '@/types'
import { defaultSecureHeaders } from '@/types'
import { securityApi } from '@/services/api'

interface SecurityStore {
  // State
  config: SecurityConfig | null
  loading: boolean
  error: string | null
  duplicateCheckResult: DuplicateCheckResult | null
  duplicateCheckLoading: boolean

  // Actions
  fetchConfig: () => Promise<void>
  enableTLSHardening: () => Promise<boolean>
  disableTLSHardening: () => Promise<boolean>
  enableSecureHeaders: () => Promise<boolean>
  disableSecureHeaders: () => Promise<boolean>
  updateSecureHeadersConfig: (config: SecureHeadersConfig) => Promise<boolean>
  checkDuplicates: (name: string, pluginName?: string) => Promise<DuplicateCheckResult | null>
  clearError: () => void
  clearDuplicateCheck: () => void
}

export const useSecurityStore = create<SecurityStore>((set, get) => ({
  // Initial state
  config: null,
  loading: false,
  error: null,
  duplicateCheckResult: null,
  duplicateCheckLoading: false,

  // Fetch security configuration
  fetchConfig: async () => {
    set({ loading: true, error: null })
    try {
      const config = await securityApi.getConfig()
      set({ config, loading: false })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch security config'
      set({ error: message, loading: false })
      // Set defaults if fetch fails
      set({
        config: {
          id: 1,
          tls_hardening_enabled: false,
          secure_headers_enabled: false,
          secure_headers: defaultSecureHeaders,
        },
      })
    }
  },

  // Enable TLS hardening globally
  enableTLSHardening: async () => {
    set({ loading: true, error: null })
    try {
      await securityApi.enableTLSHardening()
      const currentConfig = get().config
      if (currentConfig) {
        set({
          config: { ...currentConfig, tls_hardening_enabled: true },
          loading: false,
        })
      }
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to enable TLS hardening'
      set({ error: message, loading: false })
      return false
    }
  },

  // Disable TLS hardening globally
  disableTLSHardening: async () => {
    set({ loading: true, error: null })
    try {
      await securityApi.disableTLSHardening()
      const currentConfig = get().config
      if (currentConfig) {
        set({
          config: { ...currentConfig, tls_hardening_enabled: false },
          loading: false,
        })
      }
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to disable TLS hardening'
      set({ error: message, loading: false })
      return false
    }
  },

  // Enable secure headers globally
  enableSecureHeaders: async () => {
    set({ loading: true, error: null })
    try {
      await securityApi.enableSecureHeaders()
      const currentConfig = get().config
      if (currentConfig) {
        set({
          config: { ...currentConfig, secure_headers_enabled: true },
          loading: false,
        })
      }
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to enable secure headers'
      set({ error: message, loading: false })
      return false
    }
  },

  // Disable secure headers globally
  disableSecureHeaders: async () => {
    set({ loading: true, error: null })
    try {
      await securityApi.disableSecureHeaders()
      const currentConfig = get().config
      if (currentConfig) {
        set({
          config: { ...currentConfig, secure_headers_enabled: false },
          loading: false,
        })
      }
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to disable secure headers'
      set({ error: message, loading: false })
      return false
    }
  },

  // Update secure headers configuration
  updateSecureHeadersConfig: async (headersConfig: SecureHeadersConfig) => {
    set({ loading: true, error: null })
    try {
      await securityApi.updateSecureHeadersConfig(headersConfig)
      const currentConfig = get().config
      if (currentConfig) {
        set({
          config: { ...currentConfig, secure_headers: headersConfig },
          loading: false,
        })
      }
      return true
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update secure headers config'
      set({ error: message, loading: false })
      return false
    }
  },

  // Check for middleware duplicates
  checkDuplicates: async (name: string, pluginName?: string) => {
    set({ duplicateCheckLoading: true, duplicateCheckResult: null })
    try {
      const result = await securityApi.checkDuplicates({ name, plugin_name: pluginName })
      set({ duplicateCheckResult: result, duplicateCheckLoading: false })
      return result
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to check for duplicates'
      set({
        duplicateCheckResult: {
          has_duplicates: false,
          duplicates: [],
          api_available: false,
          warning_message: message,
        },
        duplicateCheckLoading: false,
      })
      return null
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Clear duplicate check result
  clearDuplicateCheck: () => set({ duplicateCheckResult: null }),
}))
