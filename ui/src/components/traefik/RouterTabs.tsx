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
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useTraefikStore } from '@/stores/traefikStore'
import { TableSkeleton } from '@/components/loading-skeleton'
import type { HTTPRouter, TCPRouter, UDPRouter } from '@/types'

function StatusBadge({ status }: { status: string }) {
  const variant = status === 'enabled' ? 'default' : 'secondary'
  const className =
    status === 'enabled'
      ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
      : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400'

  return (
    <Badge variant={variant} className={className}>
      {status}
    </Badge>
  )
}

function ProviderBadge({ provider }: { provider: string }) {
  return (
    <Badge variant="outline" className="text-xs">
      {provider}
    </Badge>
  )
}

function HTTPRouterTable({ routers }: { routers: HTTPRouter[] }) {
  if (routers.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No HTTP routers found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Rule</TableHead>
          <TableHead>Service</TableHead>
          <TableHead>Entrypoints</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {routers.map((router) => (
          <TableRow key={router.name}>
            <TableCell className="font-medium">{router.name}</TableCell>
            <TableCell>
              <Tooltip>
                <TooltipTrigger asChild>
                  <code className="text-xs bg-muted px-1 py-0.5 rounded max-w-[200px] truncate block">
                    {router.rule}
                  </code>
                </TooltipTrigger>
                <TooltipContent side="bottom" className="max-w-md">
                  <code className="text-xs">{router.rule}</code>
                </TooltipContent>
              </Tooltip>
            </TableCell>
            <TableCell>{router.service}</TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {router.entryPoints?.map((ep) => (
                  <Badge key={ep} variant="secondary" className="text-xs">
                    {ep}
                  </Badge>
                ))}
              </div>
            </TableCell>
            <TableCell>
              <StatusBadge status={router.status} />
            </TableCell>
            <TableCell>
              <ProviderBadge provider={router.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function TCPRouterTable({ routers }: { routers: TCPRouter[] }) {
  if (routers.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No TCP routers found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Rule</TableHead>
          <TableHead>Service</TableHead>
          <TableHead>Entrypoints</TableHead>
          <TableHead>TLS</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {routers.map((router) => (
          <TableRow key={router.name}>
            <TableCell className="font-medium">{router.name}</TableCell>
            <TableCell>
              <code className="text-xs bg-muted px-1 py-0.5 rounded">
                {router.rule || 'HostSNI(`*`)'}
              </code>
            </TableCell>
            <TableCell>{router.service}</TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {router.entryPoints?.map((ep) => (
                  <Badge key={ep} variant="secondary" className="text-xs">
                    {ep}
                  </Badge>
                ))}
              </div>
            </TableCell>
            <TableCell>
              {router.tls ? (
                <Badge variant="outline" className="text-green-600">
                  {router.tls.passthrough ? 'Passthrough' : 'Terminate'}
                </Badge>
              ) : (
                <span className="text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell>
              <StatusBadge status={router.status} />
            </TableCell>
            <TableCell>
              <ProviderBadge provider={router.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function UDPRouterTable({ routers }: { routers: UDPRouter[] }) {
  if (routers.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No UDP routers found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Service</TableHead>
          <TableHead>Entrypoints</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {routers.map((router) => (
          <TableRow key={router.name}>
            <TableCell className="font-medium">{router.name}</TableCell>
            <TableCell>{router.service}</TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {router.entryPoints?.map((ep) => (
                  <Badge key={ep} variant="secondary" className="text-xs">
                    {ep}
                  </Badge>
                ))}
              </div>
            </TableCell>
            <TableCell>
              <StatusBadge status={router.status} />
            </TableCell>
            <TableCell>
              <ProviderBadge provider={router.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

export function RouterTabs() {
  const {
    httpRouters,
    tcpRouters,
    udpRouters,
    isLoadingRouters,
    fetchRouters,
  } = useTraefikStore()

  useEffect(() => {
    fetchRouters('all')
  }, [fetchRouters])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Routers</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="http" className="w-full">
          <TabsList>
            <TabsTrigger value="http">
              HTTP ({httpRouters.length})
            </TabsTrigger>
            <TabsTrigger value="tcp">
              TCP ({tcpRouters.length})
            </TabsTrigger>
            <TabsTrigger value="udp">
              UDP ({udpRouters.length})
            </TabsTrigger>
          </TabsList>
          <TabsContent value="http">
            {isLoadingRouters ? (
              <TableSkeleton rows={5} columns={6} />
            ) : (
              <HTTPRouterTable routers={httpRouters} />
            )}
          </TabsContent>
          <TabsContent value="tcp">
            {isLoadingRouters ? (
              <TableSkeleton rows={5} columns={7} />
            ) : (
              <TCPRouterTable routers={tcpRouters} />
            )}
          </TabsContent>
          <TabsContent value="udp">
            {isLoadingRouters ? (
              <TableSkeleton rows={5} columns={5} />
            ) : (
              <UDPRouterTable routers={udpRouters} />
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
