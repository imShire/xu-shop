import { request } from '@/services/api'
import type {
  CommissionRecord,
  Distributor,
  WithdrawOrder,
} from '@/types/biz'

export interface DistributorProfile extends Distributor {
  total_commission_cents: number
  available_commission_cents: number
  withdrawing_cents: number
  withdrawn_cents: number
  invitee_count: number
  monthly_invitee_count: number
  monthly_order_count: number
  share_click_count: number
  share_register_count: number
  share_order_count: number
  share_gmv_cents: number
}

export function applyDistributor(data: {
  real_name: string
  phone: string
  reason?: string
}) {
  return request<{ id: string; status: string }>('/c/distributors/apply', {
    method: 'POST',
    auth: true,
    data,
  })
}

export function getMyDistributor() {
  return request<DistributorProfile | null>('/c/distributors/me', { auth: true })
}

export type CommissionStatus = 'pending' | 'locked' | 'settled' | 'canceled' | 'suspect'

export function getMyCommissions(params?: {
  status?: CommissionStatus
  page?: number
  page_size?: number
}) {
  return request<{ list: CommissionRecord[]; total: number }>(
    '/c/distributors/me/commissions',
    {
      auth: true,
      params: {
        status: params?.status,
        page: params?.page ?? 1,
        page_size: params?.page_size ?? 20,
      },
    },
  )
}

export function getMyWithdraws(params?: { page?: number; page_size?: number }) {
  return request<{ list: WithdrawOrder[]; total: number }>(
    '/c/distributors/me/withdraws',
    {
      auth: true,
      params: { page: params?.page ?? 1, page_size: params?.page_size ?? 20 },
    },
  )
}

export function applyWithdraw(data: { amount_cents: number; sms_code: string }, idempotencyKey: string) {
  return request<WithdrawOrder>('/c/distributors/me/withdraws', {
    method: 'POST',
    auth: true,
    data,
    headers: { 'Idempotency-Key': idempotencyKey },
  })
}

export function sendWithdrawSms() {
  return request<{ sent: boolean; remaining_today?: number }>(
    '/c/distributors/me/withdraw-sms',
    { method: 'POST', auth: true },
  )
}
