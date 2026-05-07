import type { Category, ProductDetail, ProductListPage, ProductSort } from '@/types/biz'
import { request } from '@/services/api'

export function getCategories() {
  return request<Category[]>('/c/categories')
}

export function getProducts(params?: {
  categoryId?: string
  keyword?: string
  sort?: ProductSort
  page?: number
  page_size?: number
}) {
  return request<ProductListPage>('/c/products', {
    params: {
      category_id: params?.categoryId,
      keyword: params?.keyword,
      sort: params?.sort,
      page: params?.page ?? 1,
      page_size: params?.page_size ?? 20,
    },
  })
}

export function getProductDetail(id: string) {
  return request<ProductDetail>(`/c/products/${id}`, { auth: true })
}
