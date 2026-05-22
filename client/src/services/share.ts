import { request } from '@/services/api'
import type { ShareLink } from '@/types/biz'

export type ShareScene = 'product' | 'activity' | 'brand' | 'invite_register'

export function createShareLink(data: {
  scene: ShareScene
  target_id?: string | null
  channel_code?: string
  ttl_days?: number
}) {
  return request<ShareLink>('/c/share/links', {
    method: 'POST',
    auth: true,
    data,
  })
}

export function generatePoster(shareLinkId: string) {
  return request<{ poster_url: string; expires_at?: string }>('/c/share/poster', {
    method: 'POST',
    auth: true,
    data: { share_link_id: shareLinkId },
  })
}

/**
 * Best-effort click attribution upload. Never throws — silently swallows errors
 * so that bad networks / missing endpoints do not block app boot.
 */
export function trackShareClick(token: string): void {
  if (!token) return
  request<void>('/share/track', {
    method: 'POST',
    data: { token },
  }).catch(() => {
    /* noop — fire-and-forget */
  })
}
