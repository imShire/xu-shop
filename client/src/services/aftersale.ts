import Taro from '@tarojs/taro'
import type {
  AftersaleApplyReq,
  AftersaleExpressReq,
  AftersaleMessageReq,
  AftersaleOrder,
  AftersaleOrderDetail,
  AftersaleStatus,
} from '@/types/biz'
import { request, getAccessToken } from '@/services/api'
import { isWeapp } from '@/utils/platform'

const BASE_URL = process.env.TARO_APP_API_BASE || 'http://localhost:8080/api/v1'

function genIdempotencyKey() {
  const c = (globalThis as { crypto?: { randomUUID?: () => string } }).crypto
  if (c?.randomUUID) return c.randomUUID()
  return `idem_${Date.now()}_${Math.random().toString(36).slice(2, 10)}`
}

export interface AftersaleListResp {
  list: AftersaleOrder[]
  total: number
  page?: number
  page_size?: number
}

export function listAftersales(params?: {
  status?: AftersaleStatus
  page?: number
  page_size?: number
}) {
  return request<AftersaleListResp>('/c/aftersales', {
    auth: true,
    params: {
      status: params?.status,
      page: params?.page ?? 1,
      page_size: params?.page_size ?? 20,
    },
  })
}

export function getAftersaleDetail(id: string) {
  return request<AftersaleOrderDetail>(`/c/aftersales/${id}`, { auth: true })
}

export function applyAftersale(data: AftersaleApplyReq) {
  return request<{ id: string; aftersale_no: string }>('/c/aftersales', {
    method: 'POST',
    auth: true,
    data,
    headers: { 'Idempotency-Key': genIdempotencyKey() },
  })
}

export function submitAftersaleExpress(id: string, data: AftersaleExpressReq) {
  return request<void>(`/c/aftersales/${id}/express`, {
    method: 'POST',
    auth: true,
    data,
    headers: { 'Idempotency-Key': genIdempotencyKey() },
  })
}

export function postAftersaleMessage(id: string, data: AftersaleMessageReq) {
  return request<void>(`/c/aftersales/${id}/messages`, {
    method: 'POST',
    auth: true,
    data,
  })
}

export function cancelAftersale(id: string) {
  return request<void>(`/c/aftersales/${id}/cancel`, {
    method: 'POST',
    auth: true,
  })
}

/**
 * 上传售后凭证图片，返回 OSS URL。
 * TODO(backend): 需要后端补 `POST /c/upload/image`（multipart, field=file, 返回 {url}）。
 * 当前后端未提供 C 端图片上传端点，调用后会返回 404；UI 已做错误兜底。
 */
export async function uploadAftersaleEvidence(filePath: string): Promise<string> {
  const token = getAccessToken()
  const header: Record<string, string> = {}
  if (token) header.Authorization = `Bearer ${token}`

  if (isWeapp) {
    const res = await Taro.uploadFile({
      url: `${BASE_URL}/c/upload/image`,
      filePath,
      name: 'file',
      header,
    })
    const payload = JSON.parse(res.data) as {
      code: number
      message: string
      data: { url: string }
    }
    if (res.statusCode < 200 || res.statusCode >= 300 || payload.code !== 0) {
      throw new Error(payload.message || '上传失败')
    }
    return payload.data.url
  }

  const blob = await fetch(filePath).then((r) => r.blob())
  const form = new FormData()
  form.append('file', blob, 'evidence.jpg')
  const response = await fetch(`${BASE_URL}/c/upload/image`, {
    method: 'POST',
    headers: header,
    credentials: 'include',
    body: form,
  })
  const payload = (await response.json()) as {
    code: number
    message: string
    data: { url: string }
  }
  if (!response.ok || payload.code !== 0) {
    throw new Error(payload.message || '上传失败')
  }
  return payload.data.url
}

// ─── 常用快递公司（10 个） ────────────────────────────────────────────────
export const CARRIER_OPTIONS: Array<{ code: string; name: string }> = [
  { code: 'sf', name: '顺丰速运' },
  { code: 'jd', name: '京东物流' },
  { code: 'yto', name: '圆通速递' },
  { code: 'zto', name: '中通快递' },
  { code: 'sto', name: '申通快递' },
  { code: 'yunda', name: '韵达速递' },
  { code: 'ems', name: 'EMS' },
  { code: 'huitongkuaidi', name: '百世快递' },
  { code: 'debangkuaidi', name: '德邦快递' },
  { code: 'youzhengguonei', name: '邮政包裹' },
]

// ─── 文案映射 ────────────────────────────────────────────────────────────

export const AFTERSALE_TYPE_LABEL: Record<string, string> = {
  refund_only: '仅退款',
  refund_return: '退货退款',
  exchange: '换货',
}

export const AFTERSALE_STATUS_LABEL: Record<AftersaleStatus, string> = {
  applying: '申请中',
  seller_agreed: '商家已同意',
  buyer_returned: '买家已寄回',
  seller_received: '商家已收货',
  completed: '已完成',
  seller_rejected: '商家已拒绝',
  cancelled: '已撤销',
  closed: '已关闭',
}

export const AFTERSALE_TERMINAL_STATUSES: AftersaleStatus[] = [
  'completed',
  'seller_rejected',
  'cancelled',
  'closed',
]

export function isAftersaleTerminal(status: AftersaleStatus): boolean {
  return AFTERSALE_TERMINAL_STATUSES.includes(status)
}
