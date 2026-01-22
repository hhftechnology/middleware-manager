import { useState } from 'react'
import { useMTLSStore } from '@/stores/mtlsStore'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2, Plus, Trash2, ShieldCheck, Info } from 'lucide-react'
import type { CreateCARequest } from '@/types'

export function CAManager() {
  const { config, loading, createCA, deleteCA } = useMTLSStore()

  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [formData, setFormData] = useState<CreateCARequest>({
    common_name: '',
    organization: '',
    country: '',
    validity_days: 1825,
  })
  const [formError, setFormError] = useState<string | null>(null)

  const handleCreate = async () => {
    if (!formData.common_name.trim()) {
      setFormError('Common Name is required')
      return
    }

    const success = await createCA(formData)
    if (success) {
      setIsCreateOpen(false)
      setFormData({
        common_name: '',
        organization: '',
        country: '',
        validity_days: 1825,
      })
      setFormError(null)
    }
  }

  const handleDelete = async () => {
    await deleteCA()
  }

  return (
    <div className="space-y-4">
      {config?.has_ca ? (
        // CA Configured - Show Details
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <ShieldCheck className="h-5 w-5 text-green-500" />
                  Certificate Authority
                </CardTitle>
                <CardDescription>
                  Your CA is configured and ready to sign client certificates
                </CardDescription>
              </div>
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" size="sm">
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete CA
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete Certificate Authority?</AlertDialogTitle>
                    <AlertDialogDescription>
                      This will permanently delete the CA and all client certificates.
                      Any devices using these certificates will lose access.
                      This action cannot be undone.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground">
                      Delete CA
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-1">
                <Label className="text-muted-foreground">Subject</Label>
                <p className="font-medium">{config.ca_subject || 'N/A'}</p>
              </div>
              <div className="space-y-1">
                <Label className="text-muted-foreground">Expiry Date</Label>
                <p className="font-medium">
                  {config.ca_expiry
                    ? new Date(config.ca_expiry).toLocaleDateString()
                    : 'N/A'}
                </p>
              </div>
              <div className="space-y-1">
                <Label className="text-muted-foreground">Certificate Path</Label>
                <p className="font-mono text-sm">{config.ca_cert_path || 'N/A'}</p>
              </div>
              <div className="space-y-1">
                <Label className="text-muted-foreground">Client Certificates</Label>
                <p className="font-medium">{config.client_count} issued</p>
              </div>
            </div>

            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                The CA certificate is stored at <code className="text-xs">{config.ca_cert_path}</code>.
                Ensure this path is mounted into your Traefik container.
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      ) : (
        // No CA - Show Create Form
        <Card>
          <CardHeader>
            <CardTitle>Create Certificate Authority</CardTitle>
            <CardDescription>
              Set up a CA to start issuing client certificates for mTLS authentication
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="h-4 w-4 mr-2" />
                  Create CA
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Create Certificate Authority</DialogTitle>
                  <DialogDescription>
                    Configure your CA details. These will appear in the certificate subject.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 py-4">
                  {formError && (
                    <Alert variant="destructive">
                      <AlertDescription>{formError}</AlertDescription>
                    </Alert>
                  )}
                  <div className="space-y-2">
                    <Label htmlFor="common_name">Common Name *</Label>
                    <Input
                      id="common_name"
                      placeholder="e.g., My Company CA"
                      value={formData.common_name}
                      onChange={(e) =>
                        setFormData({ ...formData, common_name: e.target.value })
                      }
                    />
                    <p className="text-xs text-muted-foreground">
                      The primary identifier for your CA
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="organization">Organization</Label>
                    <Input
                      id="organization"
                      placeholder="e.g., My Company Inc."
                      value={formData.organization}
                      onChange={(e) =>
                        setFormData({ ...formData, organization: e.target.value })
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="country">Country Code</Label>
                    <Input
                      id="country"
                      placeholder="e.g., US, DE, UK"
                      maxLength={2}
                      value={formData.country}
                      onChange={(e) =>
                        setFormData({ ...formData, country: e.target.value.toUpperCase() })
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="validity">Validity (days)</Label>
                    <Input
                      id="validity"
                      type="number"
                      min={365}
                      max={3650}
                      value={formData.validity_days}
                      onChange={(e) =>
                        setFormData({ ...formData, validity_days: parseInt(e.target.value) || 1825 })
                      }
                    />
                    <p className="text-xs text-muted-foreground">
                      Default: 1825 days (5 years)
                    </p>
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleCreate} disabled={loading}>
                    {loading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
                    Create CA
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>

            <div className="mt-4 p-4 bg-muted rounded-lg">
              <h4 className="font-medium mb-2">What happens when you create a CA?</h4>
              <ul className="text-sm text-muted-foreground space-y-1">
                <li>A 4096-bit RSA private key is generated</li>
                <li>A self-signed CA certificate is created</li>
                <li>The certificate is saved to the filesystem for Traefik</li>
                <li>You can then issue client certificates for your devices</li>
              </ul>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
