import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { RouteApp, RouteMapGraph } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { DetailPanelsModal } from '@/components/modals/DetailPanelsModal'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

function buildRouteGraph(routes: RouteApp[]): RouteMapGraph {
  const nodes = new Map<string, { id: string; label: string; kind: 'route' | 'service' | 'middleware' }>()
  const edges = new Map<string, { id: string; from: string; to: string }>()

  for (const route of routes) {
    const routeNodeId = `route:${route.id}`
    const serviceName = route.service_name || route.name
    const serviceNodeId = `service:${serviceName}`

    nodes.set(routeNodeId, { id: routeNodeId, label: route.name, kind: 'route' })
    nodes.set(serviceNodeId, { id: serviceNodeId, label: serviceName, kind: 'service' })
    edges.set(`${routeNodeId}->${serviceNodeId}`, {
      id: `${routeNodeId}->${serviceNodeId}`,
      from: routeNodeId,
      to: serviceNodeId,
    })

    for (const middleware of route.middlewares) {
      const middlewareNodeId = `middleware:${middleware}`
      nodes.set(middlewareNodeId, { id: middlewareNodeId, label: middleware, kind: 'middleware' })
      edges.set(`${routeNodeId}->${middlewareNodeId}`, {
        id: `${routeNodeId}->${middlewareNodeId}`,
        from: routeNodeId,
        to: middlewareNodeId,
      })
    }
  }

  return { nodes: Array.from(nodes.values()), edges: Array.from(edges.values()) }
}

export default function RouteMapPage() {
  const routesQuery = useQuery({ queryKey: ['routes'], queryFn: api.routes.list })
  const graph = buildRouteGraph(routesQuery.data?.apps ?? [])
  const [selectedNode, setSelectedNode] = useState<RouteMapGraph['nodes'][number] | null>(null)

  const routes = graph.nodes.filter((node) => node.kind === 'route')
  const services = graph.nodes.filter((node) => node.kind === 'service')
  const middlewares = graph.nodes.filter((node) => node.kind === 'middleware')

  return (
    <div className="space-y-6">
      <PageHeader
        title="Route Map"
        description="Relationship map between routes, services, and middlewares from the current config."
      />
      <div className="grid gap-6 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Routes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {routes.length ? (
              routes.map((node) => (
                <button
                  key={node.id}
                  type="button"
                  className="w-full rounded-md border bg-card px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                  onClick={() => setSelectedNode(node)}
                >
                  {node.label}
                </button>
              ))
            ) : (
              <p className="text-sm text-muted-foreground">No route nodes.</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Services</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {services.length ? (
              services.map((node) => (
                <button
                  key={node.id}
                  type="button"
                  className="w-full rounded-md border bg-card px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                  onClick={() => setSelectedNode(node)}
                >
                  {node.label}
                </button>
              ))
            ) : (
              <p className="text-sm text-muted-foreground">No service nodes.</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Middlewares</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {middlewares.length ? (
              middlewares.map((node) => (
                <button
                  key={node.id}
                  type="button"
                  className="w-full rounded-md border bg-card px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                  onClick={() => setSelectedNode(node)}
                >
                  {node.label}
                </button>
              ))
            ) : (
              <p className="text-sm text-muted-foreground">No middleware nodes.</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Edges</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {graph.edges.length ? (
            graph.edges.map((edge) => (
              <div key={edge.id} className="rounded-md border bg-card px-3 py-2 text-sm">
                <span className="font-mono text-xs">{edge.from}</span> → <span className="font-mono text-xs">{edge.to}</span>
              </div>
            ))
          ) : routesQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Building route map...</p>
          ) : (
            <p className="text-sm text-muted-foreground">No route relationships found.</p>
          )}
        </CardContent>
      </Card>

      <DetailPanelsModal
        open={Boolean(selectedNode)}
        onOpenChange={(open) => {
          if (!open) setSelectedNode(null)
        }}
        title={selectedNode?.label ?? 'Node details'}
        description="Node classification and direct relationships."
      >
        {selectedNode ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Badge>{selectedNode.kind}</Badge>
              <span className="text-sm text-muted-foreground">{selectedNode.id}</span>
            </div>
            <div className="space-y-2">
              {graph.edges
                .filter((edge) => edge.from === selectedNode.id || edge.to === selectedNode.id)
                .map((edge) => (
                  <div key={edge.id} className="rounded-md border bg-card px-3 py-2 text-sm">
                    {edge.from} → {edge.to}
                  </div>
                ))}
              {!graph.edges.some((edge) => edge.from === selectedNode.id || edge.to === selectedNode.id) ? (
                <p className="text-sm text-muted-foreground">No direct relationships.</p>
              ) : null}
            </div>
            <div className="flex justify-end">
              <Button variant="outline" onClick={() => setSelectedNode(null)}>
                Close
              </Button>
            </div>
          </div>
        ) : null}
      </DetailPanelsModal>
    </div>
  )
}
