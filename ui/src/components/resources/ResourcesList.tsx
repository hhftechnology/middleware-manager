import { useEffect, useMemo, useState } from 'react'
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
import { Search, ExternalLink, Trash2, Globe, RefreshCw, CheckSquare, Square } from 'lucide-react'
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
    deleteDisabledResources,
    clearError,
  } = useResourceStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [resourceToDelete, setResourceToDelete] = useState<Resource | null>(null)
  const [selected, setSelected] = useState<Record<string, boolean>>({})
  const [bulkModalOpen, setBulkModalOpen] = useState(false)

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

  const disabledResources = useMemo(
    () => filteredResources.filter((r) => r.status === 'disabled'),
    [filteredResources]
  )

  const selectedIds = Object.entries(selected)
    .filter(([, checked]) => checked)
    .map(([id]) => id)

  const toggleSelect = (resource: Resource) => {
    if (resource.status !== 'disabled') return
    setSelected((prev) => ({ ...prev, [resource.id]: !prev[resource.id] }))
  }

  const toggleSelectAll = () => {
    const allSelected = selectedIds.length === disabledResources.length && disabledResources.length > 0
    if (allSelected) {
      setSelected({})
    } else {
      const next: Record<string, boolean> = {}
      disabledResources.forEach((r) => {
        next[r.id] = true
      })
      setSelected(next)
    }
  }

  const handleBulkDelete = async () => {
    if (selectedIds.length === 0) return
    await deleteDisabledResources(selectedIds)
    setSelected({})
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
      <div className="flex items-center justify-between pb-2 border-b border-border/60">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Resources</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Manage your Traefik routes and configurations
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant={selectedIds.length > 0 ? 'destructive' : 'outline'}
            disabled={selectedIds.length === 0}
            onClick={() => setBulkModalOpen(true)}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Delete disabled ({selectedIds.length})
          </Button>
          <Button variant="outline" onClick={() => fetchResources()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
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
                  <TableHead className="w-12">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={toggleSelectAll}
                      disabled={disabledResources.length === 0}
                      aria-label="Select all disabled"
                    >
                      {selectedIds.length === disabledResources.length && disabledResources.length > 0 ? (
                        <CheckSquare className="h-4 w-4" />
                      ) : (
                        <Square className="h-4 w-4" />
                      )}
                    </Button>
                  </TableHead>
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
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => toggleSelect(resource)}
                          disabled={resource.status !== 'disabled'}
                          aria-label="Select resource"
                        >
                          {selected[resource.id] ? (
                            <CheckSquare className="h-4 w-4" />
                          ) : (
                            <Square className="h-4 w-4" />
                          )}
                        </Button>
                      </TableCell>
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

      {/* Bulk Delete Confirmation Modal */}
      <ConfirmationModal
        open={bulkModalOpen}
        onOpenChange={setBulkModalOpen}
        title="Delete Disabled Resources"
        description={`Delete ${selectedIds.length} disabled resource(s)? This cannot be undone.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleBulkDelete}
      />
    </div>
  )
}
