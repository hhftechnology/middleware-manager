import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import PluginsPage from './PluginsPage'

const plugins = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    traefik: {
      plugins: () => plugins(),
    },
  },
}))

function renderPage() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

  return render(
    <QueryClientProvider client={client}>
      <PluginsPage />
    </QueryClientProvider>,
  )
}

describe('PluginsPage', () => {
  beforeEach(() => {
    plugins.mockResolvedValue({
      plugins: [{ name: 'jwt', moduleName: 'github.com/example/jwt', version: 'v1.0.0', settings: { mode: 'strict' } }],
      error: '',
    })
  })

  it('renders plugin rows from static config', async () => {
    renderPage()
    expect(await screen.findByText('jwt')).toBeInTheDocument()
    expect(screen.getByText('v1.0.0')).toBeInTheDocument()
  })
})
