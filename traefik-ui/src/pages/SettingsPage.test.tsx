import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import SettingsPage from './SettingsPage'

const settingsGet = vi.fn()
const settingsSave = vi.fn()
const settingsTestConnection = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    settings: {
      get: () => settingsGet(),
      save: (payload: unknown) => settingsSave(payload),
      testConnection: (url: string) => settingsTestConnection(url),
    },
  },
}))

function renderPage() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={client}>
      <SettingsPage />
    </QueryClientProvider>,
  )
}

describe('SettingsPage', () => {
  beforeEach(() => {
    settingsGet.mockResolvedValue({
      domains: ['example.com'],
      cert_resolver: 'cloudflare',
      traefik_api_url: 'http://traefik:8080',
      visible_tabs: { certs: true, logs: false, plugins: true },
      disabled_routes: {},
      self_route: { domain: 'manager.example.com', service_url: 'http://traefik-manager:5000', router_name: 'traefik-manager' },
      acme_json_path: '/app/acme.json',
      access_log_path: '/app/logs/access.log',
      static_config_path: '/app/traefik.yml',
    })
    settingsSave.mockResolvedValue({ success: true })
    settingsTestConnection.mockResolvedValue({ ok: true })
    vi.stubGlobal('alert', vi.fn())
  })

  it('loads settings and saves edited values', async () => {
    renderPage()

    const domains = await screen.findByDisplayValue('example.com')
    fireEvent.change(domains, { target: { value: 'example.com,internal.example.com' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save Settings' }))

    await waitFor(() => {
      expect(settingsSave).toHaveBeenCalledWith(expect.objectContaining({
        domains: ['example.com', 'internal.example.com'],
        traefik_api_url: 'http://traefik:8080',
      }))
    })
  })
})
