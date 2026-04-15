/* eslint-disable no-restricted-imports */
import { useEffect, type EffectCallback } from 'react'

/**
 * Run a side effect exactly once on mount. Use only for genuine external
 * synchronization (DOM focus, third-party widget lifecycles, browser APIs).
 *
 * For data fetching use React Query. For derived state compute inline.
 * For event responses use the event handler. For resetting state on a prop
 * change use the `key` prop to force remount.
 */
export function useMountEffect(effect: EffectCallback): void {
  /* eslint-disable-next-line no-restricted-syntax */
  useEffect(effect, [])
}
