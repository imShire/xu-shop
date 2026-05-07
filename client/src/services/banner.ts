import { request } from './api'

export interface Banner {
  id: string
  title: string
  image_url: string
  link_url: string
  sort: number
}

export const getBanners = () => request<Banner[]>('/c/banners')
