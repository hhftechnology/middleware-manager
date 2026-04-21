import { useState } from 'react'
import type { FormEvent } from 'react'
import type { ConfigFileEntry, RouteApp, RouteRequest } from '@/types'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const emptyForm: RouteRequest = {
  protocol: 'http',
  configFile: '',
  serviceName: '',
  domains: [],
  subdomain: '',
  rule: '',
  target: '',
  targetPort: '',
  scheme: 'http',
  middlewares: [],
  entryPoints: ['websecure'],
  certResolver: '',
  passHostHeader: true,
  insecureSkipVerify: false,
}

function toEditForm(route: RouteApp): RouteRequest {
  return {
    protocol: route.protocol,
    configFile: route.configFile,
    serviceName: route.name,
    domains: [],
    subdomain: '',
    rule: route.rule,
    target: route.protocol === 'http' ? route.target : route.target.split(':')[0] || route.target,
    targetPort: route.protocol === 'http' ? '' : route.target.split(':').slice(1).join(':'),
    scheme: route.target.startsWith('https://') ? 'https' : 'http',
    middlewares: route.middlewares,
    entryPoints: route.entryPoints,
    certResolver: route.certResolver || '',
    passHostHeader: route.passHostHeader ?? true,
    insecureSkipVerify: route.insecureSkipVerify ?? false,
  }
}

interface RouteModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  configs: ConfigFileEntry[]
  editing: RouteApp | null
  onSubmit: (payload: RouteRequest) => Promise<void>
  pending: boolean
}

export function RouteModal({
  open,
  onOpenChange,
  configs,
  editing,
  onSubmit,
  pending,
}: RouteModalProps) {
  const [form, setForm] = useState<RouteRequest>(editing ? toEditForm(editing) : emptyForm)

  const defaultConfig = configs[0]?.label ?? ''
  const effectiveConfig = form.configFile || defaultConfig

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await onSubmit({ ...form, configFile: effectiveConfig })
    setForm(emptyForm)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{editing ? 'Edit Route' : 'Create Route'}</DialogTitle>
          <DialogDescription>
            Configure route targets, entrypoints, middleware links, and protocol settings.
          </DialogDescription>
        </DialogHeader>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="route-serviceName">Service Name</Label>
            <Input
              id="route-serviceName"
              value={form.serviceName}
              onChange={(event) => setForm({ ...form, serviceName: event.target.value })}
              placeholder="whoami"
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>Protocol</Label>
              <Select
                value={form.protocol}
                onValueChange={(value) => setForm({ ...form, protocol: value as RouteRequest['protocol'] })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="http">HTTP</SelectItem>
                  <SelectItem value="tcp">TCP</SelectItem>
                  <SelectItem value="udp">UDP</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Config File</Label>
              <Select
                value={effectiveConfig}
                onValueChange={(value) => setForm({ ...form, configFile: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select..." />
                </SelectTrigger>
                <SelectContent>
                  {configs.map((file) => (
                    <SelectItem key={file.label} value={file.label}>
                      {file.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="route-target">Target</Label>
            <Input
              id="route-target"
              value={form.target}
              onChange={(event) => setForm({ ...form, target: event.target.value })}
              placeholder={form.protocol === 'http' ? 'http://app:8080' : '10.0.0.10'}
            />
          </div>

          {form.protocol !== 'http' ? (
            <div className="space-y-2">
              <Label htmlFor="route-target-port">Target Port</Label>
              <Input
                id="route-target-port"
                value={form.targetPort}
                onChange={(event) => setForm({ ...form, targetPort: event.target.value })}
                placeholder="8080"
              />
            </div>
          ) : null}

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="route-subdomain">Subdomain</Label>
              <Input
                id="route-subdomain"
                value={form.subdomain}
                onChange={(event) => setForm({ ...form, subdomain: event.target.value })}
                placeholder="app"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="route-domains">Domains</Label>
              <Input
                id="route-domains"
                value={form.domains.join(',')}
                onChange={(event) =>
                  setForm({
                    ...form,
                    domains: event.target.value.split(',').map((item) => item.trim()).filter(Boolean),
                  })
                }
                placeholder="example.com,internal.example.com"
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="route-rule">Rule Override</Label>
            <Input
              id="route-rule"
              value={form.rule}
              onChange={(event) => setForm({ ...form, rule: event.target.value })}
              placeholder="Host(`app.example.com`)"
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="route-entrypoints">Entrypoints</Label>
              <Input
                id="route-entrypoints"
                value={form.entryPoints.join(',')}
                onChange={(event) =>
                  setForm({
                    ...form,
                    entryPoints: event.target.value.split(',').map((item) => item.trim()).filter(Boolean),
                  })
                }
                placeholder="websecure"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="route-middlewares">Middlewares</Label>
              <Input
                id="route-middlewares"
                value={form.middlewares.join(',')}
                onChange={(event) =>
                  setForm({
                    ...form,
                    middlewares: event.target.value.split(',').map((item) => item.trim()).filter(Boolean),
                  })
                }
                placeholder="security-headers,compress"
              />
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="route-cert-resolver">Cert Resolver</Label>
              <Input
                id="route-cert-resolver"
                value={form.certResolver}
                onChange={(event) => setForm({ ...form, certResolver: event.target.value })}
                placeholder="cloudflare"
              />
            </div>
            {form.protocol === 'http' ? (
              <div className="space-y-2">
                <Label>Scheme</Label>
                <Select value={form.scheme} onValueChange={(value) => setForm({ ...form, scheme: value })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="http">http</SelectItem>
                    <SelectItem value="https">https</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            ) : (
              <div />
            )}
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <label className="flex items-center gap-3 text-sm">
              <Checkbox
                checked={form.passHostHeader ?? true}
                onCheckedChange={(checked) => setForm({ ...form, passHostHeader: Boolean(checked) })}
              />
              Pass Host Header
            </label>
            <label className="flex items-center gap-3 text-sm">
              <Checkbox
                checked={Boolean(form.insecureSkipVerify)}
                onCheckedChange={(checked) => setForm({ ...form, insecureSkipVerify: Boolean(checked) })}
              />
              Insecure Skip Verify
            </label>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {editing ? 'Save Route' : 'Create Route'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
