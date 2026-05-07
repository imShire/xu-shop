import request from '@/utils/request'

export const getStatsOverview = (params: { from: string; to: string }) =>
  request.get<any, any>('/admin/stats/overview', { params })

export const getStatsTrend = (params: {
  from: string
  to: string
  granularity?: 'day' | 'week' | 'month'
}) => request.get<any, any[]>('/admin/stats/sales-trend', { params })

export const getCategoryStats = (params: { from: string; to: string }) =>
  request.get<any, any[]>('/admin/stats/category-pie', { params })

export const getProductRanking = (params: {
  from: string
  to: string
  page?: number
  page_size?: number
}) => request.get<any, any>('/admin/stats/products', { params })

export const getUserStats = (params: { from: string; to: string }) =>
  request.get<any, any>('/admin/stats/users', { params })

export const getChannelStats = (params: { from: string; to: string }) =>
  request.get<any, any[]>('/admin/stats/channels', { params })

export const exportStats = (params: any) =>
  request.get('/admin/stats/products/export', { params, responseType: 'blob' })

export const getWorkbenchStats = () => request.get<any, any>('/admin/stats/workbench')
