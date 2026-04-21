export type ProviderKey =
  | 'docker'
  | 'file-external'
  | 'http-provider'
  | 'swarm'
  | 'ecs'
  | 'kubernetes'
  | 'consul'
  | 'consulcatalog'
  | 'etcd'
  | 'redis'
  | 'zookeeper'
  | 'nomad'

export interface NavItem {
  label: string
  to: string
  featureKey?: 'certs' | 'plugins' | 'logs'
}

export interface NavGroup {
  key: string
  label: string
  items: NavItem[]
}

const providerItems: NavItem[] = [
  { label: 'Docker', to: '/providers/docker' },
  { label: 'File (External)', to: '/providers/file-external' },
  { label: 'HTTP Provider', to: '/providers/http-provider' },
  { label: 'Swarm', to: '/providers/swarm' },
  { label: 'ECS', to: '/providers/ecs' },
  { label: 'Kubernetes', to: '/providers/kubernetes' },
  { label: 'Consul', to: '/providers/consul' },
  { label: 'Consul Catalog', to: '/providers/consulcatalog' },
  { label: 'Etcd', to: '/providers/etcd' },
  { label: 'Redis', to: '/providers/redis' },
  { label: 'ZooKeeper', to: '/providers/zookeeper' },
  { label: 'Nomad', to: '/providers/nomad' },
]

const groups: NavGroup[] = [
  {
    key: 'overview',
    label: 'Overview',
    items: [
      { label: 'Dashboard', to: '/' },
      { label: 'Services', to: '/services' },
      { label: 'Route Map', to: '/routemap' },
    ],
  },
  {
    key: 'traffic',
    label: 'Traffic',
    items: [
      { label: 'Routes', to: '/routes' },
      { label: 'Middlewares', to: '/middlewares' },
    ],
  },
  {
    key: 'operations',
    label: 'Operations',
    items: [
      { label: 'Backups', to: '/backups' },
      { label: 'Certificates', to: '/certificates', featureKey: 'certs' },
      { label: 'Plugins', to: '/plugins', featureKey: 'plugins' },
      { label: 'Logs', to: '/logs', featureKey: 'logs' },
    ],
  },
  {
    key: 'configuration',
    label: 'Configuration',
    items: [{ label: 'Settings', to: '/settings' }, ...providerItems],
  },
]

export function buildNavigation(visibleTabs: Record<string, boolean>): NavGroup[] {
  return groups.map((group) => ({
    ...group,
    items: group.items.filter((item) => {
      if (!item.featureKey) return true
      return visibleTabs[item.featureKey] !== false
    }),
  }))
}
