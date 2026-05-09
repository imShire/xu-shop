import request from '@/utils/request'

export type ModuleType = 'product_list' | 'category_entry' | 'rich_text'

export interface PageModule {
  type: ModuleType
  data: Record<string, unknown>
}

export interface PageConfig {
  id: string
  page_key: string
  version: number
  modules: PageModule[]
  is_active: boolean
  created_at: string
}

export function getPageVersions(pageKey = 'home') {
  return request.get<any, PageConfig[]>('/admin/decorate/versions', {
    params: { page_key: pageKey },
  })
}

export function savePageConfig(data: { page_key: string; modules: PageModule[] }) {
  return request.post<any, PageConfig>('/admin/decorate/save', data)
}

export function activatePageConfig(id: string, pageKey = 'home') {
  return request.post<any, void>(`/admin/decorate/activate/${id}`, {}, {
    params: { page_key: pageKey },
  })
}
