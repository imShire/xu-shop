import { request } from './api'

export type PageModuleType = 'product_list' | 'category_entry' | 'rich_text'

export interface ProductListData {
  title: string
  sort: 'latest' | 'popular'
  limit: number
}

export interface CategoryEntryItem {
  title: string
  image_url: string
  link_url: string
}

export interface CategoryEntryData {
  items: CategoryEntryItem[]
}

export interface RichTextData {
  content: string
}

export interface PageModule {
  type: PageModuleType
  data: ProductListData | CategoryEntryData | RichTextData | Record<string, unknown>
}

export interface PageConfigResponse {
  id: string
  page_key: string
  version: number
  modules: PageModule[]
  is_active: boolean
  created_at: string
}

export const getPageConfig = (pageKey = 'home') =>
  request<PageConfigResponse>(`/c/page-config?page_key=${pageKey}`)
