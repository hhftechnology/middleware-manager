import { useEffect } from 'react'
import { useResourceStore } from '@/stores/resourceStore'
import { useMiddlewareStore } from '@/stores/middlewareStore'
import { useServiceStore } from '@/stores/serviceStore'
import { useAppStore } from '@/stores/appStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { StatCard } from './StatCard'
import { ResourceSummary } from './ResourceSummary'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import {
  OverviewCards,
  EntrypointList,
  VersionInfo,
  RouterTabs,
  ServiceTabs,
  MiddlewareTabs,
} from '@/components/traefik'
import { Globe, Layers, Server, Plus, ArrowRight, Activity, Network } from 'lucide-react'

export function Dashboard() {
  const { navigateTo } = useAppStore()
  const {
    resources,
    loading: resourcesLoading,
    error: resourcesError,
    fetchResources,
  } = useResourceStore()
  const {
    middlewares,
    loading: middlewaresLoading,
    fetchMiddlewares,
  } = useMiddlewareStore()
  const {
    services,
    loading: servicesLoading,
    fetchServices,
  } = useServiceStore()

  useEffect(() => {
    fetchResources()
    fetchMiddlewares()
    fetchServices()
  }, [fetchResources, fetchMiddlewares, fetchServices])

  const isLoading = resourcesLoading || middlewaresLoading || servicesLoading

  if (isLoading && resources.length === 0) {
    return <PageLoader message="Loading dashboard..." />
  }

  if (resourcesError) {
    return (
      <div className="space-y-4">
        <ErrorMessage
          message={resourcesError}
          onRetry={() => {
            fetchResources()
            fetchMiddlewares()
            fetchServices()
          }}
        />
      </div>
    )
  }

  const activeResources = resources.filter((r) => r.status === 'active').length
  const tcpResources = resources.filter((r) => r.tcp_enabled).length

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between pb-2 border-b border-border/60">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Overview of your Traefik configuration
          </p>
        </div>
        <Button onClick={() => navigateTo('middleware-form')}>
          <Plus className="h-4 w-4 mr-2" />
          New Middleware
        </Button>
      </div>

      {/* Dashboard Tabs */}
      <Tabs defaultValue="overview" className="space-y-6">
        <TabsList>
          <TabsTrigger value="overview">
            <Activity className="h-4 w-4 mr-2" />
            Overview
          </TabsTrigger>
          <TabsTrigger value="traefik">
            <Globe className="h-4 w-4 mr-2" />
            Traefik Status
          </TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-8">
          {/* Stats Grid */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard
              title="Total Resources"
              value={resources.length}
              description={`${activeResources} active`}
              icon={Globe}
              color="blue"
              onClick={() => navigateTo('resources')}
            />
            <StatCard
              title="Middlewares"
              value={middlewares.length}
              description="Configured"
              icon={Layers}
              color="violet"
              onClick={() => navigateTo('middlewares')}
            />
            <StatCard
              title="Services"
              value={services.length}
              description="Defined"
              icon={Server}
              color="emerald"
              onClick={() => navigateTo('services')}
            />
            <StatCard
              title="TCP Routes"
              value={tcpResources}
              description="SNI-based"
              icon={Network}
              color="cyan"
            />
          </div>

          {/* Recent Resources */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle>Recent Resources</CardTitle>
                <CardDescription>
                  Latest resources from your data source
                </CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigateTo('resources')}
              >
                View All
                <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            </CardHeader>
            <CardContent>
              <ResourceSummary resources={resources} limit={5} />
            </CardContent>
          </Card>
        </TabsContent>

        {/* Traefik Status Tab */}
        <TabsContent value="traefik" className="space-y-6">
          <OverviewCards />

          <div className="grid gap-4 md:grid-cols-2">
            <VersionInfo />
            <EntrypointList />
          </div>

          <RouterTabs />
          <ServiceTabs />
          <MiddlewareTabs />
        </TabsContent>
      </Tabs>
    </div>
  )
}
