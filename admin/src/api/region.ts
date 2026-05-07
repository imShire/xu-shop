import request from '@/utils/request'

export interface RegionNode {
  code: string
  parent_code?: string
  name: string
  level: number
  has_children: boolean
}

const cache = new Map<string, Promise<RegionNode[]>>()

export function getRegions(parentCode = '') {
  if (!cache.has(parentCode)) {
    cache.set(parentCode, request.get<any, RegionNode[]>('/open/regions', {
      params: parentCode ? { parent_code: parentCode } : undefined,
    }))
  }
  return cache.get(parentCode)!
}
