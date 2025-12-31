import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { Page } from '@/types'

interface AppState {
  // Navigation
  page: Page
  resourceId: string | null
  middlewareId: string | null
  serviceId: string | null
  isEditing: boolean

  // UI State
  isDarkMode: boolean
  showSettings: boolean
  sidebarCollapsed: boolean

  // Data source
  activeDataSource: string

  // Actions
  navigateTo: (page: Page, id?: string | null) => void
  setDarkMode: (enabled: boolean) => void
  toggleDarkMode: () => void
  setShowSettings: (show: boolean) => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setActiveDataSource: (source: string) => void
}

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Initial state
      page: 'dashboard',
      resourceId: null,
      middlewareId: null,
      serviceId: null,
      isEditing: false,
      isDarkMode: false,
      showSettings: false,
      sidebarCollapsed: false,
      activeDataSource: 'pangolin',

      // Navigation
      navigateTo: (page, id = null) => {
        const updates: Partial<AppState> = { page }

        switch (page) {
          case 'resource-detail':
            updates.resourceId = id
            updates.middlewareId = null
            updates.serviceId = null
            updates.isEditing = false
            break
          case 'middleware-form':
            updates.middlewareId = id
            updates.resourceId = null
            updates.serviceId = null
            updates.isEditing = !!id
            break
          case 'service-form':
            updates.serviceId = id
            updates.resourceId = null
            updates.middlewareId = null
            updates.isEditing = !!id
            break
          default:
            updates.resourceId = null
            updates.middlewareId = null
            updates.serviceId = null
            updates.isEditing = false
        }

        set(updates)
      },

      // Dark mode
      setDarkMode: (enabled) => {
        if (enabled) {
          document.documentElement.classList.add('dark')
        } else {
          document.documentElement.classList.remove('dark')
        }
        set({ isDarkMode: enabled })
      },

      toggleDarkMode: () => {
        const newValue = !get().isDarkMode
        if (newValue) {
          document.documentElement.classList.add('dark')
        } else {
          document.documentElement.classList.remove('dark')
        }
        set({ isDarkMode: newValue })
      },

      // Settings
      setShowSettings: (show) => set({ showSettings: show }),

      // Sidebar
      setSidebarCollapsed: (collapsed) => set({ sidebarCollapsed: collapsed }),

      // Data source
      setActiveDataSource: (source) => set({ activeDataSource: source }),
    }),
    {
      name: 'middleware-manager-app',
      partialize: (state) => ({
        isDarkMode: state.isDarkMode,
        sidebarCollapsed: state.sidebarCollapsed,
        activeDataSource: state.activeDataSource,
      }),
      onRehydrateStorage: () => (state) => {
        // Apply dark mode on rehydration
        if (state?.isDarkMode) {
          document.documentElement.classList.add('dark')
        }
      },
    }
  )
)
