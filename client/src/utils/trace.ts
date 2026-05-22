import { storage } from '@/utils/storage'

const TRACE_KEY = 'share_trace_id'
const TRACE_TTL_MS = 7 * 24 * 60 * 60 * 1000 // 7 days

interface TraceEntry {
  value: string
  expires_at: number
}

/** Persist a share-token trace ID with a 7-day TTL. */
export function persistTraceId(token: string | null | undefined): void {
  if (!token) {
    return
  }
  const entry: TraceEntry = {
    value: String(token),
    expires_at: Date.now() + TRACE_TTL_MS,
  }
  storage.set(TRACE_KEY, entry)
}

/** Read the active trace ID; returns empty string if expired or missing. */
export function getTraceId(): string {
  const entry = storage.get<TraceEntry>(TRACE_KEY)
  if (!entry || typeof entry !== 'object') {
    return ''
  }
  if (!entry.expires_at || Date.now() > entry.expires_at) {
    storage.remove(TRACE_KEY)
    return ''
  }
  return entry.value || ''
}

export function clearTraceId(): void {
  storage.remove(TRACE_KEY)
}

/**
 * Extract a share token from launch options (mp scene/query) or H5 query string.
 * Looks at: query.st, query.share_token, scene-decoded `st=xxx`.
 */
export function extractTokenFromLaunch(options: {
  query?: Record<string, string | undefined> | undefined
  scene?: string | number | undefined
}): string | null {
  const q = options.query ?? {}
  if (q.st) return String(q.st)
  if (q.share_token) return String(q.share_token)
  if (typeof options.scene === 'string' && options.scene.includes('st_')) {
    // weapp QR scene: e.g. "st_abc123"
    const m = options.scene.match(/st_([\w-]+)/)
    if (m) return m[1]
  }
  return null
}
