import request from '@/utils/request'
import type { PageParams } from '@/types'

// 发件地址
export const getSenderAddresses = () => request.get<any, any[]>('/admin/sender-addresses')
export const createSenderAddress = (data: any) => request.post('/admin/sender-addresses', data)
export const updateSenderAddress = (id: string, data: any) =>
  request.put(`/admin/sender-addresses/${id}`, data)
export const deleteSenderAddress = (id: string) => request.delete(`/admin/sender-addresses/${id}`)
export const setDefaultSenderAddress = (id: string) =>
  request.post(`/admin/sender-addresses/${id}/default`)

// 快递商
export const getCarriers = () => request.get<any, any[]>('/admin/carriers')
export const updateCarrier = (code: string, data: any) => request.put(`/admin/carriers/${code}`, data)

// 发货
export const shipOrder = (orderId: string, data: { carrier_code: string; tracking_no: string }) =>
  request.post(`/admin/orders/${orderId}/ship`, data)

export const batchShipOrders = (
  data: { order_id: string; carrier_code: string; tracking_no: string }[]
): Promise<any> => request.post('/admin/orders/batch-ship', data)

export const getBatchShipStatus = (taskId: string) =>
  request.get<any, any>(`/admin/orders/batch-ship/${taskId}`)

export const getShipmentList = (params: PageParams) =>
  request.get<any, any>('/admin/shipments', { params })

export const getShipmentDetail = (id: string) => request.get<any, any>(`/admin/shipments/${id}`)
