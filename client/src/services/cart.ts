import type { CartItem } from '@/types/biz'
import { request } from '@/services/api'

export interface CartListResult {
  items: CartItem[]
  total: number
}

export interface CartPrecheckConflict {
  cart_item_id: string
  sku_id: string
  reason:
    | 'sku_not_found'
    | 'price_changed'
    | 'stock_insufficient'
    | 'sku_disabled'
    | 'product_deleted'
    | 'product_offsale'
}

export interface CartPrecheckResult {
  conflicts: CartPrecheckConflict[]
  ok: boolean
}

export function getCartCount() {
  return request<{ count: number }>('/c/cart/count', { auth: true })
}

export function getCart() {
  return request<CartListResult>('/c/cart', { auth: true })
}

export function addToCart(data: { sku_id: string; qty: number }) {
  return request<void>('/c/cart', { method: 'POST', auth: true, data })
}

export function updateCartItem(id: string, qty: number) {
  return request<void>(`/c/cart/${id}`, { method: 'PUT', auth: true, data: { qty } })
}

export function deleteCartItem(id: string) {
  return request<void>(`/c/cart/${id}`, { method: 'DELETE', auth: true })
}

export function batchDeleteCartItems(ids: string[]) {
  return request<void>('/c/cart/batch-delete', {
    method: 'POST',
    auth: true,
    data: { ids },
  })
}

export function cleanInvalidCartItems() {
  return request<void>('/c/cart/clean-invalid', {
    method: 'POST',
    auth: true,
  })
}

export function precheckCartItems(ids: string[]) {
  return request<CartPrecheckResult>('/c/cart/precheck', {
    method: 'POST',
    auth: true,
    data: { ids },
  })
}
