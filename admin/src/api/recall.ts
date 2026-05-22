import request from '@/utils/request'
import type { PageParams, PageResult } from '@/types'

export interface RecallCampaign {
  id: string
  name: string
  goal?: string
  audience_filter: Record<string, any>
  actions: Array<Record<string, any>>
  trigger_type: 'cron' | 'event' | 'immediate'
  trigger_config: Record<string, any>
  effective_from?: string | null
  effective_to?: string | null
  throttle_per_user_days: number
  daily_quota: number
  total_quota: number
  attribution_window_days: number
  status: 'draft' | 'online' | 'paused' | 'closed'
  created_at?: string
}

export type RecallCampaignForm = Omit<RecallCampaign, 'id' | 'status' | 'created_at'>

export const listRecallCampaigns = (params: PageParams) =>
  request.get<any, PageResult<RecallCampaign>>('/admin/recall-campaigns', { params })

export const getRecallCampaign = (id: string) =>
  request.get<any, RecallCampaign>(`/admin/recall-campaigns/${id}`)

export const createRecallCampaign = (data: RecallCampaignForm) =>
  request.post<any, RecallCampaign>('/admin/recall-campaigns', data)

export const updateRecallCampaign = (id: string, data: RecallCampaignForm) =>
  request.put(`/admin/recall-campaigns/${id}`, data)

export const onlineRecallCampaign = (id: string) =>
  request.post(`/admin/recall-campaigns/${id}/online`)

export const pauseRecallCampaign = (id: string) =>
  request.post(`/admin/recall-campaigns/${id}/pause`)

export const closeRecallCampaign = (id: string) =>
  request.post(`/admin/recall-campaigns/${id}/close`)

export interface RecallFunnel {
  triggered: number
  opened: number
  ordered: number
  gmv_cents: number
  daily?: Array<{
    date: string
    triggered: number
    opened: number
    ordered: number
    gmv_cents: number
  }>
}

export const getRecallFunnel = (id: string, params?: { start?: string; end?: string }) =>
  request.get<any, RecallFunnel>(`/admin/recall-campaigns/${id}/funnel`, { params })

export const listRecallLogs = (
  params: { campaign_id?: string; user_id?: string; export?: boolean } & PageParams
) => request.get<any, PageResult<any> | { url: string }>('/admin/recall-logs', { params })

// 测试触达 — 复用召回日志 / 召回活动接口；如果有专属接口可以再加
export const testSendRecall = (id: string, user_id: string) =>
  request.post(`/admin/recall-campaigns/${id}/test-send`, { user_id })
