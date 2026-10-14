export interface ChainMiddlewareRef {
  id: string
  name: string
}

// resolveChainMemberId maps a stored chain entry to the middleware ID.
// Entries may be an ID, a display name, or either form plus a provider suffix.
export function resolveChainMemberId(ref: string, middlewares: ChainMiddlewareRef[]): string {
  const trimmed = ref.trim()
  if (!trimmed) {
    return trimmed
  }
  const at = trimmed.lastIndexOf('@')
  const base = at > 0 ? trimmed.slice(0, at) : trimmed

  const byId = middlewares.find((mw) => mw.id === trimmed || mw.id === base)
  if (byId) {
    return byId.id
  }
  const byName = middlewares.find((mw) => mw.name === trimmed || mw.name === base)
  if (byName) {
    return byName.id
  }
  const lower = base.toLowerCase()
  const byLower = middlewares.find(
    (mw) => mw.id.toLowerCase() === lower || mw.name.toLowerCase() === lower
  )
  if (byLower) {
    return byLower.id
  }
  return trimmed
}
