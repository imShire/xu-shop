import request from '@/utils/request'
import type { PageParams } from '@/types'

export interface CreateUserParams {
  phone: string
  password: string
  nickname?: string
}

export const getUserList = (params: PageParams & { phone?: string; nickname?: string; status?: string }) =>
  request.get<any, any>('/admin/users', { params })

export const createUser = (data: CreateUserParams) =>
  request.post<any, any>('/admin/users', data)

export const disableUser = (id: string | number) => request.post(`/admin/users/${id}/disable`)

export const enableUser = (id: string | number) => request.post(`/admin/users/${id}/enable`)

export const getUserDetail = (id: string) => request.get<any, any>(`/admin/users/${id}`)

export const exportUsers = (params: { phone?: string; nickname?: string; status?: string }) =>
  request.get('/admin/users/export', { params, responseType: 'blob' })
