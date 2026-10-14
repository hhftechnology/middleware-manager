export interface ConnectionPayload {
  success?: boolean
  message?: string
  error?: string
}

export interface ConnectionOutcome {
  success: boolean
  message: string
}

// connectionTestOutcome interprets a connection-test JSON body.
// Successful responses historically omitted `success` and only set `message`,
// which the settings dialog treated as a failure.
export function connectionTestOutcome(result: ConnectionPayload | null | undefined): ConnectionOutcome {
  if (!result) {
    return { success: false, message: 'Connection test failed' }
  }

  const success = typeof result.success === 'boolean' ? result.success : !result.error
  const message =
    result.message ||
    result.error ||
    (success ? 'Connection successful' : 'Connection failed')

  return { success, message }
}
