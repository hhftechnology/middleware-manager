import { create } from 'zustand'
import { serviceApi } from '@/services/api'
import type { Service, CreateServiceRequest, UpdateServiceRequest } from '@/types'

interface ServiceState {
  // Data
  services: Service[]
  selectedService: Service | null

  // Loading states
  loading: boolean
  loadingService: boolean
  saving: boolean

  // Error state
  error: string | null

  // Actions
  fetchServices: () => Promise<void>
  fetchService: (id: string) => Promise<void>
  createService: (data: CreateServiceRequest) => Promise<Service | null>
  updateService: (id: string, data: UpdateServiceRequest) => Promise<boolean>
  deleteService: (id: string) => Promise<boolean>
  clearError: () => void
  clearSelectedService: () => void
}

export const useServiceStore = create<ServiceState>((set) => ({
  // Initial state
  services: [],
  selectedService: null,
  loading: false,
  loadingService: false,
  saving: false,
  error: null,

  // Fetch all services
  fetchServices: async () => {
    set({ loading: true, error: null })
    try {
      const services = await serviceApi.getAll()
      set({ services, loading: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load services',
        loading: false,
      })
    }
  },

  // Fetch single service
  fetchService: async (id) => {
    set({ loadingService: true, error: null })
    try {
      const service = await serviceApi.getById(id)
      set({ selectedService: service, loadingService: false })
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load service',
        loadingService: false,
      })
    }
  },

  // Create service
  createService: async (data) => {
    set({ saving: true, error: null })
    try {
      const service = await serviceApi.create(data)
      set((state) => ({
        services: [...state.services, service],
        saving: false,
      }))
      return service
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to create service',
        saving: false,
      })
      return null
    }
  },

  // Update service
  updateService: async (id, data) => {
    set({ saving: true, error: null })
    try {
      const updated = await serviceApi.update(id, data)
      set((state) => ({
        services: state.services.map((s) => (s.id === id ? updated : s)),
        selectedService: state.selectedService?.id === id ? updated : state.selectedService,
        saving: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to update service',
        saving: false,
      })
      return false
    }
  },

  // Delete service
  deleteService: async (id) => {
    set({ loading: true, error: null })
    try {
      await serviceApi.delete(id)
      set((state) => ({
        services: state.services.filter((s) => s.id !== id),
        selectedService: state.selectedService?.id === id ? null : state.selectedService,
        loading: false,
      }))
      return true
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to delete service',
        loading: false,
      })
      return false
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Clear selected service
  clearSelectedService: () => set({ selectedService: null }),
}))
