import request from '@/utils/request'
import type { PageParams } from '@/types'

export const getOrderList = (params: PageParams) =>
  request.get<any, any>('/admin/orders', { params })

export const getOrderDetail = (id: string) => request.get<any, any>(`/admin/orders/${id}`)

export const cancelOrder = (id: string, reason: string) =>
  request.post(`/admin/orders/${id}/cancel`, { reason })

export const addOrderRemark = (id: string, content: string) =>
  request.post(`/admin/orders/${id}/remarks`, { content })

export const exportOrders = (params: PageParams) =>
  request.post('/admin/orders/export', params, { responseType: 'blob' })

// 运费模板
export const getFreightTemplates = () => request.get<any, any[]>('/admin/freight-templates')
export const createFreightTemplate = (data: any) => request.post('/admin/freight-templates', data)
export const updateFreightTemplate = (id: string, data: any) =>
  request.put(`/admin/freight-templates/${id}`, data)
export const deleteFreightTemplate = (id: string) =>
  request.delete(`/admin/freight-templates/${id}`)
