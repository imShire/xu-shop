import request from '@/utils/request'
import type { PageParams } from '@/types'

export const adminLogin = (data: {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}) => request.post<any, { access_token: string; csrf_token: string }>('/admin/auth/login', data)

export const adminLogout = () => request.post('/admin/auth/logout')

export const getCaptcha = () =>
  request.post<any, { captcha_id: string; captcha_b64: string }>('/admin/auth/captcha')

export const getAdminMe = () => request.get<any, any>('/admin/me')

export const getAdminList = (params: PageParams) => request.get<any, any>('/admin/admins', { params })

export const createAdmin = (data: any) => request.post('/admin/admins', data)

export const updateAdmin = (id: string, data: any) => request.put(`/admin/admins/${id}`, data)

export const disableAdmin = (id: string) => request.post(`/admin/admins/${id}/disable`)

export const enableAdmin = (id: string) => request.post(`/admin/admins/${id}/enable`)

export const resetAdminPwd = (id: string, data: { new_password: string }) =>
  request.post(`/admin/admins/${id}/reset-pwd`, data)

export const getRoleList = () => request.get<any, any[]>('/admin/roles')

export const createRole = (data: any) => request.post('/admin/roles', data)

export const updateRole = (id: string, data: any) => request.put(`/admin/roles/${id}`, data)

export const deleteRole = (id: string) => request.delete(`/admin/roles/${id}`)

export const getPermissions = () => request.get<any, any[]>('/admin/permissions')
