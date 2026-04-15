import { useState } from 'react'
import type { FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import type { MiddlewareEntry, MiddlewareRequest } from '@/types'
import { PageHeader } from '@/components/common/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
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

const starterYAML = 'headers:\n  customResponseHeaders:\n    X-Managed-By: traefik-manager\n'

function emptyForm(): MiddlewareRequest {
  return { name: '', configFile: '', yaml: starterYAML, originalName: '' }
}

export default function MiddlewaresPage() {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const middlewaresQuery = useQuery({ queryKey: ['middlewares'], queryFn: api.middlewares.list })
  const configsQuery = useQuery({ queryKey: ['configs'], queryFn: api.configs.list })
  const [form, setForm] = useState<MiddlewareRequest>(emptyForm)
  const [editing, setEditing] = useState<MiddlewareEntry | null>(null)

  const defaultConfig = configsQuery.data?.files[0]?.label ?? ''
  const effectiveConfigFile = form.configFile || defaultConfig
  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['middlewares'] })

  const saveMutation = useMutation({
    mutationFn: async (payload: MiddlewareRequest) => {
      const final = { ...payload, configFile: payload.configFile || defaultConfig }
      if (editing) await api.middlewares.update(editing.name, final)
      else await api.middlewares.create(final)
    },
    onSuccess: async () => {
      await invalidate()
      setEditing(null)
      setForm(emptyForm())
      toast({ title: editing ? 'Middleware updated' : 'Middleware created' })
    },
    onError: (error) =>
      toast({
        title: 'Failed to save middleware',
        description: error instanceof Error ? error.message : String(error),
        variant: 'destructive',
      }),
  })

  const deleteMutation = useMutation({
    mutationFn: ({ name, configFile }: { name: string; configFile?: string }) =>
      api.middlewares.delete(name, configFile),
    onSuccess: invalidate,
  })

  function loadForEdit(item: MiddlewareEntry) {
    setEditing(item)
    setForm({ name: item.name, configFile: item.configFile, yaml: item.yaml, originalName: item.name })
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await saveMutation.mutateAsync(form)
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Middlewares"
        description="Manage file-backed HTTP middlewares as raw YAML blocks."
      />
      <div className="grid gap-6 xl:grid-cols-[420px_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>{editing ? 'Edit Middleware' : 'Create Middleware'}</CardTitle>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="security-headers"
                />
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
              <div className="space-y-2">
                <Label htmlFor="yaml">Middleware YAML</Label>
                <Textarea
                  id="yaml"
                  rows={16}
                  value={form.yaml}
                  onChange={(e) => setForm({ ...form, yaml: e.target.value })}
                  className="font-mono text-xs"
                />
              </div>
              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={saveMutation.isPending}>
                  {editing ? 'Save Middleware' : 'Create Middleware'}
                </Button>
                {editing ? (
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setEditing(null)
                      setForm(emptyForm())
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
            <CardTitle>Middleware Inventory</CardTitle>
          </CardHeader>
          <CardContent>
            {middlewaresQuery.data?.length ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Config File</TableHead>
                    <TableHead>Preview</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {middlewaresQuery.data.map((item) => (
                    <TableRow key={item.name + item.configFile}>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {item.configFile || 'default config'}
                      </TableCell>
                      <TableCell>
                        <pre className="max-w-xl whitespace-pre-wrap font-mono text-xs">{item.yaml}</pre>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex flex-wrap justify-end gap-2">
                          <Button variant="outline" size="sm" onClick={() => loadForEdit(item)}>
                            Edit
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => deleteMutation.mutate({ name: item.name, configFile: item.configFile })}
                          >
                            Delete
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : middlewaresQuery.isLoading ? (
              <p className="text-sm text-muted-foreground">Loading middlewares...</p>
            ) : (
              <p className="text-sm text-muted-foreground">
                No middlewares found. Create one from the editor panel.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
