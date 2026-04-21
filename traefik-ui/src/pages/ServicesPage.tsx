import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { RouteApp, ServiceSummary } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { DetailPanelsModal } from '@/components/modals/DetailPanelsModal'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

function summarizeServices(routes: RouteApp[]): ServiceSummary[] {
  const map = new Map<string, ServiceSummary>()
  for (const route of routes) {
    const serviceKey = route.service_name || route.name
    const current = map.get(serviceKey)
    if (!current) {
      map.set(serviceKey, {
        id: serviceKey,
        name: serviceKey,
        protocol: route.protocol,
        target: route.target,
        routeCount: 1,
        middlewares: [...route.middlewares],
        enabled: route.enabled,
      })
      continue
    }

    current.routeCount += 1
    current.middlewares = Array.from(new Set([...current.middlewares, ...route.middlewares]))
    current.enabled = current.enabled || route.enabled
  }

  return Array.from(map.values())
}

export default function ServicesPage() {
  const routesQuery = useQuery({ queryKey: ['routes'], queryFn: api.routes.list })
  const [search, setSearch] = useState('')
  const [sortBy, setSortBy] = useState<'name' | 'routes'>('name')
  const [selected, setSelected] = useState<ServiceSummary | null>(null)

  const services = summarizeServices(routesQuery.data?.apps ?? [])
    .filter((item) => item.name.toLowerCase().includes(search.toLowerCase()))
    .sort((a, b) => {
      if (sortBy === 'routes') return b.routeCount - a.routeCount
      return a.name.localeCompare(b.name)
    })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Services"
        description="Service-level view aggregated from current route definitions."
      />
      <Card>
        <CardHeader className="gap-4 md:flex-row md:items-center md:justify-between">
          <CardTitle>Service Inventory</CardTitle>
          <div className="grid gap-3 sm:grid-cols-2">
            <Input
              placeholder="Search services..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
            <Select value={sortBy} onValueChange={(value) => setSortBy(value as 'name' | 'routes')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="name">Sort by Name</SelectItem>
                <SelectItem value="routes">Sort by Route Count</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          {services.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Service</TableHead>
                  <TableHead>Protocol</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>Routes</TableHead>
                  <TableHead>Middlewares</TableHead>
                  <TableHead className="text-right">Details</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {services.map((service) => (
                  <TableRow key={service.id}>
                    <TableCell className="font-medium">{service.name}</TableCell>
                    <TableCell className="uppercase">{service.protocol}</TableCell>
                    <TableCell className="font-mono text-xs">{service.target}</TableCell>
                    <TableCell>{service.routeCount}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {service.middlewares.length ? (
                          service.middlewares.map((mw) => <Badge key={mw}>{mw}</Badge>)
                        ) : (
                          <span className="text-xs text-muted-foreground">None</span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="outline" size="sm" onClick={() => setSelected(service)}>
                        View
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : routesQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Loading services...</p>
          ) : (
            <p className="text-sm text-muted-foreground">No services discovered from current routes.</p>
          )}
        </CardContent>
      </Card>

      <DetailPanelsModal
        open={Boolean(selected)}
        onOpenChange={(open) => {
          if (!open) setSelected(null)
        }}
        title={selected?.name ?? 'Service Details'}
        description="Route and middleware relationships for this service."
      >
        {selected ? (
          <div className="grid gap-4 sm:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Core</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <p><span className="font-medium">Protocol:</span> {selected.protocol.toUpperCase()}</p>
                <p><span className="font-medium">Target:</span> {selected.target}</p>
                <p><span className="font-medium">Routes:</span> {selected.routeCount}</p>
                <p><span className="font-medium">Enabled:</span> {selected.enabled ? 'yes' : 'no'}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Middlewares</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-1">
                  {selected.middlewares.length ? (
                    selected.middlewares.map((mw) => <Badge key={mw}>{mw}</Badge>)
                  ) : (
                    <span className="text-sm text-muted-foreground">No middleware links.</span>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        ) : null}
      </DetailPanelsModal>
    </div>
  )
}
