import request from '@/utils/request'
import type { PageParams } from '@/types'

export const getPaymentList = (params: PageParams) =>
  request.get<any, any>('/admin/payments', { params })

export const getRefundList = (params: PageParams) =>
  request.get<any, any>('/admin/refunds', { params })

export const createRefund = (orderId: string, data: { amount_cents: number; reason: string }) =>
  request.post(`/admin/orders/${orderId}/refund`, data)

export const getReconcileList = (params: PageParams) =>
  request.get<any, any>('/admin/reconciliation', { params })

export const resolveReconcile = (id: string) =>
  request.post(`/admin/reconciliation/${id}/resolve`)
