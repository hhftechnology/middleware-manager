import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CertificatesPage from './CertificatesPage'

const certs = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    traefik: {
      certs: () => certs(),
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
      <CertificatesPage />
    </QueryClientProvider>,
  )
}

describe('CertificatesPage', () => {
  beforeEach(() => {
    certs.mockResolvedValue({
      certs: [
        {
          resolver: 'cloudflare',
          main: 'example.com',
          sans: ['www.example.com'],
          not_after: '2026-12-30',
          certFile: '/etc/certs/example.pem',
        },
      ],
    })
  })

  it('renders certificate table rows', async () => {
    renderPage()
    expect(await screen.findByText('example.com')).toBeInTheDocument()
    expect(screen.getByText('cloudflare')).toBeInTheDocument()
  })
})
