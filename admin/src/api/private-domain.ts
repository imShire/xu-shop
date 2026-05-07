import request from '@/utils/request'
import type { PageParams } from '@/types'

// 渠道码
export const getChannelCodeList = (params: PageParams) =>
  request.get<any, any>('/admin/channel-codes', { params })

export const createChannelCode = (data: any) => request.post('/admin/channel-codes', data)

export const updateChannelCode = (id: string | number, data: { name: string; remark?: string }) =>
  request.put(`/admin/channel-codes/${id}`, data)

export const deleteChannelCode = (id: string) => request.delete(`/admin/channel-codes/${id}`)

// 客户标签
export const getTagList = (params?: PageParams) =>
  request.get<any, any>('/admin/tags', { params })

export const createTag = (data: any) => request.post('/admin/tags', data)

export const updateTag = (id: string, data: any) => request.put(`/admin/tags/${id}`, data)

export const deleteTag = (id: string) => request.delete(`/admin/tags/${id}`)
