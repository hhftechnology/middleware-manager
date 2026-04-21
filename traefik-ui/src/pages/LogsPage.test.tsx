import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LogsPage from './LogsPage'

const logs = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    traefik: {
      logs: () => logs(),
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
      <LogsPage />
    </QueryClientProvider>,
  )
}

describe('LogsPage', () => {
  beforeEach(() => {
    logs.mockResolvedValue({ lines: ['line-a', 'line-b'] })
  })

  it('renders log lines and supports refresh', async () => {
    renderPage()
    expect(await screen.findByText(/line-a/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }))
    await waitFor(() => expect(logs).toHaveBeenCalledTimes(2))
  })
})
