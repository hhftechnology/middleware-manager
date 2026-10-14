import { describe, expect, it } from 'vitest'
import { resolveChainMemberId } from './middlewareChain'

const middlewares = [
  { id: 'ip-allowlist', name: 'IP Allow List' },
  { id: 'rate-limit', name: 'Rate Limit' },
]

describe('resolveChainMemberId', () => {
  it('resolves display names and IDs', () => {
    expect(resolveChainMemberId('IP Allow List', middlewares)).toBe('ip-allowlist')
    expect(resolveChainMemberId('ip-allowlist', middlewares)).toBe('ip-allowlist')
    expect(resolveChainMemberId('rate-limit@file', middlewares)).toBe('rate-limit')
    expect(resolveChainMemberId('Rate Limit', middlewares)).toBe('rate-limit')
  })

  it('leaves external references unchanged', () => {
    expect(resolveChainMemberId('badger@http', middlewares)).toBe('badger@http')
  })
})
