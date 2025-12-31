import { useEffect, useState } from 'react'
import { useMiddlewareStore } from '@/stores/middlewareStore'
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
import { Search, Plus, Edit, Trash2, Layers, RefreshCw } from 'lucide-react'
import { MIDDLEWARE_TYPE_LABELS } from '@/types'
import type { Middleware, MiddlewareType } from '@/types'

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

  const [searchTerm, setSearchTerm] = useState('')
  const [deleteModalOpen, setDeleteModalOpen] = useState(false)
  const [middlewareToDelete, setMiddlewareToDelete] = useState<Middleware | null>(null)

  useEffect(() => {
    fetchMiddlewares()
  }, [fetchMiddlewares])

  const filteredMiddlewares = middlewares.filter(
    (mw) =>
      mw.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      mw.type.toLowerCase().includes(searchTerm.toLowerCase())
  )

  const handleDelete = async () => {
    if (middlewareToDelete) {
      await deleteMiddleware(middlewareToDelete.id)
      setMiddlewareToDelete(null)
    }
  }

  const openDeleteModal = (middleware: Middleware) => {
    setMiddlewareToDelete(middleware)
    setDeleteModalOpen(true)
  }

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

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Middlewares</CardTitle>
              <CardDescription>
                {filteredMiddlewares.length} of {middlewares.length} middlewares
              </CardDescription>
            </div>
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
        </CardHeader>
        <CardContent>
          {filteredMiddlewares.length === 0 ? (
            <EmptyState
              icon={Layers}
              title="No middlewares found"
              description={
                searchTerm
                  ? 'Try adjusting your search terms'
                  : 'Create your first middleware to get started'
              }
              action={
                !searchTerm
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
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>ID</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredMiddlewares.map((middleware) => (
                  <TableRow key={middleware.id}>
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
    </div>
  )
}
