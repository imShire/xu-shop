import Taro from '@tarojs/taro'
import type { ApiEnvelope } from '@/types/biz'
import { isH5 } from '@/utils/platform'
import { storage } from '@/utils/storage'

const BASE_URL = process.env.TARO_APP_API_BASE || 'http://localhost:8080/api/v1'
const ACCESS_TOKEN_KEY = 'access_token'

function getCookie(name: string) {
  if (!isH5 || typeof document === 'undefined') {
    return ''
  }

  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`))
  return match ? decodeURIComponent(match[1]) : ''
}

function buildUrl(path: string, params?: Record<string, string | number | undefined>) {
  const search = params
    ? Object.entries(params)
      .filter(([, value]) => value !== undefined)
      .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
      .join('&')
    : ''

  if (!search) {
    return `${BASE_URL}${path}`
  }

  return `${BASE_URL}${path}?${search}`
}

export function getAccessToken() {
  return storage.get<string>(ACCESS_TOKEN_KEY)
}

export function setAccessToken(token: string) {
  storage.set(ACCESS_TOKEN_KEY, token)
}

export function clearAccessToken() {
  storage.remove(ACCESS_TOKEN_KEY)
}

export async function request<T>(
  path: string,
  options: {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
    data?: unknown
    params?: Record<string, string | number | undefined>
    auth?: boolean
  } = {}
): Promise<T> {
  const { method = 'GET', data, params, auth = false } = options
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (auth) {
    const token = getAccessToken()
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
  }

  if (auth && isH5 && method !== 'GET') {
    const csrf = getCookie('csrf')
    if (csrf) {
      headers['X-CSRF-Token'] = csrf
    }
  }

  const url = buildUrl(path, params)

  if (isH5) {
    const response = await fetch(url, {
      method,
      headers,
      credentials: 'include',
      body: data && method !== 'GET' ? JSON.stringify(data) : undefined,
    })
    if (response.status === 401) {
      clearAccessToken()
      void Taro.redirectTo({ url: '/pages/auth/login/index' })
      throw new Error('登录已过期，请重新登录')
    }
    if (response.status === 403) {
      throw new Error('无权限访问')
    }
    const payload = (await response.json()) as ApiEnvelope<T>
    if (!response.ok || payload.code !== 0) {
      throw new Error(payload.message || '请求失败')
    }
    return payload.data
  }

  const response = await Taro.request<ApiEnvelope<T>>({
    url,
    method,
    data,
    header: headers,
  })

  if (response.statusCode === 401) {
    clearAccessToken()
    void Taro.redirectTo({ url: '/pages/auth/login/index' })
    throw new Error('登录已过期，请重新登录')
  }
  if (response.statusCode === 403) {
    throw new Error('无权限访问')
  }
  if (response.statusCode < 200 || response.statusCode >= 300 || response.data.code !== 0) {
    throw new Error(response.data.message || '请求失败')
  }

  return response.data.data
}
