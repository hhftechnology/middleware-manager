import { useState } from 'react'
import type { FormEvent } from 'react'
import type { Settings } from '@/types'
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

interface SettingsModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  settings: Settings
  pending: boolean
  onSubmit: (payload: Partial<Settings>) => Promise<void>
}

export function SettingsModal({
  open,
  onOpenChange,
  settings,
  pending,
  onSubmit,
}: SettingsModalProps) {
  const [draft, setDraft] = useState({
    domain: settings.self_route.domain,
    serviceURL: settings.self_route.service_url,
    routerName: settings.self_route.router_name ?? '',
    tabs: settings.visible_tabs,
  })

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await onSubmit({
      self_route: {
        domain: draft.domain,
        service_url: draft.serviceURL,
        router_name: draft.routerName,
      },
      visible_tabs: draft.tabs,
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Settings Panels</DialogTitle>
          <DialogDescription>Update self-route metadata and feature visibility in a modal panel.</DialogDescription>
        </DialogHeader>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="settings-domain">Self Route Domain</Label>
            <Input
              id="settings-domain"
              value={draft.domain}
              onChange={(event) => setDraft({ ...draft, domain: event.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="settings-service-url">Self Route Service URL</Label>
            <Input
              id="settings-service-url"
              value={draft.serviceURL}
              onChange={(event) => setDraft({ ...draft, serviceURL: event.target.value })}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="settings-router-name">Self Route Router Name</Label>
            <Input
              id="settings-router-name"
              value={draft.routerName}
              onChange={(event) => setDraft({ ...draft, routerName: event.target.value })}
            />
          </div>
          <div className="rounded-md border bg-card p-4">
            <div className="text-sm font-medium">Tab Visibility</div>
            <div className="mt-3 grid gap-3 sm:grid-cols-2">
              {Object.entries(draft.tabs).map(([key, value]) => (
                <label key={key} className="flex items-center gap-3 text-sm">
                  <Checkbox
                    checked={Boolean(value)}
                    onCheckedChange={(checked) =>
                      setDraft({ ...draft, tabs: { ...draft.tabs, [key]: Boolean(checked) } })
                    }
                  />
                  {key}
                </label>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              Save Panel Settings
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
