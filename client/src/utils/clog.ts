// 前端日志上报工具（client）
// 对应后端 D1：POST /api/v1/internal/clog/batch
//   - 平台：H5 走 window.fetch + sendBeacon；weapp 走 Taro.request
//   - 缓冲：满 20 条或 5 秒触发；失败重试 1 次后丢弃
//   - 长度截断与字段上限同后端
//   - 内部 try/catch，绝不影响主流程
import Taro from '@tarojs/taro'

export type ClogLevel = 'error' | 'warn' | 'info'
export type ClogSource = 'client_h5' | 'client_weapp'

export interface ClogEntry {
  source: ClogSource
  level: ClogLevel
  message: string
  stack?: string
  url?: string
  user_agent?: string
  release?: string
  user_id?: string
  extra?: Record<string, unknown>
}

export interface ReportOptions {
  stack?: string
  extra?: Record<string, unknown>
}

const MSG_MAX = 4 * 1024
const STACK_MAX = 16 * 1024
const EXTRA_MAX = 8 * 1024
const URL_MAX = 512
const UA_MAX = 255
const RELEASE_MAX = 64
const BATCH_SIZE = 20
const FLUSH_INTERVAL_MS = 5000
const SENSITIVE_RE = /("?)(password|token|secret|authorization|csrf)("?\s*[:=]\s*"?)([^",}\s]+)/gi

const isH5Env = process.env.TARO_ENV === 'h5'
const SOURCE: ClogSource = isH5Env ? 'client_h5' : 'client_weapp'

declare const __APP_VERSION__: string

function safeRelease(): string {
  try {
    if (typeof process !== 'undefined' && process.env && process.env.TARO_APP_VERSION) {
      return String(process.env.TARO_APP_VERSION)
    }
  } catch {
    /* noop */
  }
  try {
    if (typeof __APP_VERSION__ === 'string') return __APP_VERSION__
  } catch {
    /* noop */
  }
  return 'dev'
}

function truncate(s: string, max: number): string {
  if (!s) return s
  if (s.length <= max) return s
  return s.slice(0, max) + '...[truncated]'
}

function sanitize(s: string): string {
  if (!s) return s
  try {
    return s.replace(SENSITIVE_RE, '$1$2$3***')
  } catch {
    return s
  }
}

function sanitizeExtra(extra?: Record<string, unknown>): Record<string, unknown> | undefined {
  if (!extra) return undefined
  try {
    let json = JSON.stringify(extra)
    json = sanitize(json)
    if (json.length > EXTRA_MAX) {
      json = json.slice(0, EXTRA_MAX)
      return { _truncated: true, _raw: json }
    }
    return JSON.parse(json) as Record<string, unknown>
  } catch {
    return undefined
  }
}

function currentUrl(): string | undefined {
  try {
    if (isH5Env && typeof window !== 'undefined' && window.location) {
      return window.location.href
    }
    const router = Taro.getCurrentInstance?.()?.router
    return router?.path
  } catch {
    return undefined
  }
}

function currentUserAgent(): string | undefined {
  try {
    if (isH5Env && typeof navigator !== 'undefined') {
      return navigator.userAgent
    }
    const info = Taro.getSystemInfoSync?.() as
      | { brand?: string; model?: string; system?: string; platform?: string; version?: string }
      | undefined
    if (!info) return undefined
    const parts = [info.brand, info.model, info.system, info.platform, info.version].filter(Boolean)
    return parts.join(' ')
  } catch {
    return undefined
  }
}

function resolveEndpoint(custom?: string): string {
  if (custom) return custom
  const base = process.env.TARO_APP_API_BASE || ''
  // base 通常形如 "/api/v1" 或 "http://localhost:8080/api/v1"
  if (base) return `${base}/internal/clog/batch`
  return '/api/v1/internal/clog/batch'
}

export interface ReporterOptions {
  endpoint?: string
  getUserId?: () => string | undefined
  /** 测试注入：fetch 实现 */
  fetchImpl?: typeof fetch
  /** 测试注入：sendBeacon 实现 */
  beaconImpl?: (url: string, data: BodyInit) => boolean
  /** 测试注入：Taro.request 实现 */
  requestImpl?: (opts: { url: string; method: 'POST'; data: string; header: Record<string, string> }) => Promise<unknown>
  /** 测试注入：强制平台 */
  platform?: ClogSource
}

export class ClogReporter {
  private queue: ClogEntry[] = []
  private timer: ReturnType<typeof setTimeout> | null = null
  private endpoint: string
  private source: ClogSource
  private getUserId: () => string | undefined
  private fetchImpl?: typeof fetch
  private beaconImpl?: (url: string, data: BodyInit) => boolean
  private requestImpl?: (opts: { url: string; method: 'POST'; data: string; header: Record<string, string> }) => Promise<unknown>

  constructor(opts: ReporterOptions = {}) {
    this.source = opts.platform ?? SOURCE
    this.endpoint = resolveEndpoint(opts.endpoint)
    this.getUserId = opts.getUserId ?? (() => undefined)
    this.fetchImpl = opts.fetchImpl
    this.beaconImpl = opts.beaconImpl
    this.requestImpl = opts.requestImpl
  }

  report(level: ClogLevel, message: string, opts: ReportOptions = {}): void {
    try {
      const entry: ClogEntry = {
        source: this.source,
        level,
        message: truncate(sanitize(String(message ?? '')), MSG_MAX),
        release: truncate(safeRelease(), RELEASE_MAX),
      }
      if (opts.stack) entry.stack = truncate(sanitize(opts.stack), STACK_MAX)
      const ex = sanitizeExtra(opts.extra)
      if (ex) entry.extra = ex
      const u = currentUrl()
      if (u) entry.url = truncate(u, URL_MAX)
      const ua = currentUserAgent()
      if (ua) entry.user_agent = truncate(ua, UA_MAX)
      const uid = this.safeUserId()
      if (uid) entry.user_id = uid

      this.queue.push(entry)
      if (this.queue.length >= BATCH_SIZE) {
        void this.flush()
      } else {
        this.scheduleFlush()
      }
    } catch {
      // 静默
    }
  }

  private safeUserId(): string | undefined {
    try {
      return this.getUserId()
    } catch {
      return undefined
    }
  }

  private scheduleFlush(): void {
    if (this.timer) return
    this.timer = setTimeout(() => {
      this.timer = null
      void this.flush()
    }, FLUSH_INTERVAL_MS)
  }

  async flush(): Promise<void> {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }
    if (this.queue.length === 0) return
    const batch = this.queue.splice(0, BATCH_SIZE)
    const body = JSON.stringify({ logs: batch })
    try {
      await this.send(body)
    } catch {
      try {
        await this.send(body)
      } catch {
        // 重试仍失败，丢弃
      }
    }
  }

  private async send(body: string): Promise<void> {
    if (this.source === 'client_h5') {
      const f = this.fetchImpl ?? (typeof fetch === 'function' ? fetch : undefined)
      if (!f) throw new Error('no fetch')
      const res = await f(this.endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body,
        credentials: 'include',
        keepalive: true,
      })
      if (!res.ok) throw new Error(`clog upload failed: ${res.status}`)
      return
    }
    // weapp
    if (this.requestImpl) {
      await this.requestImpl({
        url: this.endpoint,
        method: 'POST',
        data: body,
        header: { 'Content-Type': 'application/json' },
      })
      return
    }
    const res = await Taro.request({
      url: this.endpoint,
      method: 'POST',
      data: body,
      header: { 'Content-Type': 'application/json' },
    })
    const status = (res as { statusCode?: number }).statusCode ?? 0
    if (status < 200 || status >= 300) throw new Error(`clog upload failed: ${status}`)
  }

  /** 同步刷新（H5: sendBeacon；weapp: 触发一次异步 Taro.request 但不 await） */
  flushSync(): void {
    try {
      if (this.timer) {
        clearTimeout(this.timer)
        this.timer = null
      }
      if (this.queue.length === 0) return
      const batch = this.queue.splice(0, BATCH_SIZE)
      const body = JSON.stringify({ logs: batch })

      if (this.source === 'client_h5') {
        const beacon =
          this.beaconImpl ??
          (typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function'
            ? (url: string, data: BodyInit) => navigator.sendBeacon(url, data)
            : undefined)
        if (beacon) {
          const blob = typeof Blob !== 'undefined' ? new Blob([body], { type: 'application/json' }) : body
          if (beacon(this.endpoint, blob as BodyInit)) return
        }
        const f = this.fetchImpl ?? (typeof fetch === 'function' ? fetch : undefined)
        if (f) {
          void f(this.endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body,
            credentials: 'include',
            keepalive: true,
          }).catch(() => undefined)
        }
        return
      }
      // weapp：fire and forget
      void this.send(body).catch(() => undefined)
    } catch {
      // 静默
    }
  }

  // 测试辅助
  _peekQueue(): readonly ClogEntry[] {
    return this.queue
  }

  _hasTimer(): boolean {
    return this.timer !== null
  }
}

let singleton: ClogReporter | null = null
let userIdGetter: () => string | undefined = () => undefined

export function getReporter(): ClogReporter {
  if (!singleton) {
    singleton = new ClogReporter({
      getUserId: () => {
        try {
          return userIdGetter()
        } catch {
          return undefined
        }
      },
    })
  }
  return singleton
}

export function setReporter(r: ClogReporter | null): void {
  singleton = r
}

export function configureUserIdGetter(fn: () => string | undefined): void {
  userIdGetter = fn
}

export function report(level: ClogLevel, message: string, opts: ReportOptions = {}): void {
  getReporter().report(level, message, opts)
}

/** 安装平台相关的生命周期 flush 钩子。仅在应用入口调用一次。 */
export function installClogLifecycle(): void {
  try {
    if (isH5Env && typeof window !== 'undefined') {
      window.addEventListener('beforeunload', () => {
        getReporter().flushSync()
      })
      window.addEventListener('pagehide', () => {
        getReporter().flushSync()
      })
    }
  } catch {
    // 静默
  }
}
