import Taro from '@tarojs/taro'
import type { User } from '@/types/biz'
import { request, getAccessToken } from '@/services/api'
import { isWeapp } from '@/utils/platform'

export interface FavoriteItem {
  id: string
  product_id: string
  title: string
  image: string
  price_cents: number
  market_price_cents?: number
  created_at: string
}

export function getFavorites(params?: { page?: number; page_size?: number }) {
  return request<{ list: FavoriteItem[]; total: number }>('/c/favorites', {
    auth: true,
    params: { page: params?.page ?? 1, page_size: params?.page_size ?? 20 },
  })
}

export function addFavorite(productId: string) {
  return request<void>(`/c/favorites/${productId}`, { method: 'POST', auth: true })
}

export function removeFavorite(productId: string) {
  return request<void>(`/c/favorites/${productId}`, { method: 'DELETE', auth: true })
}

export interface HistoryItem {
  id: string
  product_id: string
  title: string
  image: string
  price_cents: number
  viewed_at: string
}

export function getHistory(params?: { page?: number; page_size?: number }) {
  return request<{ list: HistoryItem[]; total: number }>('/c/view-history', {
    auth: true,
    params: { page: params?.page ?? 1, page_size: params?.page_size ?? 20 },
  })
}

export function clearHistory() {
  return request<void>('/c/view-history', { method: 'DELETE', auth: true })
}

export function updateProfile(data: { nickname?: string; gender?: number }) {
  return request<User>('/c/me', { method: 'PUT', auth: true, data })
}

export interface BalanceLog {
  id: string
  change_cents: number
  type: string
  remark?: string
  created_at: string
}

export interface BalanceData {
  balance_cents: number
  logs: BalanceLog[]
  total: number
}

export function getMyBalance(params?: { page?: number; page_size?: number }) {
  return request<BalanceData>('/c/me/balance', {
    auth: true,
    params: { page: params?.page ?? 1, page_size: params?.page_size ?? 20 },
  })
}

const BASE_URL = process.env.TARO_APP_API_BASE || 'http://localhost:8080/api/v1'

export async function uploadAvatar(filePath: string): Promise<User> {
  const token = getAccessToken()
  const header: Record<string, string> = {}
  if (token) {
    header.Authorization = `Bearer ${token}`
  }

  if (isWeapp) {
    const res = await Taro.uploadFile({
      url: `${BASE_URL}/c/me/avatar`,
      filePath,
      name: 'avatar',
      header,
    })
    const payload = JSON.parse(res.data) as { code: number; message: string; data: User }
    if (res.statusCode < 200 || res.statusCode >= 300 || payload.code !== 0) {
      throw new Error(payload.message || '上传失败')
    }
    return payload.data
  }

  // H5 path: fetch + FormData
  const blob = await fetch(filePath).then((r) => r.blob())
  const form = new FormData()
  form.append('avatar', blob, 'avatar.jpg')
  const response = await fetch(`${BASE_URL}/c/me/avatar`, {
    method: 'POST',
    headers: header,
    credentials: 'include',
    body: form,
  })
  const payload = (await response.json()) as { code: number; message: string; data: User }
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.message || '上传失败')
  }
  return payload.data
}
