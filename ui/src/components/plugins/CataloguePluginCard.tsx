import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Download, Check, Star, Loader2 } from 'lucide-react'
import type { CataloguePlugin } from '@/types'

interface CataloguePluginCardProps {
  plugin: CataloguePlugin
  onInstall: (moduleName: string, version?: string) => void
  onSelect: (plugin: CataloguePlugin) => void
  installing: boolean
}

export function CataloguePluginCard({
  plugin,
  onInstall,
  onSelect,
  installing,
}: CataloguePluginCardProps) {
  return (
    <Card
      className="cursor-pointer hover:border-primary/50 transition-colors"
      onClick={() => onSelect(plugin)}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            {plugin.iconUrl ? (
              <img
                src={plugin.iconUrl}
                alt=""
                className="h-10 w-10 rounded-lg object-cover"
              />
            ) : (
              <div className="h-10 w-10 rounded-lg bg-muted flex items-center justify-center text-muted-foreground text-lg font-semibold">
                {plugin.displayName.charAt(0).toUpperCase()}
              </div>
            )}
            <div>
              <CardTitle className="text-base line-clamp-1">{plugin.displayName}</CardTitle>
              <CardDescription className="text-xs">{plugin.author}</CardDescription>
            </div>
          </div>
          <Badge variant="secondary" className="text-xs">
            v{plugin.latestVersion.replace(/^v/, '')}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pb-4">
        <p className="text-sm text-muted-foreground line-clamp-2 mb-4 min-h-[2.5rem]">
          {plugin.summary || 'No description available'}
        </p>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <Star className="h-3.5 w-3.5" />
              {plugin.stars}
            </span>
            <Badge variant="outline" className="text-xs capitalize">
              {plugin.type}
            </Badge>
          </div>

          {plugin.isInstalled ? (
            <Badge variant="success" className="flex items-center gap-1">
              <Check className="h-3 w-3" />
              Installed
            </Badge>
          ) : (
            <Button
              size="sm"
              variant="default"
              onClick={(e) => {
                e.stopPropagation()
                onInstall(plugin.import, plugin.latestVersion)
              }}
              disabled={installing}
            >
              {installing ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <>
                  <Download className="h-4 w-4 mr-1" />
                  Install
                </>
              )}
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
