import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MiddlewaresPage from './MiddlewaresPage'

const middlewaresList = vi.fn()
const middlewaresCreate = vi.fn()
const middlewaresUpdate = vi.fn()
const middlewaresDelete = vi.fn()
const configsList = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    middlewares: {
      list: () => middlewaresList(),
      create: (payload: unknown) => middlewaresCreate(payload),
      update: (name: string, payload: unknown) => middlewaresUpdate(name, payload),
      delete: (name: string, configFile?: string) => middlewaresDelete(name, configFile),
    },
    configs: {
      list: () => configsList(),
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
      <MiddlewaresPage />
    </QueryClientProvider>,
  )
}

describe('MiddlewaresPage', () => {
  beforeEach(() => {
    middlewaresList.mockResolvedValue([
      { name: 'security-headers', configFile: 'dynamic.yml', yaml: 'headers: {}', type: 'headers' },
    ])
    configsList.mockResolvedValue({
      files: [{ label: 'dynamic.yml', path: '/tmp/dynamic.yml' }],
      configDirSet: false,
    })
    middlewaresDelete.mockResolvedValue({ ok: true })
  })

  it('renders inventory and deletes middleware', async () => {
    renderPage()
    expect(await screen.findByText('security-headers')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))
    await waitFor(() => expect(middlewaresDelete).toHaveBeenCalledWith('security-headers', 'dynamic.yml'))
  })
})
