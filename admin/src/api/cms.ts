import request from '@/utils/request'
import type { PageResult } from '@/types'

export interface Article {
  id: string
  title: string
  cover?: string
  content?: string
  status: 'draft' | 'published'
  sort: number
  created_at: string
  updated_at: string
}

export function getArticles(params?: {
  keyword?: string
  status?: string
  page?: number
  page_size?: number
}) {
  return request.get<any, PageResult<Article>>('/admin/articles', { params })
}

export function getArticle(id: string) {
  return request.get<any, Article>(`/admin/articles/${id}`)
}

export function createArticle(data: Partial<Article>) {
  return request.post<any, Article>('/admin/articles', data)
}

export function updateArticle(id: string, data: Partial<Article>) {
  return request.put<any, Article>(`/admin/articles/${id}`, data)
}

export function deleteArticle(id: string) {
  return request.delete<any, void>(`/admin/articles/${id}`)
}
