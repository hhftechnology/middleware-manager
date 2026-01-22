import { useEffect } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useTraefikStore } from '@/stores/traefikStore'
import { TableSkeleton } from '@/components/loading-skeleton'
import type { HTTPService, TCPService, UDPService } from '@/types'

function ProviderBadge({ provider }: { provider: string }) {
  return (
    <Badge variant="outline" className="text-xs">
      {provider}
    </Badge>
  )
}

function getServiceType(service: HTTPService | TCPService | UDPService): string {
  if ('loadBalancer' in service && service.loadBalancer) return 'Load Balancer'
  if ('weighted' in service && service.weighted) return 'Weighted'
  if ('mirroring' in service && service.mirroring) return 'Mirroring'
  if ('failover' in service && service.failover) return 'Failover'
  return 'Unknown'
}

function HTTPServiceTable({ services }: { services: HTTPService[] }) {
  if (services.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No HTTP services found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Servers</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {services.map((service) => (
          <TableRow key={service.name}>
            <TableCell className="font-medium">{service.name}</TableCell>
            <TableCell>
              <Badge variant="secondary">{getServiceType(service)}</Badge>
            </TableCell>
            <TableCell>
              {service.loadBalancer?.servers ? (
                <div className="space-y-1">
                  {service.loadBalancer.servers.slice(0, 3).map((server, idx) => (
                    <code key={idx} className="block text-xs bg-muted px-1 py-0.5 rounded">
                      {server.url || server.address}
                    </code>
                  ))}
                  {service.loadBalancer.servers.length > 3 && (
                    <span className="text-xs text-muted-foreground">
                      +{service.loadBalancer.servers.length - 3} more
                    </span>
                  )}
                </div>
              ) : (
                <span className="text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell>
              <ProviderBadge provider={service.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function TCPServiceTable({ services }: { services: TCPService[] }) {
  if (services.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No TCP services found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Servers</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {services.map((service) => (
          <TableRow key={service.name}>
            <TableCell className="font-medium">{service.name}</TableCell>
            <TableCell>
              <Badge variant="secondary">{getServiceType(service)}</Badge>
            </TableCell>
            <TableCell>
              {service.loadBalancer?.servers ? (
                <div className="space-y-1">
                  {service.loadBalancer.servers.slice(0, 3).map((server, idx) => (
                    <code key={idx} className="block text-xs bg-muted px-1 py-0.5 rounded">
                      {server.address}
                    </code>
                  ))}
                  {service.loadBalancer.servers.length > 3 && (
                    <span className="text-xs text-muted-foreground">
                      +{service.loadBalancer.servers.length - 3} more
                    </span>
                  )}
                </div>
              ) : (
                <span className="text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell>
              <ProviderBadge provider={service.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function UDPServiceTable({ services }: { services: UDPService[] }) {
  if (services.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No UDP services found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Servers</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {services.map((service) => (
          <TableRow key={service.name}>
            <TableCell className="font-medium">{service.name}</TableCell>
            <TableCell>
              <Badge variant="secondary">{getServiceType(service)}</Badge>
            </TableCell>
            <TableCell>
              {service.loadBalancer?.servers ? (
                <div className="space-y-1">
                  {service.loadBalancer.servers.slice(0, 3).map((server, idx) => (
                    <code key={idx} className="block text-xs bg-muted px-1 py-0.5 rounded">
                      {server.address}
                    </code>
                  ))}
                  {service.loadBalancer.servers.length > 3 && (
                    <span className="text-xs text-muted-foreground">
                      +{service.loadBalancer.servers.length - 3} more
                    </span>
                  )}
                </div>
              ) : (
                <span className="text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell>
              <ProviderBadge provider={service.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

export function ServiceTabs() {
  const {
    httpServices,
    tcpServices,
    udpServices,
    isLoadingServices,
    fetchServices,
  } = useTraefikStore()

  useEffect(() => {
    fetchServices('all')
  }, [fetchServices])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Services</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="http" className="w-full">
          <TabsList>
            <TabsTrigger value="http">
              HTTP ({httpServices.length})
            </TabsTrigger>
            <TabsTrigger value="tcp">
              TCP ({tcpServices.length})
            </TabsTrigger>
            <TabsTrigger value="udp">
              UDP ({udpServices.length})
            </TabsTrigger>
          </TabsList>
          <TabsContent value="http">
            {isLoadingServices ? (
              <TableSkeleton rows={5} columns={4} />
            ) : (
              <HTTPServiceTable services={httpServices} />
            )}
          </TabsContent>
          <TabsContent value="tcp">
            {isLoadingServices ? (
              <TableSkeleton rows={5} columns={4} />
            ) : (
              <TCPServiceTable services={tcpServices} />
            )}
          </TabsContent>
          <TabsContent value="udp">
            {isLoadingServices ? (
              <TableSkeleton rows={5} columns={4} />
            ) : (
              <UDPServiceTable services={udpServices} />
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
