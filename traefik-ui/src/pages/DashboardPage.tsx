import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { RouteApp } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface StatProps {
  label: string
  value: React.ReactNode
  hint?: string
}

interface MiddlewareStat {
  name: string
  count: number
}

function Stat({ label, value, hint }: StatProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>{label}</CardDescription>
        <CardTitle className="text-3xl font-semibold tracking-tight">{value}</CardTitle>
      </CardHeader>
      {hint ? (
        <CardContent>
          <p className="text-xs text-muted-foreground">{hint}</p>
        </CardContent>
      ) : null}
    </Card>
  )
}

function summarizeProtocols(routes: RouteApp[]): Record<string, number> {
  const counts: Record<string, number> = { http: 0, tcp: 0, udp: 0 }
  for (const route of routes) {
    const key = route.protocol.toLowerCase()
    counts[key] = (counts[key] ?? 0) + 1
  }
  return counts
}

function summarizeMiddlewares(routes: RouteApp[]): MiddlewareStat[] {
  const counts = new Map<string, number>()
  for (const route of routes) {
    for (const middleware of route.middlewares) {
      counts.set(middleware, (counts.get(middleware) ?? 0) + 1)
    }
  }
  return Array.from(counts.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
    .slice(0, 6)
}

export default function DashboardPage() {
  const overviewQuery = useQuery({ queryKey: ['overview'], queryFn: api.traefik.overview })
  const pingQuery = useQuery({ queryKey: ['ping'], queryFn: api.traefik.ping })
  const versionQuery = useQuery({ queryKey: ['manager-version'], queryFn: api.manager.version })
  const routesQuery = useQuery({ queryKey: ['routes'], queryFn: api.routes.list })

  const routes = routesQuery.data?.apps ?? []
  const middlewares = routesQuery.data?.middlewares ?? []
  const routeCount = routes.length
  const middlewareCount = middlewares.length
  const disabledCount = routes.filter((route) => !route.enabled).length
  const serviceCount = new Set(routes.map((route) => route.service_name || route.name)).size
  const protocolCounts = summarizeProtocols(routes)
  const topMiddlewares = summarizeMiddlewares(routes)
  const overviewSections = overviewQuery.data ? Object.keys(overviewQuery.data).sort() : []
  const pingHealthy = pingQuery.data?.ok ?? false

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Current Traefik Manager status, route inventory, and upstream health."
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Stat label="Routes" value={routeCount} hint="File-backed routes and disabled entries." />
        <Stat label="Services" value={serviceCount} hint="Unique services inferred from active routes." />
        <Stat
          label="Middlewares"
          value={middlewareCount}
          hint="HTTP middleware blocks discovered from config files."
        />
        <Stat
          label="Traefik API"
          value={pingHealthy ? 'Reachable' : 'Unreachable'}
          hint={pingHealthy ? `${pingQuery.data?.latency_ms ?? 'n/a'} ms probe latency` : 'Check Traefik API URL in Settings.'}
        />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.3fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Traffic Snapshot</CardTitle>
            <CardDescription>Protocol distribution and middleware usage across known routes.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-3 sm:grid-cols-3">
              <div className="rounded-md border bg-card px-3 py-2 text-sm">
                <div className="text-xs text-muted-foreground">HTTP Routes</div>
                <div className="text-xl font-semibold">{protocolCounts.http ?? 0}</div>
              </div>
              <div className="rounded-md border bg-card px-3 py-2 text-sm">
                <div className="text-xs text-muted-foreground">TCP Routes</div>
                <div className="text-xl font-semibold">{protocolCounts.tcp ?? 0}</div>
              </div>
              <div className="rounded-md border bg-card px-3 py-2 text-sm">
                <div className="text-xs text-muted-foreground">UDP Routes</div>
                <div className="text-xl font-semibold">{protocolCounts.udp ?? 0}</div>
              </div>
            </div>
            <div>
              <div className="mb-2 text-sm font-medium">Top Middleware Links</div>
              {topMiddlewares.length ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Middleware</TableHead>
                      <TableHead className="text-right">Linked Routes</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {topMiddlewares.map((item) => (
                      <TableRow key={item.name}>
                        <TableCell className="font-medium">{item.name}</TableCell>
                        <TableCell className="text-right">{item.count}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <p className="text-sm text-muted-foreground">No middleware links found in current routes.</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Runtime Health</CardTitle>
            <CardDescription>Connectivity and runtime status signals.</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              <li className="rounded-md border bg-card px-3 py-2">
                Traefik API:
                {' '}
                <span className="font-medium">{pingHealthy ? 'reachable' : 'unreachable'}</span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                Probe latency:
                {' '}
                <span className="font-medium">{pingHealthy ? `${pingQuery.data?.latency_ms ?? 'n/a'} ms` : 'n/a'}</span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                Manager version:
                {' '}
                <span className="font-medium">{versionQuery.data?.version || 'unknown'}</span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                Disabled routes:
                {' '}
                <span className="font-medium">{disabledCount}</span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                Overview status:
                {' '}
                <span className="font-medium">
                  {overviewQuery.isLoading ? 'loading' : overviewSections.length ? 'available' : 'missing'}
                </span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                <div className="mb-1 text-xs text-muted-foreground">Overview sections</div>
                <div className="flex flex-wrap gap-1">
                  {overviewSections.length ? (
                    overviewSections.slice(0, 8).map((section) => (
                      <Badge key={section} variant="secondary">
                        {section}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-xs text-muted-foreground">No overview sections returned.</span>
                  )}
                </div>
              </li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
