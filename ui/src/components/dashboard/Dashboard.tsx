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
import { Globe, Layers, Server, Plus, ArrowRight, Activity } from 'lucide-react'

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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            Overview of your Traefik configuration
          </p>
        </div>
        <div className="flex gap-2">
          <Button onClick={() => navigateTo('middleware-form')}>
            <Plus className="h-4 w-4 mr-2" />
            New Middleware
          </Button>
        </div>
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
        <TabsContent value="overview" className="space-y-6">
          {/* Stats Grid */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard
              title="Total Resources"
              value={resources.length}
              description={`${activeResources} active`}
              icon={Globe}
              onClick={() => navigateTo('resources')}
            />
            <StatCard
              title="Middlewares"
              value={middlewares.length}
              description="Available configurations"
              icon={Layers}
              onClick={() => navigateTo('middlewares')}
            />
            <StatCard
              title="Services"
              value={services.length}
              description="Load balancer configs"
              icon={Server}
              onClick={() => navigateTo('services')}
            />
            <StatCard
              title="TCP Routes"
              value={tcpResources}
              description="SNI-based routing"
              icon={Globe}
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

          {/* Quick Actions */}
          <div className="grid gap-4 md:grid-cols-3">
            <Card
              className="cursor-pointer hover:bg-accent/50 transition-colors"
              onClick={() => navigateTo('middlewares')}
            >
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Layers className="h-5 w-5" />
                  Manage Middlewares
                </CardTitle>
                <CardDescription>
                  Create and configure Traefik middlewares
                </CardDescription>
              </CardHeader>
            </Card>

            <Card
              className="cursor-pointer hover:bg-accent/50 transition-colors"
              onClick={() => navigateTo('services')}
            >
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Server className="h-5 w-5" />
                  Manage Services
                </CardTitle>
                <CardDescription>
                  Configure load balancing and routing
                </CardDescription>
              </CardHeader>
            </Card>

            <Card
              className="cursor-pointer hover:bg-accent/50 transition-colors"
              onClick={() => navigateTo('plugin-hub')}
            >
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Plus className="h-5 w-5" />
                  Plugin Hub
                </CardTitle>
                <CardDescription>
                  Discover and install Traefik plugins
                </CardDescription>
              </CardHeader>
            </Card>
          </div>
        </TabsContent>

        {/* Traefik Status Tab */}
        <TabsContent value="traefik" className="space-y-6">
          {/* Traefik Overview Cards */}
          <OverviewCards />

          {/* Version and Entrypoints */}
          <div className="grid gap-4 md:grid-cols-2">
            <VersionInfo />
            <EntrypointList />
          </div>

          {/* Routers */}
          <RouterTabs />

          {/* Services */}
          <ServiceTabs />

          {/* Middlewares */}
          <MiddlewareTabs />
        </TabsContent>
      </Tabs>
    </div>
  )
}
