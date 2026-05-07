import request from '@/utils/request'
import type { PageParams } from '@/types'

// 分类
export const getCategoryList = () => request.get<any, any[]>('/admin/categories')
export const createCategory = (data: any) => request.post('/admin/categories', data)
export const updateCategory = (id: string, data: any) => request.put(`/admin/categories/${id}`, data)
export const deleteCategory = (id: string) => request.delete(`/admin/categories/${id}`)

// 商品
export const getProductList = (params: PageParams) =>
  request.get<any, any>('/admin/products', { params })
export const getProduct = (id: string) => request.get<any, any>(`/admin/products/${id}`)
export const createProduct = (data: any) => request.post('/admin/products', data)
export const updateProduct = (id: string, data: any) => request.put(`/admin/products/${id}`, data)
export const deleteProduct = (id: string) => request.delete(`/admin/products/${id}`)
export const putOnSale = (id: string) => request.post(`/admin/products/${id}/onsale`)
export const putOffSale = (id: string) => request.post(`/admin/products/${id}/offsale`)
export const copyProduct = (id: string) => request.post(`/admin/products/${id}/copy`)
export const batchOnSale = (ids: string[]) =>
  request.post('/admin/products/batch-status', { ids, status: 'onsale' })
export const batchOffSale = (ids: string[]) =>
  request.post('/admin/products/batch-status', { ids, status: 'offsale' })

// SKU
export const batchUpdateSkuPrice = (
  data: { sku_id: string; price_cents: number; original_price_cents: number }[]
) => request.post('/admin/skus/batch-price', data)

// 图片上传
export const uploadImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return request.post<any, { url: string }>('/admin/upload/image', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
