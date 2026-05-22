import request from '@/utils/request'
import type { PageParams, PageResult } from '@/types'

// ========== 优惠券模板 ==========
export interface CouponTemplate {
  id: string
  name: string
  description?: string
  type: 'amount' | 'discount' | 'no_threshold' | 'exchange'
  value_cents: number
  discount_rate?: number | null
  max_discount_cents: number
  min_amount_cents: number
  scope_type: 'all' | 'category' | 'spu' | 'sku'
  scope_targets: string[]
  validity_mode: 'absolute' | 'relative'
  valid_from?: string | null
  valid_to?: string | null
  valid_days?: number | null
  total_quota: number
  used_count: number
  per_user_limit: number
  per_order_limit: number
  stack_with_points: boolean
  status: 'draft' | 'online' | 'offline'
  claim_start_at?: string | null
  claim_end_at?: string | null
  created_at?: string
}

export type CouponTemplateForm = Omit<CouponTemplate, 'id' | 'used_count' | 'status' | 'created_at'>

export const listCouponTemplates = (params: PageParams) =>
  request.get<any, PageResult<CouponTemplate>>('/admin/coupon-templates', { params })

export const getCouponTemplate = (id: string) =>
  request.get<any, CouponTemplate>(`/admin/coupon-templates/${id}`)

export const createCouponTemplate = (data: CouponTemplateForm) =>
  request.post<any, CouponTemplate>('/admin/coupon-templates', data)

export const updateCouponTemplate = (id: string, data: CouponTemplateForm) =>
  request.put(`/admin/coupon-templates/${id}`, data)

export const onlineCouponTemplate = (id: string) =>
  request.post(`/admin/coupon-templates/${id}/online`)

export const offlineCouponTemplate = (id: string) =>
  request.post(`/admin/coupon-templates/${id}/offline`)

// ========== 定向发放任务 ==========
export interface CouponGrantTask {
  id: string
  template_id: string
  template_name?: string
  filter: Record<string, any>
  total: number
  succeeded: number
  failed: number
  status: 'pending' | 'running' | 'done' | 'failed'
  fail_csv_url?: string
  created_at: string
}

export const listCouponGrantTasks = (params: PageParams) =>
  request.get<any, PageResult<CouponGrantTask>>('/admin/coupon-grant-tasks', { params })

export const getCouponGrantTask = (id: string) =>
  request.get<any, CouponGrantTask>(`/admin/coupon-grant-tasks/${id}`)

export const createCouponGrantTask = (data: { template_id: string; filter: Record<string, any> }) =>
  request.post<any, CouponGrantTask>('/admin/coupon-grant-tasks', data)

// ========== 兑换码 ==========
export interface CouponRedeemBatch {
  id: string
  template_id: string
  template_name?: string
  count: number
  used_count: number
  csv_url?: string
  created_at: string
}

export const listCouponRedeemBatches = (params: PageParams) =>
  request.get<any, PageResult<CouponRedeemBatch>>('/admin/coupon-redeem-codes/batches', { params })

export const createCouponRedeemBatch = (data: { template_id: string; count: number }) =>
  request.post<any, CouponRedeemBatch>('/admin/coupon-redeem-codes/batches', data)

// ========== 用户券 / 报表 ==========
export const listUserCoupons = (params: { user_id?: string; status?: string } & PageParams) =>
  request.get<any, PageResult<any>>('/admin/user-coupons', { params })

// ========== 积分规则 ==========
export interface PointRule {
  code: string
  enabled: boolean
  config: Record<string, any>
}

export const listPointRules = () =>
  request.get<any, PointRule[]>('/admin/point-rules')

export const updatePointRule = (code: string, data: { enabled?: boolean; config?: Record<string, any> }) =>
  request.put(`/admin/point-rules/${code}`, data)

// ========== 积分流水 ==========
export interface PointTransaction {
  id: string
  user_id: string
  change: number
  type: string
  ref_type?: string
  ref_id?: string
  balance_after: number
  expire_at?: string
  reason: string
  created_at: string
}

export const listPointTransactions = (
  params: { user_id?: string; type?: string; start?: string; end?: string } & PageParams
) => request.get<any, PageResult<PointTransaction>>('/admin/point-transactions', { params })

// ========== 积分调整工单 ==========
export interface PointAdjustTicket {
  id: string
  user_id: string
  change: number
  reason: string
  status: 'pending' | 'approved' | 'rejected'
  applicant_id: string
  approver_id?: string
  created_at: string
}

export const listPointAdjustTickets = (params: PageParams) =>
  request.get<any, PageResult<PointAdjustTicket>>('/admin/point-adjust-tickets', { params })

export const createPointAdjustTicket = (data: { user_id: string; change: number; reason: string }) =>
  request.post<any, PointAdjustTicket>('/admin/point-adjust-tickets', data)

export const approvePointAdjustTicket = (id: string) =>
  request.post(`/admin/point-adjust-tickets/${id}/approve`)

export const rejectPointAdjustTicket = (id: string) =>
  request.post(`/admin/point-adjust-tickets/${id}/reject`)

// ========== 会员等级 ==========
export interface MemberLevel {
  code: string
  name: string
  threshold_amount_cents: number
  points_multiplier: number
  benefits: Record<string, any>
  sort: number
  is_active: boolean
}

export const listMemberLevels = () =>
  request.get<any, MemberLevel[]>('/admin/member-levels')

export const updateMemberLevel = (code: string, data: MemberLevel) =>
  request.put(`/admin/member-levels/${code}`, data)
