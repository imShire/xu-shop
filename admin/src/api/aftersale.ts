import request from '@/utils/request'
import type { PageParams } from '@/types'

export const getAftersaleList = (params: PageParams) =>
  request.get<any, any>('/admin/aftersales', { params })

export const approveAftersale = (id: string) => request.post(`/admin/aftersales/${id}/approve`)

export const rejectAftersale = (id: string, reason: string) =>
  request.post(`/admin/aftersales/${id}/reject`, { reason })
