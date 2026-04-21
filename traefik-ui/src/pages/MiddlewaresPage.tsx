import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { MiddlewareEntry, MiddlewareRequest } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { MiddlewareModal } from '@/components/modals/MiddlewareModal'
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
import { useToast } from '@/hooks/use-toast'

export default function MiddlewaresPage() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const middlewaresQuery = useQuery({ queryKey: ['middlewares'], queryFn: api.middlewares.list })
  const configsQuery = useQuery({ queryKey: ['configs'], queryFn: api.configs.list })
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<MiddlewareEntry | null>(null)
  const [viewing, setViewing] = useState<MiddlewareEntry | null>(null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['middlewares'] })

  const saveMutation = useMutation({
    mutationFn: async (payload: MiddlewareRequest) => {
      if (editing) await api.middlewares.update(editing.name, payload)
      else await api.middlewares.create(payload)
    },
    onSuccess: async () => {
      await invalidate()
      setModalOpen(false)
      setEditing(null)
      toast({ title: editing ? 'Middleware updated' : 'Middleware created' })
    },
    onError: (error) => {
      toast({
        title: 'Failed to save middleware',
        description: error instanceof Error ? error.message : String(error),
        variant: 'destructive',
      })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: ({ name, configFile }: { name: string; configFile?: string }) =>
      api.middlewares.delete(name, configFile),
    onSuccess: invalidate,
  })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Middlewares"
        description="Manage file-backed HTTP middlewares and inspect YAML blocks."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setModalOpen(true)
            }}
          >
            New Middleware
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <CardTitle>Middleware Inventory</CardTitle>
        </CardHeader>
        <CardContent>
          {middlewaresQuery.data?.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Config File</TableHead>
                  <TableHead>Preview</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {middlewaresQuery.data.map((item) => (
                  <TableRow key={`${item.name}-${item.configFile}`}>
                    <TableCell className="font-medium">{item.name}</TableCell>
                    <TableCell className="text-xs text-muted-foreground">
                      {item.configFile || 'default config'}
                    </TableCell>
                    <TableCell>
                      <pre className="max-w-xl whitespace-pre-wrap font-mono text-xs">{item.yaml}</pre>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex flex-wrap justify-end gap-2">
                        <Button variant="outline" size="sm" onClick={() => setViewing(item)}>
                          View
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setEditing(item)
                            setModalOpen(true)
                          }}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => deleteMutation.mutate({ name: item.name, configFile: item.configFile })}
                        >
                          Delete
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : middlewaresQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Loading middlewares...</p>
          ) : (
            <p className="text-sm text-muted-foreground">
              No middlewares found. Create one from the button above.
            </p>
          )}
        </CardContent>
      </Card>

      <MiddlewareModal
        key={editing?.name || 'new-middleware'}
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
        title={viewing?.name ?? 'Middleware Details'}
        description="Inspect the middleware YAML payload and source config metadata."
      >
        {viewing ? (
          <Card>
            <CardHeader>
              <CardTitle className="text-base">YAML</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="whitespace-pre-wrap rounded-md bg-muted p-4 font-mono text-xs">{viewing.yaml}</pre>
            </CardContent>
          </Card>
        ) : null}
      </DetailPanelsModal>
    </div>
  )
}
