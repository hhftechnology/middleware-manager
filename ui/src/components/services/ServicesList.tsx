import { useEffect, useState, useMemo } from 'react'
import { useServiceStore } from '@/stores/serviceStore'
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
import { PageLoader } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { EmptyState } from '@/components/common/EmptyState'
import { ConfirmationModal } from '@/components/common/ConfirmationModal'
import {
  Search,
  Plus,
  Edit,
  Trash2,
  Server,
  RefreshCw,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
} from 'lucide-react'
import { SERVICE_TYPE_LABELS } from '@/types'
import type { Service, ServiceType } from '@/types'

type SortField = 'name' | 'type' | 'id'
type SortOrder = 'asc' | 'desc'

export function ServicesList() {
  const { navigateTo } = useAppStore()
  const {
    services,
    loading,
    error,
    fetchServices,
    deleteService,
    clearError,
  } = useServiceStore()

  const [searchTerm, setSearchTerm] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [sortField, setSortField] = useState<SortField>('name')
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc')
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [serviceToDelete, setServiceToDelete] = useState<Service | null>(null)
  const [bulkDeleteModalOpen, setBulkDeleteModalOpen] = useState(false)
  const [bulkDeleting, setBulkDeleting] = useState(false)

  useEffect(() => {
    fetchServices()
  }, [fetchServices])

  // Get unique service types for filter
  const serviceTypes = useMemo(() => {
    const types = new Set(services.map((svc) => svc.type))
    return Array.from(types).sort()
  }, [services])

  // Filter and sort services
  const filteredServices = useMemo(() => {
    let result = services.filter((svc) => {
      const matchesSearch =
        svc.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        svc.type.toLowerCase().includes(searchTerm.toLowerCase()) ||
        svc.id.toLowerCase().includes(searchTerm.toLowerCase())

      const matchesType = typeFilter === 'all' || svc.type === typeFilter

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
  }, [services, searchTerm, typeFilter, sortField, sortOrder])

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
      setSelectedIds(new Set(filteredServices.map((svc) => svc.id)))
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
    if (serviceToDelete) {
      const success = await deleteService(serviceToDelete.id)
      if (success) {
        setServiceToDelete(null)
      } else {
        // Get the error from the store
        const currentError = useServiceStore.getState().error
        toast({
          title: 'Cannot delete service',
          description: currentError || 'This service may be assigned to one or more resources. Remove it from all resources first.',
          variant: 'destructive',
        })
        setServiceToDelete(null)
      }
    }
  }

  const handleBulkDelete = async () => {
    setBulkDeleting(true)
    const idsToDelete = Array.from(selectedIds)
    const failedIds: string[] = []
    for (const id of idsToDelete) {
      const success = await deleteService(id)
      if (!success) {
        failedIds.push(id)
      }
    }
    if (failedIds.length > 0) {
      toast({
        title: 'Some services could not be deleted',
        description: `${failedIds.length} service(s) are assigned to resources and cannot be deleted. Remove them from all resources first.`,
        variant: 'destructive',
      })
    }
    setSelectedIds(new Set())
    setBulkDeleting(false)
    setBulkDeleteModalOpen(false)
  }

  const openDeleteModal = (service: Service) => {
    setServiceToDelete(service)
    setDeleteModalOpen(true)
  }

  const isAllSelected = filteredServices.length > 0 &&
    filteredServices.every((svc) => selectedIds.has(svc.id))

  if (loading && services.length === 0) {
    return <PageLoader message="Loading services..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between pb-2 border-b border-border/60">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Services</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Create and manage Traefik service configurations
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => fetchServices()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button onClick={() => navigateTo('service-form')}>
            <Plus className="h-4 w-4 mr-2" />
            New Service
          </Button>
        </div>
      </div>

      {error && (
        <ErrorMessage
          message={error}
          onRetry={fetchServices}
          onDismiss={clearError}
        />
      )}

      {/* Bulk Actions Bar */}
      {selectedIds.size > 0 && (
        <Card className="border-primary">
          <CardContent className="py-3">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">
                {selectedIds.size} service{selectedIds.size !== 1 ? 's' : ''} selected
              </span>
              <div className="flex gap-2">
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
              <CardTitle>All Services</CardTitle>
              <CardDescription>
                {filteredServices.length} of {services.length} services
              </CardDescription>
            </div>
            <div className="flex items-center gap-4">
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger className="w-40">
                  <SelectValue placeholder="Filter by type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  {serviceTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {SERVICE_TYPE_LABELS[type as ServiceType] || type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <div className="relative w-64">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search services..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredServices.length === 0 ? (
            <EmptyState
              icon={Server}
              title="No services found"
              description={
                searchTerm || typeFilter !== 'all'
                  ? 'Try adjusting your search or filter'
                  : 'Create your first service to get started'
              }
              action={
                !searchTerm && typeFilter === 'all'
                  ? {
                      label: 'Create Service',
                      onClick: () => navigateTo('service-form'),
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
                {filteredServices.map((service) => (
                  <TableRow
                    key={service.id}
                    className={selectedIds.has(service.id) ? 'bg-muted/50' : ''}
                  >
                    <TableCell>
                      <Checkbox
                        checked={selectedIds.has(service.id)}
                        onCheckedChange={(checked) =>
                          handleSelectOne(service.id, checked as boolean)
                        }
                      />
                    </TableCell>
                    <TableCell className="font-medium">
                      {service.name}
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">
                        {SERVICE_TYPE_LABELS[service.type as ServiceType] ||
                          service.type}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">
                      {service.id}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigateTo('service-form', service.id)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openDeleteModal(service)}
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
        title="Delete Service"
        description={`Are you sure you want to delete "${serviceToDelete?.name}"? This may affect resources using this service.`}
        confirmLabel="Delete"
        variant="destructive"
        onConfirm={handleDelete}
      />

      {/* Bulk Delete Confirmation Modal */}
      <ConfirmationModal
        open={bulkDeleteModalOpen}
        onOpenChange={setBulkDeleteModalOpen}
        title="Delete Selected Services"
        description={`Are you sure you want to delete ${selectedIds.size} service${selectedIds.size !== 1 ? 's' : ''}? This may affect resources using these services.`}
        confirmLabel={bulkDeleting ? 'Deleting...' : 'Delete All'}
        variant="destructive"
        onConfirm={handleBulkDelete}
      />
    </div>
  )
}
