import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import { PageHeader } from '@/components/common/PageHeader'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface StatProps {
  label: string
  value: React.ReactNode
  hint?: string
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

export default function DashboardPage() {
  const overviewQuery = useQuery({ queryKey: ['overview'], queryFn: api.traefik.overview })
  const pingQuery = useQuery({ queryKey: ['ping'], queryFn: api.traefik.ping })
  const versionQuery = useQuery({ queryKey: ['manager-version'], queryFn: api.manager.version })
  const routesQuery = useQuery({ queryKey: ['routes'], queryFn: api.routes.list })

  const routeCount = routesQuery.data?.apps.length ?? 0
  const middlewareCount = routesQuery.data?.middlewares.length ?? 0
  const disabledCount = routesQuery.data?.apps.filter((route) => !route.enabled).length ?? 0

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Current Traefik Manager status, route inventory, and upstream health."
      />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Stat label="Routes" value={routeCount} hint="File-backed routes and disabled entries." />
        <Stat label="Middlewares" value={middlewareCount} hint="HTTP middleware blocks discovered from config files." />
        <Stat
          label="Traefik Ping"
          value={pingQuery.data?.latency_ms ?? 'n/a'}
          hint={pingQuery.data?.ok ? 'Milliseconds to /ping' : 'Traefik not reachable'}
        />
        <Stat
          label="Manager Version"
          value={versionQuery.data?.version || 'unknown'}
          hint={versionQuery.data?.repo || 'GitHub latest release'}
        />
      </div>
      <div className="grid gap-6 xl:grid-cols-[1.3fr_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Traefik Overview</CardTitle>
            <CardDescription>Raw payload returned by the configured Traefik API.</CardDescription>
          </CardHeader>
          <CardContent>
            {overviewQuery.data ? (
              <pre className="overflow-x-auto rounded-md bg-muted p-4 text-xs">
                {JSON.stringify(overviewQuery.data, null, 2)}
              </pre>
            ) : overviewQuery.isLoading ? (
              <p className="text-sm text-muted-foreground">Loading Traefik overview...</p>
            ) : (
              <p className="text-sm text-muted-foreground">Traefik did not return overview data.</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Current Signals</CardTitle>
            <CardDescription>Quick snapshot of manager state.</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm">
              <li className="rounded-md border bg-card px-3 py-2">
                Traefik API: <span className="font-medium">{pingQuery.data?.ok ? 'reachable' : 'unreachable'}</span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                Latest release: <span className="font-medium">{versionQuery.data?.version || 'unknown'}</span>
              </li>
              <li className="rounded-md border bg-card px-3 py-2">
                Disabled routes: <span className="font-medium">{disabledCount}</span>
              </li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
