import request from '@/utils/request'
import type { LinkConfig } from '@/types/link'

export interface NavIcon {
  id: string
  title: string
  icon_url: string
  link_url: string
  sort: number
  is_active: boolean
  created_at: string
  link_config?: LinkConfig | null
}

export interface NavIconForm {
  title?: string
  icon_url: string
  link_url?: string
  sort?: number
  link_config?: LinkConfig | null
}

export interface SortItem {
  id: string
  sort: number
}

export const getNavIcons = () =>
  request.get<any, NavIcon[]>('/admin/nav-icons')

export const createNavIcon = (data: NavIconForm) =>
  request.post<any, NavIcon>('/admin/nav-icons', data)

export const updateNavIcon = (id: string, data: NavIconForm) =>
  request.put<any, NavIcon>(`/admin/nav-icons/${id}`, data)

export const deleteNavIcon = (id: string) =>
  request.delete(`/admin/nav-icons/${id}`)

export const toggleNavIcon = (id: string) =>
  request.patch<any, NavIcon>(`/admin/nav-icons/${id}/toggle`)

export const sortNavIcons = (items: SortItem[]) =>
  request.patch('/admin/nav-icons/sort', { items })
