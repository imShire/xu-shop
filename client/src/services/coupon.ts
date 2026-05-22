import { request } from '@/services/api'
import type { CouponTemplate, UserCoupon } from '@/types/biz'

export function getAvailableCoupons(params?: { page?: number; page_size?: number }) {
  return request<{ list: CouponTemplate[]; total: number; page: number; page_size: number }>(
    '/c/coupons/available',
    {
      auth: true,
      params: { page: params?.page ?? 1, page_size: params?.page_size ?? 20 },
    },
  )
}

export function claimCoupon(templateId: string, idempotencyKey?: string) {
  return request<void>(`/c/coupons/${templateId}/claim`, {
    method: 'POST',
    auth: true,
    headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : undefined,
  })
}

export function redeemCoupon(code: string) {
  return request<{ user_coupon_id?: string }>('/c/coupons/redeem', {
    method: 'POST',
    auth: true,
    data: { code },
  })
}

export type UserCouponStatus = 'unused' | 'locked' | 'used' | 'expired'

export function getMyCoupons(params?: {
  status?: UserCouponStatus
  page?: number
  page_size?: number
}) {
  return request<{ list: UserCoupon[]; total: number }>('/c/me/coupons', {
    auth: true,
    params: {
      status: params?.status,
      page: params?.page ?? 1,
      page_size: params?.page_size ?? 20,
    },
  })
}
