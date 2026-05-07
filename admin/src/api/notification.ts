import request from '@/utils/request'
import type { PageParams } from '@/types'

export const getNotificationList = (params: PageParams) =>
  request.get<any, any>('/admin/notifications', { params })

export const getTemplateList = () =>
  request.get<any, any[]>('/admin/notification-templates')

export const updateTemplate = (code: string, data: any) =>
  request.put(`/admin/notification-templates/${code}`, data)

export const testSendNotification = (code: string, data: { openid: string }) =>
  request.post(`/admin/notification-templates/${code}/test`, data)
