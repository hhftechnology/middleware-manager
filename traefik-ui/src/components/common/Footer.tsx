import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'

export function Footer() {
  const versionQuery = useQuery({ queryKey: ['manager-version'], queryFn: api.manager.version })
  const healthQuery = useQuery({ queryKey: ['health'], queryFn: api.health })
  return (
    <footer className="border-t border-border/80 bg-background/60 py-4">
      <div className="container flex flex-col gap-2 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <div>Traefik Manager · mode {healthQuery.data?.mode ?? '—'}</div>
        <div>Version {versionQuery.data?.version ?? 'unknown'}</div>
      </div>
    </footer>
  )
}
