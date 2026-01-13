import { useEffect, useState, useMemo } from 'react'
import { useMiddlewareStore } from '@/stores/middlewareStore'
import { useResourceStore } from '@/stores/resourceStore'
import { useAppStore } from '@/stores/appStore'
import { toast } from '@/hooks/use-toast'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { EmptyState } from '@/components/common/EmptyState'
import { ConfirmationModal } from '@/components/common/ConfirmationModal'
import {
  Search,
  Plus,
  Edit,
  Trash2,
  Layers,
  RefreshCw,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Link,
  Loader2,
} from 'lucide-react'
import { MIDDLEWARE_TYPE_LABELS } from '@/types'
import type { Middleware, MiddlewareType } from '@/types'

type SortField = 'name' | 'type' | 'id'
type SortOrder = 'asc' | 'desc'

export function MiddlewaresList() {
  const { navigateTo } = useAppStore()
  const {
    middlewares,
    loading,
    error,
    fetchMiddlewares,
    deleteMiddleware,
    clearError,
  } = useMiddlewareStore()
  const {
    resources,
    fetchResources,
    assignMiddleware,
  } = useResourceStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [sortField, setSortField] = useState<SortField>('name')
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc')
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [middlewareToDelete, setMiddlewareToDelete] = useState<Middleware | null>(null)
  const [bulkDeleteModalOpen, setBulkDeleteModalOpen] = useState(false)
  const [bulkAssignModalOpen, setBulkAssignModalOpen] = useState(false)
  const [selectedResourceIds, setSelectedResourceIds] = useState<Set<string>>(new Set())
  const [bulkAssigning, setBulkAssigning] = useState(false)
  const [bulkDeleting, setBulkDeleting] = useState(false)

  useEffect(() => {
    fetchMiddlewares()
    fetchResources()
  }, [fetchMiddlewares, fetchResources])

  // Get unique middleware types for filter
  const middlewareTypes = useMemo(() => {
    const types = new Set(middlewares.map((mw) => mw.type))
    return Array.from(types).sort()
  }, [middlewares])

  // Filter and sort middlewares
  const filteredMiddlewares = useMemo(() => {
    let result = middlewares.filter((mw) => {
      const matchesSearch =
        mw.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        mw.type.toLowerCase().includes(searchTerm.toLowerCase()) ||
        mw.id.toLowerCase().includes(searchTerm.toLowerCase())

      const matchesType = typeFilter === 'all' || mw.type === typeFilter

      return matchesSearch && matchesType
    })

    // Sort
    result = result.sort((a, b) => {
      let comparison = 0
      switch (sortField) {
        case 'name':
          comparison = a.name.localeCompare(b.name)
          break
        case 'type':
          comparison = a.type.localeCompare(b.type)
          break
        case 'id':
          comparison = a.id.localeCompare(b.id)
          break
      }
      return sortOrder === 'asc' ? comparison : -comparison
    })

    return result
  }, [middlewares, searchTerm, typeFilter, sortField, sortOrder])

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortOrder('asc')
    }
  }

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) return <ArrowUpDown className="h-4 w-4 ml-1" />
    return sortOrder === 'asc' ? (
      <ArrowUp className="h-4 w-4 ml-1" />
    ) : (
      <ArrowDown className="h-4 w-4 ml-1" />
    )
  }

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(filteredMiddlewares.map((mw) => mw.id)))
    } else {
      setSelectedIds(new Set())
    }
  }

  const handleSelectOne = (id: string, checked: boolean) => {
    const newSelected = new Set(selectedIds)
    if (checked) {
      newSelected.add(id)
    } else {
      newSelected.delete(id)
    }
    setSelectedIds(newSelected)
  }

  const handleDelete = async () => {
    if (middlewareToDelete) {
      const success = await deleteMiddleware(middlewareToDelete.id)
      if (success) {
        setMiddlewareToDelete(null)
      } else {
        // Get the error from the store
        const currentError = useMiddlewareStore.getState().error
        toast({
          title: 'Cannot delete middleware',
          description: currentError || 'This middleware may be assigned to one or more resources. Remove it from all resources first.',
          variant: 'destructive',
        })
        setMiddlewareToDelete(null)
      }
    }
  }

  const handleBulkDelete = async () => {
    setBulkDeleting(true)
    const idsToDelete = Array.from(selectedIds)
    const failedIds: string[] = []
    for (const id of idsToDelete) {
      const success = await deleteMiddleware(id)
      if (!success) {
        failedIds.push(id)
      }
    }
    if (failedIds.length > 0) {
      toast({
        title: 'Some middlewares could not be deleted',
        description: `${failedIds.length} middleware(s) are assigned to resources and cannot be deleted. Remove them from all resources first.`,
        variant: 'destructive',
      })
    }
    setSelectedIds(new Set())
    setBulkDeleting(false)
    setBulkDeleteModalOpen(false)
  }

  const handleBulkAssign = async () => {
    if (selectedResourceIds.size === 0 || selectedIds.size === 0) return

    setBulkAssigning(true)
    const middlewareIds = Array.from(selectedIds)
    const resourceIds = Array.from(selectedResourceIds)

    for (const resourceId of resourceIds) {
      for (const middlewareId of middlewareIds) {
        await assignMiddleware(resourceId, {
          middleware_id: middlewareId,
          priority: 100,
        })
      }
    }

    setBulkAssigning(false)
    setBulkAssignModalOpen(false)
    setSelectedResourceIds(new Set())
    setSelectedIds(new Set())
  }

  const openDeleteModal = (middleware: Middleware) => {
    setMiddlewareToDelete(middleware)
    setDeleteModalOpen(true)
  }

  const isAllSelected = filteredMiddlewares.length > 0 &&
    filteredMiddlewares.every((mw) => selectedIds.has(mw.id))

  if (loading && middlewares.length === 0) {
    return <PageLoader message="Loading middlewares..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Middlewares</h1>
          <p className="text-muted-foreground">
            Create and manage Traefik middleware configurations
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => fetchMiddlewares()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => navigateTo('middleware-form')}>
            <Plus className="h-4 w-4 mr-2" />
            New Middleware
          </Button>
        </div>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onRetry={fetchMiddlewares}
          onDismiss={clearError}
        />
      )}

      {/* Bulk Actions Bar */}
      {selectedIds.size > 0 && (
        <Card className="border-primary">
          <CardContent className="py-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">
                {selectedIds.size} middleware{selectedIds.size !== 1 ? 's' : ''} selected
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setBulkAssignModalOpen(true)}
                >
                  <Link className="h-4 w-4 mr-2" />
                  Assign to Resources
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => setBulkDeleteModalOpen(true)}
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete Selected
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setSelectedIds(new Set())}
                >
                  Clear Selection
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Middlewares</CardTitle>
              <CardDescription>
                {filteredMiddlewares.length} of {middlewares.length} middlewares
              </CardDescription>
            </div>
            <div className="flex items-center gap-4">
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="Filter by type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  {middlewareTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {MIDDLEWARE_TYPE_LABELS[type as MiddlewareType] || type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <div className="relative w-64">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search middlewares..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredMiddlewares.length === 0 ? (
            <EmptyState
              icon={Layers}
              title="No middlewares found"
              description={
                searchTerm || typeFilter !== 'all'
                  ? 'Try adjusting your search or filter'
                  : 'Create your first middleware to get started'
              }
              action={
                !searchTerm && typeFilter === 'all'
                  ? {
                      label: 'Create Middleware',
                      onClick: () => navigateTo('middleware-form'),
                    }
                  : undefined
              }
            />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <Checkbox
                      checked={isAllSelected}
                      onCheckedChange={handleSelectAll}
                    />
                  </TableHead>
                  <TableHead
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleSort('name')}
                  >
                    <div className="flex items-center">
                      Name {getSortIcon('name')}
                    </div>
                  </TableHead>
                  <TableHead
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleSort('type')}
                  >
                    <div className="flex items-center">
                      Type {getSortIcon('type')}
                    </div>
                  </TableHead>
                  <TableHead
                    className="cursor-pointer hover:bg-muted/50"
                    onClick={() => handleSort('id')}
                  >
                    <div className="flex items-center">
                      ID {getSortIcon('id')}
                    </div>
                  </TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredMiddlewares.map((middleware) => (
                  <TableRow
                    key={middleware.id}
                    className={selectedIds.has(middleware.id) ? 'bg-muted/50' : ''}
                  >
                    <TableCell>
                      <Checkbox
                        checked={selectedIds.has(middleware.id)}
                        onCheckedChange={(checked) =>
                          handleSelectOne(middleware.id, checked as boolean)
                        }
                      />
                    </TableCell>
                    <TableCell className="font-medium">
                      {middleware.name}
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">
                        {MIDDLEWARE_TYPE_LABELS[middleware.type as MiddlewareType] ||
                          middleware.type}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">
                      {middleware.id}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigateTo('middleware-form', middleware.id)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openDeleteModal(middleware)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Delete Confirmation Modal */}
      <ConfirmationModal
        open={deleteModalOpen}
        onOpenChange={setDeleteModalOpen}
        title="Delete Middleware"
        description={`Are you sure you want to delete "${middlewareToDelete?.name}"? This may affect resources using this middleware.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleDelete}
      />

      {/* Bulk Delete Confirmation Modal */}
      <ConfirmationModal
        open={bulkDeleteModalOpen}
        onOpenChange={setBulkDeleteModalOpen}
        title="Delete Selected Middlewares"
        description={`Are you sure you want to delete ${selectedIds.size} middleware${selectedIds.size !== 1 ? 's' : ''}? This may affect resources using these middlewares.`}
        confirmLabel={bulkDeleting ? 'Deleting...' : 'Delete All'}
        variant="destructive"
        onConfirm={handleBulkDelete}
      />

      {/* Bulk Assign Modal */}
      <Dialog open={bulkAssignModalOpen} onOpenChange={setBulkAssignModalOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Assign Middlewares to Resources</DialogTitle>
            <DialogDescription>
              Select the resources to assign the selected {selectedIds.size} middleware{selectedIds.size !== 1 ? 's' : ''} to.
            </DialogDescription>
          </DialogHeader>
          <div className="max-h-64 overflow-y-auto space-y-2 my-4">
            {resources.map((resource) => (
              <label
                key={resource.id}
                className="flex items-center gap-3 p-2 rounded hover:bg-muted cursor-pointer"
              >
                <Checkbox
                  checked={selectedResourceIds.has(resource.id)}
                  onCheckedChange={(checked) => {
                    const newSelected = new Set(selectedResourceIds)
                    if (checked) {
                      newSelected.add(resource.id)
                    } else {
                      newSelected.delete(resource.id)
                    }
                    setSelectedResourceIds(newSelected)
                  }}
                />
                <div>
                  <p className="font-medium">{resource.host}</p>
                  <p className="text-xs text-muted-foreground">{resource.id}</p>
                </div>
              </label>
            ))}
            {resources.length === 0 && (
              <p className="text-sm text-muted-foreground text-center py-4">
                No resources available
              </p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setBulkAssignModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleBulkAssign}
              disabled={selectedResourceIds.size === 0 || bulkAssigning}
            >
              {bulkAssigning ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Assigning...
                </>
              ) : (
                <>
                  <Link className="h-4 w-4 mr-2" />
                  Assign to {selectedResourceIds.size} Resource{selectedResourceIds.size !== 1 ? 's' : ''}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
