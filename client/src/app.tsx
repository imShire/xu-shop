import { PropsWithChildren, useEffect, useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Taro, { useDidShow, useLaunch } from '@tarojs/taro'
import ErrorBoundary from '@/components/ErrorBoundary'
import H5Shell from '@/components/H5Shell'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { ConfigProvider } from '@/ui/nutui'
import { isH5 } from '@/utils/platform'
import { configureUserIdGetter, getReporter, installClogLifecycle, report } from '@/utils/clog'
import { extractTokenFromLaunch, persistTraceId } from '@/utils/trace'
import { trackShareClick } from '@/services/share'
import '@nutui/nutui-react-taro/dist/style.css'
import './app.scss'

type TaroSystemCompat = typeof Taro & {
  getAppBaseInfo?: () => Record<string, unknown>
  getDeviceInfo?: () => Record<string, unknown>
  getSystemInfoSync?: () => Record<string, unknown>
  getWindowInfo?: () => Record<string, unknown>
}

function createH5SystemInfoFallback() {
  const windowWidth = window.innerWidth || document.documentElement.clientWidth || 375
  const windowHeight = window.innerHeight || document.documentElement.clientHeight || 667
  const screenWidth = window.screen?.width || windowWidth
  const screenHeight = window.screen?.height || windowHeight
  const pixelRatio = window.devicePixelRatio || 1
  const language = navigator.language || 'zh-CN'

  return {
    brand: 'web',
    language,
    model: 'H5',
    pixelRatio,
    platform: /iphone|ipad|ipod/i.test(navigator.userAgent) ? 'ios' : 'web',
    safeArea: {
      bottom: windowHeight,
      height: windowHeight,
      left: 0,
      right: windowWidth,
      top: 0,
      width: windowWidth,
    },
    screenHeight,
    screenWidth,
    system: navigator.userAgent,
    windowHeight,
    windowWidth,
  }
}

function ensureTaroSystemApis() {
  if (process.env.TARO_ENV !== 'h5' || typeof window === 'undefined') {
    return
  }

  const taroCompat = Taro as TaroSystemCompat
  const getSystemInfo = () => createH5SystemInfoFallback()
  const getSystemInfoSync = getSystemInfo as unknown as NonNullable<TaroSystemCompat['getSystemInfoSync']>
  const getWindowInfo = getSystemInfo as unknown as NonNullable<TaroSystemCompat['getWindowInfo']>
  const getDeviceInfo = getSystemInfo as unknown as NonNullable<TaroSystemCompat['getDeviceInfo']>
  const getAppBaseInfo = getSystemInfo as unknown as NonNullable<TaroSystemCompat['getAppBaseInfo']>

  if (typeof taroCompat.getSystemInfoSync !== 'function') {
    taroCompat.getSystemInfoSync = getSystemInfoSync
  }

  if (typeof taroCompat.getWindowInfo !== 'function') {
    taroCompat.getWindowInfo = getWindowInfo
  }

  if (typeof taroCompat.getDeviceInfo !== 'function') {
    taroCompat.getDeviceInfo = getDeviceInfo
  }

  if (typeof taroCompat.getAppBaseInfo !== 'function') {
    taroCompat.getAppBaseInfo = getAppBaseInfo
  }
}

ensureTaroSystemApis()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 0,
      staleTime: 30_000,
      refetchOnWindowFocus: false,
    },
  },
})

function App({ children }: PropsWithChildren) {
  const hydrate = useAuthStore((state) => state.hydrate)
  const refreshCount = useCartStore((state) => state.refreshCount)
  const [currentRoute, setCurrentRoute] = useState('pages/home/index')
  const [stackLength, setStackLength] = useState(1)

  useEffect(() => {
    // 把 user_id 注入 clog
    configureUserIdGetter(() => useAuthStore.getState().user?.id ?? undefined)
    installClogLifecycle()

    if (isH5 && typeof window !== 'undefined') {
      const onErr = (
        _msg: Event | string,
        source?: string,
        lineno?: number,
        colno?: number,
        error?: Error
      ) => {
        try {
          const msg = error?.message ?? (typeof _msg === 'string' ? _msg : 'window.onerror')
          report('error', msg, {
            stack: error?.stack,
            extra: { source, lineno, colno },
          })
        } catch {
          /* noop */
        }
      }
      const onRej = (ev: PromiseRejectionEvent) => {
        try {
          const reason = ev?.reason
          const msg = reason instanceof Error ? reason.message : String(reason ?? 'unhandledrejection')
          report('error', `unhandledrejection: ${msg}`, {
            stack: reason instanceof Error ? reason.stack : undefined,
          })
        } catch {
          /* noop */
        }
      }
      window.addEventListener('error', onErr as unknown as EventListener)
      window.addEventListener('unhandledrejection', onRej)
      return () => {
        window.removeEventListener('error', onErr as unknown as EventListener)
        window.removeEventListener('unhandledrejection', onRej)
      }
    }

    // weapp 全局错误
    try {
      Taro.onError?.((err) => {
        const msg = typeof err === 'string' ? err : (err as Error)?.message ?? String(err)
        const stack = typeof err === 'string' ? err : (err as Error)?.stack
        report('error', `weapp_onerror: ${String(msg).slice(0, 200)}`, { stack })
      })
      Taro.onUnhandledRejection?.((res: { reason?: unknown }) => {
        const reason = res?.reason
        const msg = reason instanceof Error ? reason.message : String(reason ?? 'unhandledrejection')
        report('error', `weapp_unhandledrejection: ${msg}`, {
          stack: reason instanceof Error ? reason.stack : undefined,
        })
      })
    } catch {
      /* noop */
    }
    return undefined
  }, [])

  // weapp 应用切后台时刷一次队列
  Taro.useDidHide?.(() => {
    try {
      void getReporter().flush()
    } catch {
      /* noop */
    }
  })

  function syncRouteState() {
    if (isH5 && typeof window !== 'undefined') {
      const route = (window.location.hash.replace(/^#/, '').split('?')[0] || '/pages/home/index')
        .replace(/^\//, '')

      setCurrentRoute(route)
      setStackLength(Math.max(window.history.length, 1))
      return
    }

    const pages = Taro.getCurrentPages()
    const current = pages[pages.length - 1]
    const fallbackPath = Taro.getCurrentInstance().router?.path

    const route = (current?.route ?? fallbackPath ?? '/pages/home/index')
      .replace(/^\//, '')
      .split('?')[0]

    setCurrentRoute(route)
    setStackLength(pages.length || 1)
  }

  useLaunch((options) => {
    const scene = options?.query?.scene
    // ─── share trace ─────────────────────────────────────────────
    const token = extractTokenFromLaunch({
      query: options?.query as Record<string, string> | undefined,
      scene: typeof scene === 'string' ? scene : undefined,
    })
    if (token) {
      persistTraceId(token)
      trackShareClick(token)
    }
    void hydrate(typeof scene === 'string' ? scene : undefined).then(() => refreshCount())
    syncRouteState()
  })

  useDidShow(() => {
    void refreshCount()
    syncRouteState()
  })

  useEffect(() => {
    if (!isH5 || typeof window === 'undefined') {
      return
    }

    const handleRouteChange = () => {
      syncRouteState()
    }

    handleRouteChange()
    window.addEventListener('hashchange', handleRouteChange)

    return () => {
      window.removeEventListener('hashchange', handleRouteChange)
    }
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        theme={{
          primaryColor: '#0F4C36',
          primaryColorEnd: '#0F4C36',
        }}
      >
        <ErrorBoundary>
          <H5Shell currentRoute={currentRoute} stackLength={stackLength}>
            {children}
          </H5Shell>
        </ErrorBoundary>
      </ConfigProvider>
    </QueryClientProvider>
  )
}

export default App
