import { useState } from 'react'
import type { FormEvent } from 'react'
import type { ConfigFileEntry, MiddlewareEntry, MiddlewareRequest } from '@/types'
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
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const starterYAML = 'headers:\n  customResponseHeaders:\n    X-Managed-By: traefik-manager\n'

function toForm(entry: MiddlewareEntry): MiddlewareRequest {
  return { name: entry.name, configFile: entry.configFile, yaml: entry.yaml, originalName: entry.name }
}

interface MiddlewareModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  configs: ConfigFileEntry[]
  editing: MiddlewareEntry | null
  onSubmit: (payload: MiddlewareRequest) => Promise<void>
  pending: boolean
}

export function MiddlewareModal({
  open,
  onOpenChange,
  configs,
  editing,
  onSubmit,
  pending,
}: MiddlewareModalProps) {
  const [form, setForm] = useState<MiddlewareRequest>(
    editing ? toForm(editing) : { name: '', configFile: '', yaml: starterYAML, originalName: '' },
  )

  const defaultConfig = configs[0]?.label ?? ''
  const effectiveConfig = form.configFile || defaultConfig

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await onSubmit({ ...form, configFile: effectiveConfig })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{editing ? 'Edit Middleware' : 'Create Middleware'}</DialogTitle>
          <DialogDescription>Manage file-backed middleware YAML definitions.</DialogDescription>
        </DialogHeader>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="middleware-name">Name</Label>
            <Input
              id="middleware-name"
              value={form.name}
              onChange={(event) => setForm({ ...form, name: event.target.value })}
              placeholder="security-headers"
            />
          </div>
          <div className="space-y-2">
            <Label>Config File</Label>
            <Select value={effectiveConfig} onValueChange={(value) => setForm({ ...form, configFile: value })}>
              <SelectTrigger>
                <SelectValue />
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
          <div className="space-y-2">
            <Label htmlFor="middleware-yaml">Middleware YAML</Label>
            <Textarea
              id="middleware-yaml"
              rows={16}
              value={form.yaml}
              onChange={(event) => setForm({ ...form, yaml: event.target.value })}
              className="font-mono text-xs"
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {editing ? 'Save Middleware' : 'Create Middleware'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
