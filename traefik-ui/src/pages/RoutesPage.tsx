import { useState } from 'react'
import type { FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { RouteApp, RouteRequest } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useToast } from '@/hooks/use-toast'

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

export default function RoutesPage() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const routesQuery = useQuery({ queryKey: ['routes'], queryFn: api.routes.list })
  const configsQuery = useQuery({ queryKey: ['configs'], queryFn: api.configs.list })
  const [form, setForm] = useState<RouteRequest>(emptyForm)
  const [editing, setEditing] = useState<RouteApp | null>(null)

  const defaultConfig = configsQuery.data?.files[0]?.label ?? ''
  const effectiveConfigFile = form.configFile || defaultConfig

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['routes'] })

  const saveMutation = useMutation({
    mutationFn: async (payload: RouteRequest) => {
      const finalPayload = { ...payload, configFile: payload.configFile || defaultConfig }
      if (editing) await api.routes.update(editing.id, finalPayload)
      else await api.routes.create(finalPayload)
    },
    onSuccess: async () => {
      await invalidate()
      setEditing(null)
      setForm({ ...emptyForm })
      toast({ title: editing ? 'Route updated' : 'Route created' })
    },
    onError: (error) =>
      toast({
        title: 'Failed to save route',
        description: error instanceof Error ? error.message : String(error),
        variant: 'destructive',
      }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.routes.delete(id),
    onSuccess: invalidate,
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enable }: { id: string; enable: boolean }) => api.routes.toggle(id, enable),
    onSuccess: invalidate,
  })

  function loadForEdit(route: RouteApp) {
    setEditing(route)
    setForm({
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
    })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await saveMutation.mutateAsync(form)
  }

  return (
    <div className="space-y-6">
      <PageHeader title="Routes" description="Create, edit, disable, and remove file-backed Traefik routes." />
      <div className="grid gap-6 xl:grid-cols-[420px_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>{editing ? 'Edit Route' : 'Create Route'}</CardTitle>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <Label htmlFor="serviceName">Service Name</Label>
                <Input
                  id="serviceName"
                  value={form.serviceName}
                  onChange={(e) => setForm({ ...form, serviceName: e.target.value })}
                  placeholder="whoami"
                />
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>Protocol</Label>
                  <Select
                    value={form.protocol}
                    onValueChange={(v) => setForm({ ...form, protocol: v as RouteRequest['protocol'] })}
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
                    value={effectiveConfigFile}
                    onValueChange={(v) => setForm({ ...form, configFile: v })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select..." />
                    </SelectTrigger>
                    <SelectContent>
                      {configsQuery.data?.files.map((file) => (
                        <SelectItem key={file.label} value={file.label}>
                          {file.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="target">Target</Label>
                <Input
                  id="target"
                  value={form.target}
                  onChange={(e) => setForm({ ...form, target: e.target.value })}
                  placeholder={form.protocol === 'http' ? 'http://app:8080' : '10.0.0.10'}
                />
              </div>

              {form.protocol !== 'http' ? (
                <div className="space-y-2">
                  <Label htmlFor="targetPort">Target Port</Label>
                  <Input
                    id="targetPort"
                    value={form.targetPort}
                    onChange={(e) => setForm({ ...form, targetPort: e.target.value })}
                    placeholder="8080"
                  />
                </div>
              ) : null}

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="subdomain">Subdomain</Label>
                  <Input
                    id="subdomain"
                    value={form.subdomain}
                    onChange={(e) => setForm({ ...form, subdomain: e.target.value })}
                    placeholder="app"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="domains">Domains</Label>
                  <Input
                    id="domains"
                    value={form.domains.join(',')}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        domains: e.target.value.split(',').map((i) => i.trim()).filter(Boolean),
                      })
                    }
                    placeholder="example.com,internal.example.com"
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label htmlFor="rule">Rule Override</Label>
                <Input
                  id="rule"
                  value={form.rule}
                  onChange={(e) => setForm({ ...form, rule: e.target.value })}
                  placeholder="Host(`app.example.com`)"
                />
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="entryPoints">Entrypoints</Label>
                  <Input
                    id="entryPoints"
                    value={form.entryPoints.join(',')}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        entryPoints: e.target.value.split(',').map((i) => i.trim()).filter(Boolean),
                      })
                    }
                    placeholder="websecure"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="middlewares">Middlewares</Label>
                  <Input
                    id="middlewares"
                    value={form.middlewares.join(',')}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        middlewares: e.target.value.split(',').map((i) => i.trim()).filter(Boolean),
                      })
                    }
                    placeholder="security-headers,compress"
                  />
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="certResolver">Cert Resolver</Label>
                  <Input
                    id="certResolver"
                    value={form.certResolver}
                    onChange={(e) => setForm({ ...form, certResolver: e.target.value })}
                    placeholder="cloudflare"
                  />
                </div>
                {form.protocol === 'http' ? (
                  <div className="space-y-2">
                    <Label>Scheme</Label>
                    <Select
                      value={form.scheme}
                      onValueChange={(v) => setForm({ ...form, scheme: v })}
                    >
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

              <div className="grid gap-3 md:grid-cols-2">
                <label className="flex items-center gap-3 text-sm">
                  <Checkbox
                    checked={form.passHostHeader ?? true}
                    onCheckedChange={(c) => setForm({ ...form, passHostHeader: Boolean(c) })}
                  />
                  Pass Host Header
                </label>
                <label className="flex items-center gap-3 text-sm">
                  <Checkbox
                    checked={Boolean(form.insecureSkipVerify)}
                    onCheckedChange={(c) => setForm({ ...form, insecureSkipVerify: Boolean(c) })}
                  />
                  Insecure Skip Verify
                </label>
              </div>

              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={saveMutation.isPending}>
                  {editing ? 'Save Route' : 'Create Route'}
                </Button>
                {editing ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setEditing(null)
                      setForm({ ...emptyForm })
                    }}
                  >
                    Cancel
                  </Button>
                ) : null}
              </div>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Route Inventory</CardTitle>
          </CardHeader>
          <CardContent>
            {routesQuery.data?.apps.length ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Protocol</TableHead>
                    <TableHead>Target</TableHead>
                    <TableHead>Rule</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {routesQuery.data.apps.map((route) => (
                    <TableRow key={route.id}>
                      <TableCell>
                        <div className="font-medium">{route.name}</div>
                        <div className="mt-1 text-xs text-muted-foreground">
                          {route.configFile || 'default config'}
                        </div>
                      </TableCell>
                      <TableCell className="uppercase">{route.protocol}</TableCell>
                      <TableCell className="font-mono text-xs">{route.target}</TableCell>
                      <TableCell className="font-mono text-xs">{route.rule || 'n/a'}</TableCell>
                      <TableCell className="text-right">
                        <div className="flex flex-wrap justify-end gap-2">
                          <Button variant="outline" size="sm" onClick={() => loadForEdit(route)}>
                            Edit
                          </Button>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => toggleMutation.mutate({ id: route.id, enable: !route.enabled })}
                          >
                            {route.enabled ? 'Disable' : 'Enable'}
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => deleteMutation.mutate(route.id)}
                          >
                            Delete
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : routesQuery.isLoading ? (
              <p className="text-sm text-muted-foreground">Loading routes...</p>
            ) : (
              <p className="text-sm text-muted-foreground">
                No routes found. Create one from the panel on the left.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
