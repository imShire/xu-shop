import request from '@/utils/request'
import type { PageParams, PageResult } from '@/types'
import type {
  AftersaleOrder,
  AftersaleOrderDetail,
  AftersaleStatus,
  AftersaleType,
} from '@/types/aftersale'

export interface AftersaleListParams extends PageParams {
  status?: AftersaleStatus | ''
  type?: AftersaleType | ''
  keyword?: string
  applied_from?: string
  applied_to?: string
}

export const getAftersaleList = (params: AftersaleListParams) =>
  request.get<unknown, PageResult<AftersaleOrder>>('/admin/aftersales', { params })

export const getAftersaleDetail = (id: string) =>
  request.get<unknown, AftersaleOrderDetail>(`/admin/aftersales/${id}`)

export interface AgreeReq {
  seller_remark?: string
  return_address_id?: string | null
}
export const agreeAftersale = (id: string, data: AgreeReq = {}, idempotencyKey?: string) =>
  request.post(`/admin/aftersales/${id}/agree`, data, {
    headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {},
  })

export const rejectAftersale = (id: string, reason: string) =>
  request.post(`/admin/aftersales/${id}/reject`, { reason })

export interface ConfirmReceivedReq {
  seller_remark?: string
}
export const confirmReceivedAftersale = (
  id: string,
  data: ConfirmReceivedReq = {},
  idempotencyKey?: string,
) =>
  request.post(`/admin/aftersales/${id}/confirm-received`, data, {
    headers: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {},
  })

export interface AftersaleMessageReq {
  content?: string
  evidence?: string[]
}
export const postAftersaleMessage = (id: string, data: AftersaleMessageReq) =>
  request.post(`/admin/aftersales/${id}/messages`, data)

export const closeAftersale = (id: string, reason: string) =>
  request.post(`/admin/aftersales/${id}/close`, { reason })
