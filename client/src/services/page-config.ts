import { request } from './api'

export type PageModuleType = 'product_list' | 'category_entry' | 'rich_text' | 'image_ad'

export interface ProductListData {
  title: string
  sort: 'latest' | 'popular'
  limit: number
  product_ids?: string[]
  manual?: boolean
}

export interface ImageAdData {
  image_url: string
  alt?: string
  link_config?: { type: string; url: string; target_id?: string; target_name?: string } | null
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
  data: ProductListData | CategoryEntryData | RichTextData | ImageAdData | Record<string, unknown>
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
