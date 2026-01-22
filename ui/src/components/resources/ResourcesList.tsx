import { useEffect, useState } from 'react'
import { useResourceStore } from '@/stores/resourceStore'
import { useAppStore } from '@/stores/appStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { EmptyState } from '@/components/common/EmptyState'
import { ConfirmationModal } from '@/components/common/ConfirmationModal'
import { Search, ExternalLink, Trash2, Globe, RefreshCw } from 'lucide-react'
import { truncate } from '@/lib/utils'
import type { Resource } from '@/types'

export function ResourcesList() {
  const { navigateTo } = useAppStore()
  const {
    resources,
    loading,
    error,
    fetchResources,
    deleteResource,
    clearError,
  } = useResourceStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [resourceToDelete, setResourceToDelete] = useState<Resource | null>(null)

  useEffect(() => {
    fetchResources()
  }, [fetchResources])

  const filteredResources = resources.filter((resource) =>
    resource.host.toLowerCase().includes(searchTerm.toLowerCase()) ||
    resource.id.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const handleDelete = async () => {
    if (resourceToDelete) {
      await deleteResource(resourceToDelete.id)
      setResourceToDelete(null)
    }
  }

  const openDeleteModal = (resource: Resource) => {
    setResourceToDelete(resource)
    setDeleteModalOpen(true)
  }

  if (loading && resources.length === 0) {
    return <PageLoader message="Loading resources..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Resources</h1>
          <p className="text-muted-foreground">
            Manage your Traefik routes and configurations
          </p>
        </div>
        <Button variant="outline" onClick={() => fetchResources()}>
          <RefreshCw className="h-4 w-4 mr-2" />
          Refresh
        </Button>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onRetry={fetchResources}
          onDismiss={clearError}
        />
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Resources</CardTitle>
              <CardDescription>
                {filteredResources.length} of {resources.length} resources
              </CardDescription>
            </div>
            <div className="relative w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search resources..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-8"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredResources.length === 0 ? (
            <EmptyState
              icon={Globe}
              title="No resources found"
              description={
                searchTerm
                  ? 'Try adjusting your search terms'
                  : 'Resources will appear here when discovered from your data source'
              }
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Host</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Source</TableHead>
                  <TableHead>Entrypoints</TableHead>
                  <TableHead>Middlewares</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredResources.map((resource) => {
                  const middlewareCount = resource.middlewares
                    ? resource.middlewares.split(',').filter(Boolean).length
                    : 0

                  return (
                    <TableRow key={resource.id}>
                      <TableCell className="font-medium">
                        <div className="flex items-center gap-2">
                          {truncate(resource.host, 35)}
                          {resource.tcp_enabled && (
                            <Badge variant="secondary" className="text-xs">
                              TCP
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={resource.status === 'active' ? 'success' : 'secondary'}
                        >
                          {resource.status}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {resource.source_type || 'unknown'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">
                          {resource.entrypoints || 'websecure'}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{middlewareCount}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => navigateTo('resource-detail', resource.id)}
                          >
                            <ExternalLink className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => openDeleteModal(resource)}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Delete Confirmation Modal */}
      <ConfirmationModal
        open={deleteModalOpen}
        onOpenChange={setDeleteModalOpen}
        title="Delete Resource"
        description={`Are you sure you want to delete "${resourceToDelete?.host}"? This action cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </div>
  )
}
