import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { RouteApp, RouteRequest } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { RouteModal } from '@/components/modals/RouteModal'
import { DetailPanelsModal } from '@/components/modals/DetailPanelsModal'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/hooks/use-toast'

export default function RoutesPage() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const routesQuery = useQuery({ queryKey: ['routes'], queryFn: api.routes.list })
  const configsQuery = useQuery({ queryKey: ['configs'], queryFn: api.configs.list })
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<RouteApp | null>(null)
  const [viewing, setViewing] = useState<RouteApp | null>(null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['routes'] })

  const saveMutation = useMutation({
    mutationFn: async (payload: RouteRequest) => {
      if (editing) await api.routes.update(editing.id, payload)
      else await api.routes.create(payload)
    },
    onSuccess: async () => {
      await invalidate()
      setModalOpen(false)
      setEditing(null)
      toast({ title: editing ? 'Route updated' : 'Route created' })
    },
    onError: (error) => {
      toast({
        title: 'Failed to save route',
        description: error instanceof Error ? error.message : String(error),
        variant: 'destructive',
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.routes.delete(id),
    onSuccess: invalidate,
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enable }: { id: string; enable: boolean }) => api.routes.toggle(id, enable),
    onSuccess: invalidate,
  })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Routes"
        description="Create, edit, disable, and remove file-backed Traefik routes."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setModalOpen(true)
            }}
          >
            New Route
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle>Route Inventory</CardTitle>
        </CardHeader>
        <CardContent>
          {routesQuery.data?.apps.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Protocol</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>Rule</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {routesQuery.data.apps.map((route) => (
                  <TableRow key={route.id}>
                    <TableCell>
                      <div className="font-medium">{route.name}</div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        {route.configFile || 'default config'}
                      </div>
                    </TableCell>
                    <TableCell className="uppercase">{route.protocol}</TableCell>
                    <TableCell className="font-mono text-xs">{route.target}</TableCell>
                    <TableCell className="font-mono text-xs">{route.rule || 'n/a'}</TableCell>
                    <TableCell>
                      <Badge variant={route.enabled ? 'default' : 'secondary'}>
                        {route.enabled ? 'Enabled' : 'Disabled'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex flex-wrap justify-end gap-2">
                        <Button variant="outline" size="sm" onClick={() => setViewing(route)}>
                          View
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setEditing(route)
                            setModalOpen(true)
                          }}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => toggleMutation.mutate({ id: route.id, enable: !route.enabled })}
                        >
                          {route.enabled ? 'Disable' : 'Enable'}
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => deleteMutation.mutate(route.id)}
                        >
                          Delete
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : routesQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Loading routes...</p>
          ) : (
            <p className="text-sm text-muted-foreground">
              No routes found. Create one from the button above.
            </p>
          )}
        </CardContent>
      </Card>

      <RouteModal
        key={editing?.id || 'new-route'}
        open={modalOpen}
        onOpenChange={setModalOpen}
        configs={configsQuery.data?.files ?? []}
        editing={editing}
        onSubmit={(payload) => saveMutation.mutateAsync(payload)}
        pending={saveMutation.isPending}
      />

      <DetailPanelsModal
        open={Boolean(viewing)}
        onOpenChange={(open) => {
          if (!open) setViewing(null)
        }}
        title={viewing?.name ?? 'Route Details'}
        description="Inspect full route fields without leaving the inventory table."
      >
        {viewing ? (
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">General</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <p><span className="font-medium">ID:</span> {viewing.id}</p>
                <p><span className="font-medium">Protocol:</span> {viewing.protocol.toUpperCase()}</p>
                <p><span className="font-medium">Target:</span> {viewing.target}</p>
                <p><span className="font-medium">Rule:</span> {viewing.rule || 'n/a'}</p>
                <p><span className="font-medium">Provider:</span> {viewing.provider}</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Routing</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                <p><span className="font-medium">Entrypoints:</span> {viewing.entryPoints.join(', ') || 'n/a'}</p>
                <p><span className="font-medium">Middlewares:</span> {viewing.middlewares.join(', ') || 'n/a'}</p>
                <p><span className="font-medium">Cert Resolver:</span> {viewing.certResolver || 'n/a'}</p>
                <p><span className="font-medium">Pass Host Header:</span> {String(viewing.passHostHeader ?? true)}</p>
                <p><span className="font-medium">Insecure Skip Verify:</span> {String(viewing.insecureSkipVerify ?? false)}</p>
              </CardContent>
            </Card>
          </div>
        ) : null}
      </DetailPanelsModal>
    </div>
  )
}
