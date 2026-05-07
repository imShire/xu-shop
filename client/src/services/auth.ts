import type { User } from '@/types/biz'
import { request } from '@/services/api'

export async function sendSmsCode(phone: string, purpose: 'register' | 'reset') {
  return request<void>('/c/auth/sms/send', {
    method: 'POST',
    data: { phone, purpose },
  })
}

export async function phoneRegister(data: { phone: string; code: string; password: string }) {
  return request<{ token: string; user: User }>('/c/auth/phone-register', {
    method: 'POST',
    data,
  })
}

export async function phoneLogin(data: { phone: string; password: string }) {
  return request<{ token: string; user: User }>('/c/auth/phone-login', {
    method: 'POST',
    data,
  })
}

export async function resetPassword(data: { phone: string; code: string; new_password: string }) {
  return request<void>('/c/auth/reset-password', {
    method: 'POST',
    data,
  })
}

export function getMe() {
  return request<User | null>('/c/me', { auth: true })
}

export function logout() {
  return request<void>('/c/auth/logout', {
    method: 'POST',
    auth: true,
  })
}
