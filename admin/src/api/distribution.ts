import request from '@/utils/request'
import type { PageParams, PageResult } from '@/types'

// ========== 分销员 ==========
export interface Distributor {
  id: string
  user_id: string
  nickname?: string
  avatar?: string
  phone?: string
  real_name?: string
  level: 'normal' | 'senior'
  rate: number
  rate_override?: number | null
  status: 'pending' | 'active' | 'disabled'
  apply_at: string
  approved_at?: string | null
  total_commission_cents?: number
  invited_count?: number
}

export const listDistributors = (params: PageParams & { status?: string; level?: string; keyword?: string }) =>
  request.get<any, PageResult<Distributor>>('/admin/distributors', { params })

export const approveDistributor = (id: string) =>
  request.post(`/admin/distributors/${id}/approve`)

export const rejectDistributor = (id: string, reason?: string) =>
  request.post(`/admin/distributors/${id}/reject`, { reason })

export const disableDistributor = (id: string) =>
  request.post(`/admin/distributors/${id}/disable`)

export const setDistributorRate = (id: string, rate_override: number | null) =>
  request.put(`/admin/distributors/${id}/rate`, { rate_override })

// ========== 佣金 ==========
export interface CommissionRecord {
  id: string
  order_id: string
  order_no: string
  distributor_user_id: string
  distributor_nickname?: string
  level: string
  rate: number
  base_amount_cents: number
  amount_cents: number
  status: 'pending' | 'locked' | 'settled' | 'canceled' | 'suspect'
  suspect_reason?: string | null
  freeze_until: string
  settled_at?: string | null
  created_at: string
}

export const listCommissions = (
  params: PageParams & { status?: string; distributor_user_id?: string }
) => request.get<any, PageResult<CommissionRecord>>('/admin/commissions', { params })

export const releaseCommission = (id: string) =>
  request.post(`/admin/commissions/${id}/release`)

export const cancelCommission = (id: string, reason: string) =>
  request.post(`/admin/commissions/${id}/cancel`, { reason })

// ========== 提现 ==========
export interface WithdrawOrder {
  id: string
  withdraw_no: string
  user_id: string
  nickname?: string
  amount_cents: number
  channel: 'wx_transfer'
  status: 'pending' | 'processing' | 'success' | 'failed' | 'canceled'
  wx_transfer_no?: string | null
  wx_transfer_state?: string | null
  fail_reason?: string | null
  applied_at: string
  finished_at?: string | null
}

export const listWithdraws = (params: PageParams & { status?: string }) =>
  request.get<any, PageResult<WithdrawOrder>>('/admin/withdraws', { params })

export const retryWithdraw = (id: string) =>
  request.post(`/admin/withdraws/${id}/retry`)

// ========== 分享链接 ==========
export interface ShareLinkStat {
  id: string
  user_id: string
  nickname?: string
  scene: 'product' | 'activity' | 'brand' | 'invite_register'
  target_id?: string | null
  channel_code: string
  short_token: string
  click_count: number
  register_count: number
  order_count: number
  gmv_cents: number
  expire_at: string
}

export const listShareLinks = (params: PageParams & { user_id?: string; scene?: string }) =>
  request.get<any, PageResult<ShareLinkStat>>('/admin/share-links', { params })

// ========== 分销看板 ==========
export interface DistributionDashboard {
  share_count: number
  click_uv: number
  register_count: number
  order_count: number
  gmv_cents: number
  commission_cents: number
  daily?: Array<{
    date: string
    share_count: number
    click_uv: number
    register_count: number
    order_count: number
    gmv_cents: number
    commission_cents: number
  }>
}

export const getDistributionDashboard = (params?: { start?: string; end?: string }) =>
  request.get<any, DistributionDashboard>('/admin/distribution/dashboard', { params })

// ========== 全局设置（写入 system settings group=distribution）==========
export interface DistributionSettings {
  apply_mode: 'open' | 'invite_only' | 'audit'
  self_purchase_allowed: boolean
  binding_ttl_days: number
  freeze_days: number
  default_rate: number
  senior_rate: number
}

export const getDistributionSettings = () =>
  request.get<any, Record<string, string>>('/admin/settings/distribution')

export const updateDistributionSettings = (data: Record<string, string>) =>
  request.put('/admin/settings/distribution', data)
