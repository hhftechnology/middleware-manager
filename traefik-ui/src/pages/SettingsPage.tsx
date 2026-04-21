import { useState } from 'react'
import type { FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { Settings } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { SettingsModal } from '@/components/modals/SettingsModal'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { useToast } from '@/hooks/use-toast'

function SettingsForm({ initial }: { initial: Settings }) {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [form, setForm] = useState<Settings>(initial)
  const [testResult, setTestResult] = useState<string>('')

  const saveMutation = useMutation({
    mutationFn: (payload: Partial<Settings>) => api.settings.save(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['settings'] })
      toast({ title: 'Settings saved' })
    },
    onError: (error) => {
      toast({
        title: 'Failed to save',
        description: error instanceof Error ? error.message : String(error),
        variant: 'destructive',
      })
    },
  })

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await saveMutation.mutateAsync(form)
  }

  async function testConnection() {
    const result = await api.settings.testConnection(form.traefik_api_url)
    setTestResult(result.ok ? 'Connection successful' : result.error || 'Connection failed')
  }

  return (
    <form className="grid gap-6 xl:grid-cols-2" onSubmit={handleSave}>
      <Card>
        <CardHeader>
          <CardTitle>Core</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="domains">Domains (comma separated)</Label>
            <Input
              id="domains"
              value={form.domains.join(',')}
              onChange={(event) =>
                setForm({
                  ...form,
                  domains: event.target.value.split(',').map((item) => item.trim()).filter(Boolean),
                })
              }
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="cert_resolver">Cert Resolver</Label>
            <Input
              id="cert_resolver"
              value={form.cert_resolver}
              onChange={(event) => setForm({ ...form, cert_resolver: event.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="traefik_api_url">Traefik API URL</Label>
            <div className="flex gap-2">
              <Input
                id="traefik_api_url"
                value={form.traefik_api_url}
                onChange={(event) => setForm({ ...form, traefik_api_url: event.target.value })}
              />
              <Button type="button" variant="outline" onClick={testConnection}>
                Test
              </Button>
            </div>
            {testResult ? (
              <p className="text-xs text-muted-foreground">{testResult}</p>
            ) : null}
          </div>
          <div className="space-y-2">
            <Label htmlFor="acme_json_path">ACME JSON Path</Label>
            <Input
              id="acme_json_path"
              value={form.acme_json_path}
              onChange={(event) => setForm({ ...form, acme_json_path: event.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="access_log_path">Access Log Path</Label>
            <Input
              id="access_log_path"
              value={form.access_log_path}
              onChange={(event) => setForm({ ...form, access_log_path: event.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="static_config_path">Static Config Path</Label>
            <Input
              id="static_config_path"
              value={form.static_config_path}
              onChange={(event) => setForm({ ...form, static_config_path: event.target.value })}
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Self Route & Tabs</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="sr_domain">Self Route Domain</Label>
            <Input
              id="sr_domain"
              value={form.self_route.domain}
              onChange={(event) =>
                setForm({ ...form, self_route: { ...form.self_route, domain: event.target.value } })
              }
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="sr_service">Self Route Service URL</Label>
            <Input
              id="sr_service"
              value={form.self_route.service_url}
              onChange={(event) =>
                setForm({ ...form, self_route: { ...form.self_route, service_url: event.target.value } })
              }
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="sr_router">Self Route Router Name</Label>
            <Input
              id="sr_router"
              value={form.self_route.router_name ?? ''}
              onChange={(event) =>
                setForm({ ...form, self_route: { ...form.self_route, router_name: event.target.value } })
              }
            />
          </div>

          <div className="rounded-md border bg-card p-4">
            <div className="text-sm font-medium">Tab Visibility</div>
            <div className="mt-3 grid gap-3 sm:grid-cols-2">
              {Object.entries(form.visible_tabs).map(([key, value]) => (
                <label key={key} className="flex items-center gap-3 text-sm">
                  <Checkbox
                    checked={Boolean(value)}
                    onCheckedChange={(checked) =>
                      setForm({
                        ...form,
                        visible_tabs: { ...form.visible_tabs, [key]: Boolean(checked) },
                      })
                    }
                  />
                  <span>{key}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="flex justify-end">
            <Button type="submit" disabled={saveMutation.isPending}>
              Save Settings
            </Button>
          </div>
        </CardContent>
      </Card>
    </form>
  )
}

export default function SettingsPage() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const settingsQuery = useQuery({ queryKey: ['settings'], queryFn: api.settings.get })
  const [panelModalOpen, setPanelModalOpen] = useState(false)

  const panelSaveMutation = useMutation({
    mutationFn: (payload: Partial<Settings>) => api.settings.save(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['settings'] })
      toast({ title: 'Panel settings saved' })
    },
    onError: (error) => {
      toast({
        title: 'Failed to save panel settings',
        description: error instanceof Error ? error.message : String(error),
        variant: 'destructive',
      })
    },
  })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Update domains, Traefik API settings, file paths, and the optional self-route."
        actions={
          <Button variant="outline" onClick={() => setPanelModalOpen(true)}>
            Open Settings Panel
          </Button>
        }
      />
      {settingsQuery.data ? (
        <SettingsForm key={settingsQuery.dataUpdatedAt} initial={settingsQuery.data} />
      ) : (
        <Card>
          <CardContent className="pt-6 text-sm text-muted-foreground">Loading settings...</CardContent>
        </Card>
      )}
      {settingsQuery.data ? (
        <SettingsModal
          key={`settings-panel-${settingsQuery.dataUpdatedAt}`}
          open={panelModalOpen}
          onOpenChange={setPanelModalOpen}
          settings={settingsQuery.data}
          pending={panelSaveMutation.isPending}
          onSubmit={async (payload) => {
            await panelSaveMutation.mutateAsync(payload)
          }}
        />
      ) : null}
    </div>
  )
}
