import { request } from '@/services/api'
import type { PointAccount, PointTransaction } from '@/types/biz'

export function getPointSummary() {
  return request<PointAccount & { expiring_soon?: number; expiring_at?: string | null }>(
    '/c/me/points/summary',
    { auth: true },
  )
}

export function getPointTransactions(params?: {
  type?: string
  page?: number
  page_size?: number
}) {
  return request<{ list: PointTransaction[]; total: number }>('/c/me/points/transactions', {
    auth: true,
    params: {
      type: params?.type,
      page: params?.page ?? 1,
      page_size: params?.page_size ?? 20,
    },
  })
}

export function dailySignIn() {
  return request<{ change?: number; balance_after?: number }>('/c/me/sign-in', {
    method: 'POST',
    auth: true,
  })
}
