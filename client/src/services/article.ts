import { request } from './api'

export interface Article {
  id: string
  title: string
  cover?: string
  content?: string
  status: string
  sort: number
  published_at?: string
  created_at: string
}

export interface ArticleListPage {
  list: Article[]
  total: number
  page: number
  page_size: number
}

export function getArticles(params?: { page?: number; page_size?: number }) {
  return request<ArticleListPage>('/c/articles', {
    params: {
      status: 'published',
      page: params?.page ?? 1,
      page_size: params?.page_size ?? 20,
    },
  })
}

export function getArticle(id: string) {
  return request<Article>(`/c/articles/${id}`)
}
