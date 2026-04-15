import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { PageHeader } from '@/components/common/PageHeader'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

export default function BackupsPage() {
  const queryClient = useQueryClient()
  const backupsQuery = useQuery({ queryKey: ['backups'], queryFn: api.backups.list })
  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['backups'] })

  const createMutation = useMutation({ mutationFn: () => api.backups.create(), onSuccess: invalidate })
  const restoreMutation = useMutation({ mutationFn: (name: string) => api.backups.restore(name), onSuccess: invalidate })
  const deleteMutation = useMutation({ mutationFn: (name: string) => api.backups.remove(name), onSuccess: invalidate })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Backups"
        description="Create manual backups before mutations and restore or prune old snapshots."
        actions={
          <Button onClick={() => createMutation.mutate()} disabled={createMutation.isPending}>
            Create Backup
          </Button>
        }
      />
      <Card>
        <CardContent className="pt-6">
          {backupsQuery.data?.length ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Size</TableHead>
                  <TableHead>Modified</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {backupsQuery.data.map((backup) => (
                  <TableRow key={backup.name}>
                    <TableCell className="font-medium">{backup.name}</TableCell>
                    <TableCell>{backup.size}</TableCell>
                    <TableCell>{backup.modified}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button variant="outline" size="sm" onClick={() => restoreMutation.mutate(backup.name)}>
                          Restore
                        </Button>
                        <Button variant="destructive" size="sm" onClick={() => deleteMutation.mutate(backup.name)}>
                          Delete
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : backupsQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">Loading backups...</p>
          ) : (
            <p className="text-sm text-muted-foreground">
              No backups yet. Create one before editing dynamic config files.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
