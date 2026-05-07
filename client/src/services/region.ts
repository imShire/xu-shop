import type { Region } from '@/types/biz'
import { request } from '@/services/api'

const cache = new Map<string, Promise<Region[]>>()

export function getRegions(parentCode = '') {
  if (!cache.has(parentCode)) {
    cache.set(parentCode, request<Region[]>('/open/regions', {
      params: parentCode ? { parent_code: parentCode } : undefined,
    }))
  }
  return cache.get(parentCode)!
}

export function clearRegionCache() {
  cache.clear()
}
