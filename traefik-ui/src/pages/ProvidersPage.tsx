import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { PageHeader } from '@/components/common/PageHeader'
import type { ProviderConfigDraft, ProviderFieldSchema } from '@/types'
import { useToast } from '@/hooks/use-toast'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'

interface ProviderSchema {
  key: string
  title: string
  description: string
  fields: ProviderFieldSchema[]
}

const providerSchemas: ProviderSchema[] = [
  {
    key: 'docker',
    title: 'Docker',
    description: 'Configure Docker endpoint and watch settings.',
    fields: [
      { key: 'endpoint', label: 'Endpoint', placeholder: 'unix:///var/run/docker.sock', required: true },
      { key: 'network', label: 'Default Network', placeholder: 'traefik-public' },
    ],
  },
  {
    key: 'file-external',
    title: 'File (External)',
    description: 'Manage external file provider directories and polling.',
    fields: [
      { key: 'directory', label: 'Directory', placeholder: '/etc/traefik/dynamic', required: true },
      { key: 'watch', label: 'Watch Interval', placeholder: '5s' },
    ],
  },
  {
    key: 'http-provider',
    title: 'HTTP Provider',
    description: 'Configure remote HTTP provider URL and timeout.',
    fields: [
      { key: 'endpoint', label: 'Endpoint URL', placeholder: 'https://provider.example.com/config', required: true },
      { key: 'pollInterval', label: 'Poll Interval', placeholder: '15s' },
    ],
  },
  {
    key: 'swarm',
    title: 'Swarm',
    description: 'Configure Docker Swarm manager endpoint and constraints.',
    fields: [
      { key: 'endpoint', label: 'Swarm Endpoint', placeholder: 'tcp://swarm-manager:2377', required: true },
      { key: 'constraints', label: 'Constraints', placeholder: 'Label(`traefik.enable`,`true`)' },
    ],
  },
  {
    key: 'ecs',
    title: 'ECS',
    description: 'Configure ECS region and cluster discovery settings.',
    fields: [
      { key: 'region', label: 'AWS Region', placeholder: 'us-east-1', required: true },
      { key: 'cluster', label: 'Cluster Name', placeholder: 'production-cluster', required: true },
    ],
  },
  {
    key: 'kubernetes',
    title: 'Kubernetes',
    description: 'Configure Kubernetes API endpoint and namespaces.',
    fields: [
      { key: 'apiServer', label: 'API Server', placeholder: 'https://kubernetes.default.svc', required: true },
      { key: 'namespaces', label: 'Namespaces', placeholder: 'default,ingress' },
    ],
  },
  {
    key: 'consul',
    title: 'Consul',
    description: 'Configure Consul catalog endpoint and datacenter.',
    fields: [
      { key: 'endpoint', label: 'Consul Endpoint', placeholder: 'http://consul.service.consul:8500', required: true },
      { key: 'datacenter', label: 'Datacenter', placeholder: 'dc1' },
    ],
  },
  {
    key: 'consulcatalog',
    title: 'Consul Catalog',
    description: 'Configure service catalog integration and tags.',
    fields: [
      { key: 'endpoint', label: 'Catalog Endpoint', placeholder: 'http://consul.service.consul:8500', required: true },
      { key: 'prefix', label: 'Prefix', placeholder: 'traefik' },
    ],
  },
  {
    key: 'etcd',
    title: 'Etcd',
    description: 'Configure Etcd endpoint and root key.',
    fields: [
      { key: 'endpoint', label: 'Etcd Endpoint', placeholder: 'http://etcd:2379', required: true },
      { key: 'rootKey', label: 'Root Key', placeholder: '/traefik' },
    ],
  },
  {
    key: 'redis',
    title: 'Redis',
    description: 'Configure Redis endpoint and key prefix.',
    fields: [
      { key: 'endpoint', label: 'Redis Endpoint', placeholder: 'redis://redis:6379', required: true },
      { key: 'rootKey', label: 'Root Key', placeholder: 'traefik' },
    ],
  },
  {
    key: 'zookeeper',
    title: 'ZooKeeper',
    description: 'Configure ZooKeeper quorum and path.',
    fields: [
      { key: 'endpoints', label: 'Endpoints', placeholder: 'zk1:2181,zk2:2181', required: true },
      { key: 'path', label: 'Path', placeholder: '/traefik' },
    ],
  },
  {
    key: 'nomad',
    title: 'Nomad',
    description: 'Configure Nomad address and namespace.',
    fields: [
      { key: 'endpoint', label: 'Nomad Endpoint', placeholder: 'http://nomad.service.consul:4646', required: true },
      { key: 'namespace', label: 'Namespace', placeholder: 'default' },
    ],
  },
]

function buildInitialDrafts(): Record<string, ProviderConfigDraft> {
  return providerSchemas.reduce<Record<string, ProviderConfigDraft>>((acc, schema) => {
    acc[schema.key] = {
      enabled: false,
      fields: schema.fields.reduce<Record<string, string>>((fieldMap, field) => {
        fieldMap[field.key] = ''
        return fieldMap
      }, {}),
    }
    return acc
  }, {})
}

export default function ProvidersPage() {
  const { toast } = useToast()
  const navigate = useNavigate()
  const params = useParams()
  const [drafts, setDrafts] = useState<Record<string, ProviderConfigDraft>>(buildInitialDrafts)
  const fallbackProvider = providerSchemas[0]?.key ?? 'docker'
  const activeTab = providerSchemas.some((provider) => provider.key === params.provider)
    ? (params.provider as string)
    : fallbackProvider

  function validate(schema: ProviderSchema, draft: ProviderConfigDraft): string[] {
    return schema.fields
      .filter((field) => field.required && !(draft.fields[field.key] ?? '').trim())
      .map((field) => field.label)
  }

  function saveProvider(schema: ProviderSchema) {
    const draft = drafts[schema.key] ?? { enabled: false, fields: {} }
    const missing = validate(schema, draft)
    if (missing.length) {
      toast({
        title: `${schema.title} validation failed`,
        description: `Missing required fields: ${missing.join(', ')}`,
        variant: 'destructive',
      })
      return
    }

    toast({
      title: `${schema.title} configuration prepared`,
      description: 'Provider-specific backend endpoint is not available in this build. Configuration is staged in UI only.',
    })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Providers"
        description="Provider-specific configuration panels with validation and staged save flow."
      />
      <Tabs value={activeTab} onValueChange={(value) => navigate(`/providers/${value}`)} className="space-y-4">
        <TabsList className="h-auto w-full flex-wrap justify-start gap-2 bg-transparent p-0">
          {providerSchemas.map((provider) => (
            <TabsTrigger
              key={provider.key}
              value={provider.key}
              className="rounded-md border bg-card data-[state=active]:bg-accent"
            >
              {provider.title}
            </TabsTrigger>
          ))}
        </TabsList>
        {providerSchemas.map((provider) => {
          const draft = drafts[provider.key] ?? { enabled: false, fields: {} }
          return (
            <TabsContent key={provider.key} value={provider.key} className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>{provider.title}</CardTitle>
                  <CardDescription>{provider.description}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-5">
                  <div className="flex items-center justify-between rounded-md border bg-card p-3">
                    <div>
                      <p className="text-sm font-medium">Enable Provider</p>
                      <p className="text-xs text-muted-foreground">Toggle this provider's configuration state.</p>
                    </div>
                    <Switch
                      checked={draft.enabled}
                      onCheckedChange={(checked) =>
                        setDrafts((current) => ({
                          ...current,
                          [provider.key]: { ...draft, enabled: Boolean(checked) },
                        }))
                      }
                    />
                  </div>
                  <div className="grid gap-4 md:grid-cols-2">
                    {provider.fields.map((field) => (
                      <div key={field.key} className="space-y-2">
                        <Label htmlFor={`${provider.key}-${field.key}`}>
                          {field.label}
                          {field.required ? ' *' : ''}
                        </Label>
                        <Input
                          id={`${provider.key}-${field.key}`}
                          placeholder={field.placeholder}
                          value={draft.fields[field.key] ?? ''}
                          onChange={(event) =>
                            setDrafts((current) => ({
                              ...current,
                              [provider.key]: {
                                ...draft,
                                fields: { ...draft.fields, [field.key]: event.target.value },
                              },
                            }))
                          }
                        />
                      </div>
                    ))}
                  </div>
                  <div className="flex justify-end">
                    <Button onClick={() => saveProvider(provider)}>Save {provider.title}</Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          )
        })}
      </Tabs>
    </div>
  )
}
