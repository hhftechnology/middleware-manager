import { useEffect, useState } from 'react'
import { useServiceStore } from '@/stores/serviceStore'
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
import { Search, Plus, Edit, Trash2, Server, RefreshCw } from 'lucide-react'
import { SERVICE_TYPE_LABELS } from '@/types'
import type { Service, ServiceType } from '@/types'

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
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [serviceToDelete, setServiceToDelete] = useState<Service | null>(null)

  useEffect(() => {
    fetchServices()
  }, [fetchServices])

  const filteredServices = services.filter(
    (svc) =>
      svc.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      svc.type.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const handleDelete = async () => {
    if (serviceToDelete) {
      await deleteService(serviceToDelete.id)
      setServiceToDelete(null)
    }
  }

  const openDeleteModal = (service: Service) => {
    setServiceToDelete(service)
    setDeleteModalOpen(true)
  }

  if (loading && services.length === 0) {
    return <PageLoader message="Loading services..." />
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Services</h1>
          <p className="text-muted-foreground">
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

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Services</CardTitle>
              <CardDescription>
                {filteredServices.length} of {services.length} services
              </CardDescription>
            </div>
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
        </CardHeader>
        <CardContent>
          {filteredServices.length === 0 ? (
            <EmptyState
              icon={Server}
              title="No services found"
              description={
                searchTerm
                  ? 'Try adjusting your search terms'
                  : 'Create your first service to get started'
              }
              action={
                !searchTerm
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
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>ID</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredServices.map((service) => (
                  <TableRow key={service.id}>
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
    </div>
  )
}
