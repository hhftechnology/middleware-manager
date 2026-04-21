import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import BackupsPage from './BackupsPage'

const backupsList = vi.fn()
const backupsCreate = vi.fn()
const backupsRestore = vi.fn()
const backupsRemove = vi.fn()

vi.mock('@/api/client', () => ({
  api: {
    backups: {
      list: () => backupsList(),
      create: () => backupsCreate(),
      restore: (name: string) => backupsRestore(name),
      remove: (name: string) => backupsRemove(name),
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
      <BackupsPage />
    </QueryClientProvider>,
  )
}

describe('BackupsPage', () => {
  beforeEach(() => {
    backupsList.mockResolvedValue([{ name: 'backup-1.tar.gz', size: 12, modified: '2026-01-01T00:00:00Z' }])
    backupsCreate.mockResolvedValue({ success: true, name: 'backup-2.tar.gz' })
    backupsRestore.mockResolvedValue({ success: true })
    backupsRemove.mockResolvedValue({ success: true })
  })

  it('renders backups and executes key actions', async () => {
    renderPage()
    expect(await screen.findByText('backup-1.tar.gz')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Create Backup' }))
    fireEvent.click(screen.getByRole('button', { name: 'Restore' }))
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }))
    await waitFor(() => {
      expect(backupsCreate).toHaveBeenCalled()
      expect(backupsRestore).toHaveBeenCalledWith('backup-1.tar.gz')
      expect(backupsRemove).toHaveBeenCalledWith('backup-1.tar.gz')
    })
  })
})
