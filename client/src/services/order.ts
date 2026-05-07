import type { Order } from '@/types/biz'
import { request } from '@/services/api'

export interface CreateOrderParams {
  address_id: string
  items: Array<{ sku_id: string; qty: number }>
  buyer_remark?: string
  idempotency_key?: string
  use_balance?: boolean
}

export function getOrders(params?: { status?: string; page?: number; page_size?: number }) {
  return request<{ list: Order[]; total: number }>('/c/orders', {
    auth: true,
    params: {
      status: params?.status,
      page: params?.page ?? 1,
      page_size: params?.page_size ?? 20,
    },
  })
}

export function getOrderDetail(id: string) {
  return request<Order>(`/c/orders/${id}`, { auth: true })
}

export function createOrder(data: CreateOrderParams) {
  return request<{ order_id: string; order_no: string }>('/c/orders', {
    method: 'POST',
    auth: true,
    data,
  })
}

export function cancelOrder(id: string) {
  return request<void>(`/c/orders/${id}/cancel`, { method: 'POST', auth: true })
}

export function confirmReceived(id: string) {
  return request<void>(`/c/orders/${id}/confirm`, { method: 'POST', auth: true })
}

export function getShipTracks(id: string) {
  return request<import('@/types/biz').ShipTrack[]>(`/c/orders/${id}/tracks`, { auth: true })
}
