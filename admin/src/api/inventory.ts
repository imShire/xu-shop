import request from '@/utils/request'
import type { PageParams } from '@/types'

export const getInventoryAlerts = (params: PageParams) =>
  request.get<any, any>('/admin/inventory/alerts', { params })

export const markAlertRead = (id: string) =>
  request.post(`/admin/inventory/alerts/${id}/read`)

export const markAllAlertsRead = () => request.post('/admin/inventory/alerts/read-all')

export const getInventoryLogs = (params: PageParams) =>
  request.get<any, any>('/admin/inventory/logs', { params })

export const adjustInventory = (
  skuId: string,
  data: { change_qty: number; remark?: string }
) => request.post(`/admin/skus/${skuId}/adjust`, data)
