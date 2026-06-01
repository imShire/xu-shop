// 前端日志上报工具（admin）
// 对应后端 D1: POST /api/v1/internal/clog/batch
// 设计目标：
//   - 内存缓冲，满 20 条或 5 秒 flush；
//   - beforeunload 时 sendBeacon 同步发送；
//   - 自身错误静默，绝不影响主流程；
//   - 敏感字段（password/token/secret）自动脱敏。

export type ClogLevel = 'error' | 'warn' | 'info'

export interface ClogEntry {
  source: 'admin'
  level: ClogLevel
  message: string
  stack?: string
  url?: string
  user_agent?: string
  release?: string
  admin_id?: string
  extra?: Record<string, unknown>
}

export interface ReportOptions {
  stack?: string
  extra?: Record<string, unknown>
}

const MSG_MAX = 4 * 1024
const STACK_MAX = 16 * 1024
const EXTRA_MAX = 8 * 1024
const BATCH_SIZE = 20
const FLUSH_INTERVAL_MS = 5000
const SENSITIVE_RE = /("?)(password|token|secret|authorization|csrf)("?\s*[:=]\s*"?)([^",}\s]+)/gi

declare const __APP_VERSION__: string

function safeRelease(): string {
  try {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    return typeof __APP_VERSION__ === 'string' ? __APP_VERSION__ : 'dev'
  } catch {
    return 'dev'
  }
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
      json = json.slice(0, EXTRA_MAX) + '"...[truncated]"'
      try {
        return JSON.parse(json) as Record<string, unknown>
      } catch {
        return { _truncated: true, _raw: json.slice(0, EXTRA_MAX) }
      }
    }
    return JSON.parse(json) as Record<string, unknown>
  } catch {
    return undefined
  }
}

export interface ReporterOptions {
  endpoint?: string
  getAdminId?: () => string | undefined
  // 测试注入
  fetchImpl?: typeof fetch
  beaconImpl?: (url: string, data: BodyInit) => boolean
  now?: () => number
}

export class ClogReporter {
  private queue: ClogEntry[] = []
  private timer: ReturnType<typeof setTimeout> | null = null
  private endpoint: string
  private getAdminId: () => string | undefined
  private fetchImpl: typeof fetch
  private beaconImpl?: (url: string, data: BodyInit) => boolean

  constructor(opts: ReporterOptions = {}) {
    this.endpoint = opts.endpoint ?? '/api/v1/internal/clog/batch'
    this.getAdminId = opts.getAdminId ?? (() => undefined)
    this.fetchImpl = opts.fetchImpl ?? ((...args) => fetch(...args))
    this.beaconImpl = opts.beaconImpl
  }

  report(level: ClogLevel, message: string, opts: ReportOptions = {}): void {
    try {
      const entry: ClogEntry = {
        source: 'admin',
        level,
        message: truncate(sanitize(String(message ?? '')), MSG_MAX),
        release: safeRelease(),
      }
      if (opts.stack) entry.stack = truncate(sanitize(opts.stack), STACK_MAX)
      const ex = sanitizeExtra(opts.extra)
      if (ex) entry.extra = ex
      if (typeof window !== 'undefined') {
        entry.url = window.location?.href
        entry.user_agent = window.navigator?.userAgent
      }
      const adminId = this.getAdminId()
      if (adminId) entry.admin_id = adminId

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
      // 重试一次
      try {
        await this.send(body)
      } catch {
        // 仍失败，丢弃，避免循环
      }
    }
  }

  private async send(body: string): Promise<void> {
    const res = await this.fetchImpl(this.endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
      credentials: 'same-origin',
      keepalive: true,
    })
    if (!res.ok) throw new Error(`clog upload failed: ${res.status}`)
  }

  flushSync(): void {
    try {
      if (this.timer) {
        clearTimeout(this.timer)
        this.timer = null
      }
      if (this.queue.length === 0) return
      const batch = this.queue.splice(0, BATCH_SIZE)
      const body = JSON.stringify({ logs: batch })
      const beacon =
        this.beaconImpl ??
        (typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function'
          ? (url: string, data: BodyInit) => navigator.sendBeacon(url, data)
          : undefined)
      if (beacon) {
        const blob = new Blob([body], { type: 'application/json' })
        const ok = beacon(this.endpoint, blob)
        if (ok) return
      }
      // 降级：fetch keepalive，无 await
      void this.fetchImpl(this.endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body,
        credentials: 'same-origin',
        keepalive: true,
      }).catch(() => undefined)
    } catch {
      // 静默
    }
  }

  // 测试辅助
  _peekQueue(): readonly ClogEntry[] {
    return this.queue
  }
}

let singleton: ClogReporter | null = null
let adminIdGetter: () => string | undefined = () => undefined

export function getReporter(): ClogReporter {
  if (!singleton) {
    singleton = new ClogReporter({
      getAdminId: () => {
        try {
          return adminIdGetter()
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

export function configureAdminIdGetter(fn: () => string | undefined): void {
  adminIdGetter = fn
}

export function report(level: ClogLevel, message: string, opts: ReportOptions = {}): void {
  getReporter().report(level, message, opts)
}

export function installClogLifecycle(): void {
  if (typeof window === 'undefined') return
  window.addEventListener('beforeunload', () => {
    getReporter().flushSync()
  })
}
