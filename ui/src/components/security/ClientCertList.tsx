import { useEffect, useState } from 'react'
import { useMTLSStore } from '@/stores/mtlsStore'
import { mtlsApi } from '@/services/api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Loader2,
  Plus,
  Download,
  Ban,
  Trash2,
  User,
  AlertTriangle,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import type { CreateClientRequest, MTLSClient } from '@/types'

export function ClientCertList() {
  const {
    config,
    clients,
    loadingClients,
    fetchClients,
    createClient,
    revokeClient,
    deleteClient,
  } = useMTLSStore()

  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [formData, setFormData] = useState<CreateClientRequest>({
    name: '',
    validity_days: 730,
    p12_password: '',
  })
  const [confirmPassword, setConfirmPassword] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [creating, setCreating] = useState(false)
  const [newClientId, setNewClientId] = useState<string | null>(null)

  useEffect(() => {
    if (config?.has_ca) {
      fetchClients()
    }
  }, [config?.has_ca, fetchClients])

  const handleCreate = async () => {
    setFormError(null)

    if (!formData.name.trim()) {
      setFormError('Client name is required')
      return
    }

    if (!formData.p12_password) {
      setFormError('Password is required to protect the certificate')
      return
    }

    if (formData.p12_password.length < 4) {
      setFormError('Password must be at least 4 characters')
      return
    }

    if (formData.p12_password !== confirmPassword) {
      setFormError('Passwords do not match')
      return
    }

    setCreating(true)
    const client = await createClient(formData)
    setCreating(false)

    if (client) {
      setNewClientId(client.id)
      setIsCreateOpen(false)
      setFormData({ name: '', validity_days: 730, p12_password: '' })
      setConfirmPassword('')
    }
  }

  const handleDownload = (client: MTLSClient) => {
    const url = mtlsApi.getClientP12Url(client.id)
    window.open(url, '_blank')
  }

  const handleRevoke = async (id: string) => {
    await revokeClient(id)
  }

  const handleDelete = async (id: string) => {
    await deleteClient(id)
    if (newClientId === id) {
      setNewClientId(null)
    }
  }

  if (!config?.has_ca) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Client Certificates</CardTitle>
          <CardDescription>
            Create a Certificate Authority first to issue client certificates
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert>
            <AlertTriangle className="h-4 w-4" />
            <AlertDescription>
              Go to the Certificate Authority tab to create a CA before issuing client certificates.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <User className="h-5 w-5" />
              Client Certificates
            </CardTitle>
            <CardDescription>
              Issue and manage client certificates for device authentication
            </CardDescription>
          </div>
          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                New Client
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Client Certificate</DialogTitle>
                <DialogDescription>
                  Generate a new certificate for a device or user.
                  Remember to save the password - you'll need it to import the certificate.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                {formError && (
                  <Alert variant="destructive">
                    <AlertDescription>{formError}</AlertDescription>
                  </Alert>
                )}
                <div className="space-y-2">
                  <Label htmlFor="client_name">Client Name *</Label>
                  <Input
                    id="client_name"
                    placeholder="e.g., john-laptop, admin-phone"
                    value={formData.name}
                    onChange={(e) =>
                      setFormData({ ...formData, name: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-') })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Use lowercase letters, numbers, and hyphens
                  </p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="p12_password">Certificate Password *</Label>
                  <Input
                    id="p12_password"
                    type="password"
                    placeholder="Enter a strong password"
                    value={formData.p12_password}
                    onChange={(e) =>
                      setFormData({ ...formData, p12_password: e.target.value })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="confirm_password">Confirm Password *</Label>
                  <Input
                    id="confirm_password"
                    type="password"
                    placeholder="Confirm the password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="validity_days">Validity (days)</Label>
                  <Input
                    id="validity_days"
                    type="number"
                    min={30}
                    max={1825}
                    value={formData.validity_days}
                    onChange={(e) =>
                      setFormData({ ...formData, validity_days: parseInt(e.target.value) || 730 })
                    }
                  />
                  <p className="text-xs text-muted-foreground">
                    Default: 730 days (2 years)
                  </p>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                  Cancel
                </Button>
                <Button onClick={handleCreate} disabled={creating}>
                  {creating && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                  Create Certificate
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </CardHeader>
      <CardContent>
        {newClientId && (
          <Alert className="mb-4">
            <CheckCircle className="h-4 w-4" />
            <AlertDescription className="flex items-center justify-between">
              <span>Certificate created! Download it now and import it to your device.</span>
              <Button
                size="sm"
                variant="outline"
                onClick={() => {
                  const client = clients.find((c) => c.id === newClientId)
                  if (client) handleDownload(client)
                }}
              >
                <Download className="h-4 w-4 mr-2" />
                Download .p12
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {loadingClients ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        ) : clients.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <User className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>No client certificates yet</p>
            <p className="text-sm">Create your first client certificate to get started</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Subject</TableHead>
                <TableHead>Expiry</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {clients.map((client) => (
                <TableRow key={client.id}>
                  <TableCell className="font-medium">{client.name}</TableCell>
                  <TableCell className="text-sm text-muted-foreground max-w-[200px] truncate">
                    {client.subject}
                  </TableCell>
                  <TableCell>
                    {client.expiry
                      ? new Date(client.expiry).toLocaleDateString()
                      : 'N/A'}
                  </TableCell>
                  <TableCell>
                    {client.revoked ? (
                      <Badge variant="destructive" className="flex items-center gap-1 w-fit">
                        <XCircle className="h-3 w-3" />
                        Revoked
                      </Badge>
                    ) : (
                      <Badge variant="default" className="flex items-center gap-1 w-fit">
                        <CheckCircle className="h-3 w-3" />
                        Active
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDownload(client)}
                      disabled={client.revoked}
                    >
                      <Download className="h-4 w-4" />
                    </Button>
                    {!client.revoked && (
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="outline" size="sm">
                            <Ban className="h-4 w-4" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Revoke Certificate?</AlertDialogTitle>
                            <AlertDialogDescription>
                              This will mark the certificate as revoked. The device will lose access
                              once you generate a new CRL (if using CRL-based revocation).
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction onClick={() => handleRevoke(client.id)}>
                              Revoke
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    )}
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button variant="outline" size="sm">
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete Certificate?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This will permanently delete this client certificate.
                            This action cannot be undone.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            onClick={() => handleDelete(client.id)}
                            className="bg-destructive text-destructive-foreground"
                          >
                            Delete
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  )
}
