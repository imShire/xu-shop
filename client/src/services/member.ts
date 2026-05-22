import { request } from '@/services/api'
import type { MemberLevel } from '@/types/biz'

export interface MyMemberLevel {
  current: MemberLevel
  next: MemberLevel | null
  /** 累计可计入升级的消费金额（cents） */
  cumulative_amount_cents: number
  /** 距离下一档还需 cents */
  to_next_cents: number
  /** 进度 0..1 */
  progress: number
}

export function getMyMemberLevel() {
  return request<MyMemberLevel>('/c/me/level', { auth: true })
}
