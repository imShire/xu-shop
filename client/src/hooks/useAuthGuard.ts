import { useCallback, useRef } from 'react'
import Taro from '@tarojs/taro'
import { useAuthStore } from '@/stores/auth'

function buildLoginUrl(redirect?: string) {
  // On WeApp, go to the phone login page (which also has WeChat quick-login tab).
  // On H5, the same login page handles both modes.
  const base = '/pages/auth/login/index'
  if (!redirect) {
    return base
  }
  return `${base}?redirect=${encodeURIComponent(redirect)}`
}

export function useAuthGuard() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const isHydrating = useAuthStore((state) => state.isHydrating)
  const hydrate = useAuthStore((state) => state.hydrate)
  const checkingRef = useRef(false)

  return useCallback(async (onPass?: () => void, redirect?: string) => {
    if (isLoggedIn) {
      if (checkingRef.current || isHydrating) {
        return false
      }

      checkingRef.current = true
      try {
        await hydrate()
        if (useAuthStore.getState().isLoggedIn) {
          onPass?.()
          return true
        }
      } finally {
        checkingRef.current = false
      }
    } else {
      checkingRef.current = false
    }

    await Taro.navigateTo({ url: buildLoginUrl(redirect) })
    return false
  }, [hydrate, isHydrating, isLoggedIn])
}
