import { describe, expect, it } from 'vitest'
import { connectionTestOutcome } from './connectionResult'

describe('connectionTestOutcome', () => {
  it('treats a message-only 200 body as success', () => {
    expect(connectionTestOutcome({ message: 'Connection test successful' })).toEqual({
      success: true,
      message: 'Connection test successful',
    })
  })

  it('honors an explicit success flag', () => {
    expect(connectionTestOutcome({ success: true, message: 'ok' })).toEqual({
      success: true,
      message: 'ok',
    })
  })

  it('treats explicit failure and error text as an error', () => {
    expect(connectionTestOutcome({ success: false, error: 'dial tcp: refused' })).toEqual({
      success: false,
      message: 'dial tcp: refused',
    })
    expect(connectionTestOutcome({ error: 'timeout' })).toEqual({
      success: false,
      message: 'timeout',
    })
  })

  it('rejects an empty payload', () => {
    expect(connectionTestOutcome(undefined)).toEqual({
      success: false,
      message: 'Connection test failed',
    })
  })
})
