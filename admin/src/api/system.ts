import request from '@/utils/request'
import type { PageParams } from '@/types'

export const getAuditLogs = (params: PageParams) =>
  request.get<any, any>('/admin/audit-logs', { params })

export const getUploadSettings = () =>
  request.get<any, any>('/admin/settings/upload')

export const updateUploadSettings = (data: any) =>
  request.put('/admin/settings/upload', data)

export const testUploadSettings = (data: any) =>
  request.post('/admin/settings/upload/test', data)

export const probeUploadSettings = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<any, { url: string }>('/admin/settings/upload/probe', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export const getSettings = (group: string) =>
  request.get<any, Record<string, string>>(`/admin/settings/${group}`)

export const updateSettings = (group: string, data: Record<string, string>) =>
  request.put(`/admin/settings/${group}`, data)
