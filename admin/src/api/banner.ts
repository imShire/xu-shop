import request from '@/utils/request'

export interface Banner {
  id: string
  title: string
  image_url: string
  link_url: string
  sort: number
  is_active: boolean
  created_at: string
}

export interface BannerForm {
  title?: string
  image_url: string
  link_url?: string
  sort?: number
}

export interface SortItem {
  id: string
  sort: number
}

export const getBanners = () =>
  request.get<any, Banner[]>('/admin/banners')

export const createBanner = (data: BannerForm) =>
  request.post<any, Banner>('/admin/banners', data)

export const updateBanner = (id: string, data: BannerForm) =>
  request.put<any, Banner>(`/admin/banners/${id}`, data)

export const deleteBanner = (id: string) =>
  request.delete(`/admin/banners/${id}`)

export const toggleBanner = (id: string) =>
  request.patch<any, Banner>(`/admin/banners/${id}/toggle`)

export const sortBanners = (items: SortItem[]) =>
  request.patch('/admin/banners/sort', { items })
