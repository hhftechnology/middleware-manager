import { useEffect } from 'react'
import {
  Network,
  Server,
  Layers,
  AlertTriangle,
  XCircle,
  CheckCircle,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useTraefikStore } from '@/stores/traefikStore'
import { OverviewCardsSkeleton } from '@/components/loading-skeleton'
import type { ProtocolOverview } from '@/types'

interface StatCardProps {
  title: string
  icon: React.ReactNode
  stats: ProtocolOverview | undefined
  protocol: 'HTTP' | 'TCP' | 'UDP'
}

function StatCard({ title, icon, stats, protocol }: StatCardProps) {
  const total = stats?.routers?.total ?? 0
  const warnings = stats?.routers?.warnings ?? 0
  const errors = stats?.routers?.errors ?? 0
  const healthy = total - warnings - errors

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className="text-muted-foreground">{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{total}</div>
        <div className="mt-2 flex flex-wrap gap-2">
          {healthy > 0 && (
            <Tooltip>
              <TooltipTrigger>
                <Badge variant="outline" className="text-green-600 border-green-200 bg-green-50 dark:bg-green-900/20">
                  <CheckCircle className="mr-1 h-3 w-3" />
                  {healthy}
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                {healthy} healthy {protocol.toLowerCase()} routers
              </TooltipContent>
            </Tooltip>
          )}
          {warnings > 0 && (
            <Tooltip>
              <TooltipTrigger>
                <Badge variant="outline" className="text-yellow-600 border-yellow-200 bg-yellow-50 dark:bg-yellow-900/20">
                  <AlertTriangle className="mr-1 h-3 w-3" />
                  {warnings}
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                {warnings} {protocol.toLowerCase()} routers with warnings
              </TooltipContent>
            </Tooltip>
          )}
          {errors > 0 && (
            <Tooltip>
              <TooltipTrigger>
                <Badge variant="outline" className="text-red-600 border-red-200 bg-red-50 dark:bg-red-900/20">
                  <XCircle className="mr-1 h-3 w-3" />
                  {errors}
                </Badge>
              </TooltipTrigger>
              <TooltipContent>
                {errors} {protocol.toLowerCase()} routers with errors
              </TooltipContent>
            </Tooltip>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

interface ServiceStatCardProps {
  title: string
  icon: React.ReactNode
  stats: ProtocolOverview | undefined
  protocol: 'HTTP' | 'TCP' | 'UDP'
}

function ServiceStatCard({ title, icon, stats }: Omit<ServiceStatCardProps, 'protocol'>) {
  const total = stats?.services?.total ?? 0
  const warnings = stats?.services?.warnings ?? 0
  const errors = stats?.services?.errors ?? 0
  const healthy = total - warnings - errors

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className="text-muted-foreground">{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{total}</div>
        <div className="mt-2 flex flex-wrap gap-2">
          {healthy > 0 && (
            <Badge variant="outline" className="text-green-600 border-green-200 bg-green-50 dark:bg-green-900/20">
              <CheckCircle className="mr-1 h-3 w-3" />
              {healthy}
            </Badge>
          )}
          {warnings > 0 && (
            <Badge variant="outline" className="text-yellow-600 border-yellow-200 bg-yellow-50 dark:bg-yellow-900/20">
              <AlertTriangle className="mr-1 h-3 w-3" />
              {warnings}
            </Badge>
          )}
          {errors > 0 && (
            <Badge variant="outline" className="text-red-600 border-red-200 bg-red-50 dark:bg-red-900/20">
              <XCircle className="mr-1 h-3 w-3" />
              {errors}
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export function OverviewCards() {
  const { overview, isLoadingOverview, fetchOverview } = useTraefikStore()

  useEffect(() => {
    fetchOverview()
  }, [fetchOverview])

  if (isLoadingOverview) {
    return <OverviewCardsSkeleton />
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {/* Router Stats */}
      <StatCard
        title="HTTP Routers"
        icon={<Network className="h-4 w-4" />}
        stats={overview?.http}
        protocol="HTTP"
      />
      <StatCard
        title="TCP Routers"
        icon={<Network className="h-4 w-4" />}
        stats={overview?.tcp}
        protocol="TCP"
      />
      <StatCard
        title="UDP Routers"
        icon={<Network className="h-4 w-4" />}
        stats={overview?.udp}
        protocol="UDP"
      />

      {/* Service Stats */}
      <ServiceStatCard
        title="HTTP Services"
        icon={<Server className="h-4 w-4" />}
        stats={overview?.http}
      />
      <ServiceStatCard
        title="TCP Services"
        icon={<Server className="h-4 w-4" />}
        stats={overview?.tcp}
      />
      <ServiceStatCard
        title="UDP Services"
        icon={<Server className="h-4 w-4" />}
        stats={overview?.udp}
      />
    </div>
  )
}

export function MiddlewareOverviewCards() {
  const { overview } = useTraefikStore()

  const httpMiddlewares = overview?.http?.middlewares?.total ?? 0
  const tcpMiddlewares = overview?.tcp?.middlewares?.total ?? 0

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">HTTP Middlewares</CardTitle>
          <Layers className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{httpMiddlewares}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">TCP Middlewares</CardTitle>
          <Layers className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{tcpMiddlewares}</div>
        </CardContent>
      </Card>
    </div>
  )
}
