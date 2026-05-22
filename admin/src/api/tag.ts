import request from '@/utils/request'
import type { PageParams, PageResult } from '@/types'

export type TagCategory =
  | 'rfm'
  | 'lifecycle'
  | 'category_pref'
  | 'price_band'
  | 'source'
  | 'business'
  | 'member'
  | 'system'

export const tagCategoryLabels: Record<TagCategory, string> = {
  rfm: 'RFM',
  lifecycle: '生命周期',
  category_pref: '品类偏好',
  price_band: '价格带',
  source: '来源',
  business: '业务',
  member: '会员',
  system: '系统',
}

export interface UserTag {
  code: string
  name: string
  category: TagCategory
  parent_code?: string | null
  color?: string | null
  description?: string
  source: 'auto' | 'manual'
  config: Record<string, any>
  enabled: boolean
  sort: number
}

export type UserTagForm = UserTag

export const listUserTags = (params?: { category?: string; source?: string }) =>
  request.get<any, UserTag[]>('/admin/user-tags', { params })

export const createUserTag = (data: UserTagForm) =>
  request.post<any, UserTag>('/admin/user-tags', data)

export const updateUserTag = (code: string, data: UserTagForm) =>
  request.put(`/admin/user-tags/${code}`, data)

export const deleteUserTag = (code: string) =>
  request.delete(`/admin/user-tags/${code}`)

export interface UserTagBinding {
  tag_code: string
  tag_name: string
  category: TagCategory
  source: 'auto' | 'manual'
  score?: number
  expire_at?: string | null
  granted_at: string
  granted_by?: string
}

export const getUserTags = (userId: string) =>
  request.get<any, UserTagBinding[]>(`/admin/users/${userId}/tags`)

export const grantUserTag = (
  userId: string,
  data: { tag_code: string; expire_at?: string | null }
) => request.post(`/admin/users/${userId}/tags`, data)

export const revokeUserTag = (userId: string, tagCode: string) =>
  request.delete(`/admin/users/${userId}/tags/${tagCode}`)

export interface AudiencePreviewResp {
  total: number
  sample?: Array<{ user_id: string; nickname?: string; phone?: string }>
}

export const previewAudience = (filter: Record<string, any>) =>
  request.post<any, AudiencePreviewResp>('/admin/audience/preview', { filter })

// 用户搜索（用于"按 user_id/手机号"查标签）
export const searchUserBrief = (params: { keyword?: string } & PageParams) =>
  request.get<any, PageResult<{ id: string; nickname?: string; phone?: string }>>('/admin/users', { params })
