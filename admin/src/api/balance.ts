import request from '@/utils/request'

export interface BalanceLog {
  id: string
  user_id: string
  change_cents: number
  type: string
  ref_type?: string
  ref_id?: string
  balance_before_cents: number
  balance_after_cents: number
  operator_id?: string
  remark?: string
  created_at: string
}

export function getUserBalanceLogs(userId: string, params?: { page?: number; page_size?: number }) {
  return request.get<any, { balance_cents: number; list: BalanceLog[]; total: number; page: number; page_size: number }>(
    `/admin/users/${userId}/balance-logs`,
    { params }
  )
}

export function rechargeBalance(userId: string, data: { amount_cents: number; remark?: string }) {
  return request.post<any, void>(`/admin/users/${userId}/recharge`, data)
}
