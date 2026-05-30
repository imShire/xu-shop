export type AftersaleType = 'refund_only' | 'refund_return' | 'exchange'

export type AftersaleStatus =
  | 'applying'
  | 'seller_agreed'
  | 'buyer_returned'
  | 'seller_received'
  | 'completed'
  | 'seller_rejected'
  | 'cancelled'
  | 'closed'

export type NegotiationRole = 'buyer' | 'seller' | 'system'

export interface AftersaleExpress {
  carrier_code: string
  waybill_no: string
  shipped_at?: string
}

export interface AftersaleNegotiation {
  id: string
  role: NegotiationRole
  admin_id?: string | null
  content: string
  evidence: string[]
  created_at: string
}

export interface AftersaleItemSnapshot {
  product_name?: string
  sku_attrs?: string
  price_cents?: number
  qty?: number
  image?: string
  [k: string]: unknown
}

export interface AftersaleOrder {
  id: string
  aftersale_no: string
  order_id: string
  order_no: string
  order_item_id?: string | null
  user_id: string
  type: AftersaleType
  status: AftersaleStatus
  reason: string
  refund_amount_cents: number
  buyer_evidence: string[]
  buyer_express?: AftersaleExpress | null
  seller_remark: string
  refund_id?: string | null
  applied_at: string
  agreed_at?: string | null
  returned_at?: string | null
  received_at?: string | null
  completed_at?: string | null
  closed_at?: string | null
  auto_close_at: string
  item_snapshot?: AftersaleItemSnapshot | null
  // 列表场景下后端可能附带的展示字段
  user_nickname?: string
}

export interface AftersaleOrderDetail extends AftersaleOrder {
  negotiations: AftersaleNegotiation[]
}

export const AFTERSALE_TYPE_LABEL: Record<AftersaleType, string> = {
  refund_only: '仅退款',
  refund_return: '退货退款',
  exchange: '换货',
}

export const AFTERSALE_STATUS_LABEL: Record<AftersaleStatus, { label: string; type: '' | 'success' | 'warning' | 'info' | 'danger' | 'primary' }> = {
  applying: { label: '待商家处理', type: 'warning' },
  seller_agreed: { label: '商家已同意', type: 'primary' },
  buyer_returned: { label: '买家已寄回', type: 'primary' },
  seller_received: { label: '商家已收货', type: 'primary' },
  completed: { label: '已完成', type: 'success' },
  seller_rejected: { label: '商家已拒绝', type: 'danger' },
  cancelled: { label: '已取消', type: 'info' },
  closed: { label: '已关闭', type: 'info' },
}

export const NEGOTIATION_ROLE_LABEL: Record<NegotiationRole, string> = {
  buyer: '买家',
  seller: '商家',
  system: '系统',
}

/** 终态：不允许再发起任何业务操作（仅协商留言/手动关闭除外，见 ACTION_MATRIX） */
export const TERMINAL_STATUSES: AftersaleStatus[] = ['completed', 'seller_rejected', 'cancelled', 'closed']

export function isTerminal(status: AftersaleStatus): boolean {
  return TERMINAL_STATUSES.includes(status)
}
