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
import type { HTTPMiddleware, TCPMiddleware } from '@/types'

function ProviderBadge({ provider }: { provider?: string }) {
  if (!provider) return null
  return (
    <Badge variant="outline" className="text-xs">
      {provider}
    </Badge>
  )
}

function StatusBadge({ status }: { status?: string }) {
  if (!status) return null
  const className =
    status === 'enabled'
      ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
      : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400'

  return (
    <Badge variant="default" className={className}>
      {status}
    </Badge>
  )
}

function HTTPMiddlewareTable({ middlewares }: { middlewares: HTTPMiddleware[] }) {
  if (middlewares.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No HTTP middlewares found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {middlewares.map((middleware) => (
          <TableRow key={middleware.name}>
            <TableCell className="font-medium">{middleware.name}</TableCell>
            <TableCell>
              {middleware.type ? (
                <Badge variant="secondary">{middleware.type}</Badge>
              ) : (
                <span className="text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell>
              <StatusBadge status={middleware.status} />
            </TableCell>
            <TableCell>
              <ProviderBadge provider={middleware.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

function TCPMiddlewareTable({ middlewares }: { middlewares: TCPMiddleware[] }) {
  if (middlewares.length === 0) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        No TCP middlewares found
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Provider</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {middlewares.map((middleware) => (
          <TableRow key={middleware.name}>
            <TableCell className="font-medium">{middleware.name}</TableCell>
            <TableCell>
              {middleware.type ? (
                <Badge variant="secondary">{middleware.type}</Badge>
              ) : middleware.inFlightConn ? (
                <Badge variant="secondary">InFlightConn</Badge>
              ) : middleware.ipAllowList ? (
                <Badge variant="secondary">IPAllowList</Badge>
              ) : (
                <span className="text-muted-foreground">-</span>
              )}
            </TableCell>
            <TableCell>
              <StatusBadge status={middleware.status} />
            </TableCell>
            <TableCell>
              <ProviderBadge provider={middleware.provider} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

export function MiddlewareTabs() {
  const {
    httpMiddlewares,
    tcpMiddlewares,
    isLoadingMiddlewares,
    fetchMiddlewares,
  } = useTraefikStore()

  useEffect(() => {
    fetchMiddlewares('all')
  }, [fetchMiddlewares])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Middlewares</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="http" className="w-full">
          <TabsList>
            <TabsTrigger value="http">
              HTTP ({httpMiddlewares.length})
            </TabsTrigger>
            <TabsTrigger value="tcp">
              TCP ({tcpMiddlewares.length})
            </TabsTrigger>
          </TabsList>
          <TabsContent value="http">
            {isLoadingMiddlewares ? (
              <TableSkeleton rows={5} columns={4} />
            ) : (
              <HTTPMiddlewareTable middlewares={httpMiddlewares} />
            )}
          </TabsContent>
          <TabsContent value="tcp">
            {isLoadingMiddlewares ? (
              <TableSkeleton rows={5} columns={4} />
            ) : (
              <TCPMiddlewareTable middlewares={tcpMiddlewares} />
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
