import { request } from './api'

export interface NavIcon {
  id: string
  title: string
  icon_url: string
  link_url: string
  sort: number
}

export const getNavIcons = () => request<NavIcon[]>('/c/nav-icons')
