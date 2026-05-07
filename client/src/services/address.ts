import type { Address } from '@/types/biz'
import { request } from '@/services/api'

export function getAddresses() {
  return request<Address[]>('/c/addresses', { auth: true })
}

export function getAddress(id: string) {
  return request<Address>(`/c/addresses/${id}`, { auth: true })
}

export interface AddressPayload {
  name: string
  phone: string
  province_code?: string
  province: string
  city_code?: string
  city: string
  district_code?: string
  district: string
  street_code?: string
  street?: string
  detail: string
  is_default?: boolean
}

export function createAddress(data: AddressPayload) {
  return request<Address>('/c/addresses', { method: 'POST', auth: true, data })
}

export function updateAddress(id: string, data: AddressPayload) {
  return request<Address>(`/c/addresses/${id}`, { method: 'PUT', auth: true, data })
}

export function deleteAddress(id: string) {
  return request<void>(`/c/addresses/${id}`, { method: 'DELETE', auth: true })
}

export function setDefaultAddress(id: string) {
  return request<void>(`/c/addresses/${id}/default`, { method: 'POST', auth: true })
}
