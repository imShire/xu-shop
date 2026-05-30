export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  request_id?: string
}

export interface User {
  id: string
  phone: string | null
  nickname: string | null
  avatar: string | null
  gender: number
  status: 'active' | 'disabled' | 'deactivating' | 'deactivated'
  created_at?: string
}

export interface AuthTokenPayload {
  access_token: string
  refresh_token?: string
  expires_in: number
  user_id?: string
}

export interface Category {
  id: string
  parent_id?: string
  name: string
  icon?: string
  sort?: number
  status?: string
  tagline?: string
  children?: Category[]
}

export interface Product {
  id: string
  title: string
  subtitle?: string
  main_image: string
  price_min_cents: number
  price_max_cents: number
  tags?: string[] | null
  category_id?: string
  status?: string
  sales?: number
}

export type ProductSort = 'latest' | 'popular'

export interface ProductListPage {
  list: Product[]
  page: number
  page_size: number
  total: number
}

export interface SkuAttr {
  name: string
  values: string[]
}

export interface Sku {
  id: string
  product_id: string
  attrs: Record<string, string>
  price_cents: number
  original_price_cents?: number
  stock: number
  image?: string
}

export interface SpecValue {
  id: string
  value: string
  sort: number
}

export interface Spec {
  id: string
  name: string
  sort: number
  values: SpecValue[]
}

export interface ProductDetail extends Product {
  images?: string[]
  video_url?: string
  detail_html?: string
  specs: Spec[]
  skus: Sku[]
  is_favorite?: boolean
}

export interface CartItem {
  id: string
  sku_id: string
  product_id: string
  product?: Product
  product_title: string
  sku_image: string
  sku_attrs: string[] | Record<string, unknown> | string | null
  qty: number
  snapshot_price_cents: number
  current_price_cents: number
  available_stock: number
  is_available: boolean
  unavailable_reason?: string
}

export interface TimelineItem {
  id: string
  title: string
  summary: string
  time: string
}


export interface Address {
  id: string
  user_id?: string
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
  is_default: boolean
}

export type OrderStatus =
  | 'pending_payment'
  | 'paid'
  | 'shipped'
  | 'delivered'
  | 'completed'
  | 'cancelled'
  | 'refunding'
  | 'refunded'

export interface AddressSnapshot {
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
}

export interface Region {
  code: string
  parent_code?: string
  name: string
  level: number
  has_children: boolean
}

export interface ProductSnapshot {
  title?: string
  main_image?: string
  attrs?: Record<string, string>
}

export interface OrderItem {
  id: string
  order_id: string
  sku_id: string
  product_id: string
  product_snapshot: ProductSnapshot | null
  price_cents: number
  qty: number
  weight_g: number
  created_at: string
}

export interface Order {
  id: string
  order_no: string
  status: OrderStatus
  goods_cents: number
  freight_cents: number
  discount_cents: number
  coupon_discount_cents: number
  total_cents: number
  pay_cents: number
  address_snapshot: AddressSnapshot
  buyer_remark?: string
  expire_at: string
  paid_at?: string
  shipped_at?: string
  completed_at?: string
  cancelled_at?: string
  created_at: string
  updated_at: string
  items?: OrderItem[]
  cancel_request_pending?: boolean
  cancel_request_reason?: string
  cancel_request_at?: string
}

export interface ShipTrack {
  time: string
  content: string
  status?: string
}

// ─── 售后（v1.4）──────────────────────────────────────────────────────────

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

export interface AftersaleExpress {
  carrier_code: string
  waybill_no: string
  shipped_at?: string
}

export interface AftersaleNegotiation {
  id: string
  role: 'buyer' | 'seller' | 'system'
  admin_id?: string | null
  content: string
  evidence?: string[]
  created_at: string
}

export interface AftersaleItemSnapshot {
  product_name?: string
  sku_attrs?: Record<string, string> | string[]
  price_cents?: number
  qty?: number
  image?: string
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
  buyer_evidence?: string[]
  buyer_express?: AftersaleExpress | null
  seller_remark?: string
  refund_id?: string | null
  applied_at: string
  agreed_at?: string | null
  returned_at?: string | null
  received_at?: string | null
  completed_at?: string | null
  closed_at?: string | null
  auto_close_at: string
  item_snapshot?: AftersaleItemSnapshot | null
}

export interface AftersaleOrderDetail extends AftersaleOrder {
  negotiations?: AftersaleNegotiation[]
}

export interface AftersaleApplyReq {
  order_id: string
  order_item_id?: string | null
  type: AftersaleType
  reason: string
  refund_amount_cents: number
  evidence?: string[]
}

export interface AftersaleExpressReq {
  carrier_code: string
  waybill_no: string
}

export interface AftersaleMessageReq {
  content?: string
  evidence?: string[]
}

// ─── v1.2 会员 / 优惠券 ──────────────────────────────────────────────────────

export type CouponType = 'amount' | 'discount' | 'no_threshold' | 'exchange'
export type CouponScopeType = 'all' | 'category' | 'spu' | 'sku'
export type CouponStatus = 'unused' | 'locked' | 'used' | 'expired'

export interface CouponTemplate {
  id: string
  name: string
  description?: string
  type: CouponType
  value_cents: number
  discount_rate?: number | null
  max_discount_cents?: number
  min_amount_cents: number
  scope_type: CouponScopeType
  scope_targets?: string[]
  validity_mode?: 'absolute' | 'relative'
  valid_from?: string | null
  valid_to?: string | null
  valid_days?: number | null
  total_quota?: number
  claimed_count?: number
  used_count?: number
  per_user_limit?: number
  stack_with_points?: boolean
  status?: 'draft' | 'online' | 'offline'
  claim_start_at?: string | null
  claim_end_at?: string | null
  /** 是否已被当前用户领取（后端附加字段，可选） */
  is_claimed?: boolean
}

export interface UserCoupon {
  id: string
  coupon_template_id: string
  name: string
  type: CouponType
  value_cents: number
  discount_rate?: number | null
  max_discount_cents?: number
  min_amount_cents: number
  status: CouponStatus
  order_id?: string | null
  claimed_at: string
  expire_at: string
  used_at?: string | null
  /** 当前订单上下文是否可用（来自 quote 接口） */
  applicable?: boolean
  /** 不可用原因（来自 quote 接口） */
  reason?: string
}

export interface PointAccount {
  balance: number
  locked: number
  total_earned: number
  total_spent: number
}

export interface PointTransaction {
  id: string
  change: number
  type: 'earn' | 'spend' | 'expire' | 'refund' | 'admin_adjust' | 'freeze' | 'unfreeze'
  ref_type?: string | null
  ref_id?: string | null
  balance_after: number
  expire_at?: string | null
  reason: string
  created_at: string
}

export interface MemberLevel {
  code: string
  name: string
  threshold_amount_cents: number
  points_multiplier: number
  benefits?: Record<string, unknown>
  sort?: number
  is_active?: boolean
}

export interface OrderQuoteItem {
  sku_id: string
  qty: number
}

export interface OrderQuoteReq {
  items: OrderQuoteItem[]
  address_id?: string | null
  coupon_id?: string | null
  point_used?: number
}

export interface OrderQuoteResp {
  goods_amount_cents: number
  freight_cents: number
  coupon_amount_cents: number
  point_used: number
  point_deduct_cents: number
  pay_amount_cents: number
  point_earn_estimated: number
  applicable_coupons: UserCoupon[]
  max_point_used: number
}

// ─── v1.2 分销 / 分享 ────────────────────────────────────────────────────────

export type ShareScene = 'product' | 'activity' | 'brand' | 'invite_register'

export interface ShareLink {
  id: string
  scene: ShareScene
  target_id?: string | null
  channel_code: string
  short_token: string
  h5_url: string
  wxapp_path: string
  expire_at: string
  click_count: number
  register_count: number
  order_count: number
  gmv_cents: number
}

export interface Distributor {
  id: string
  user_id: string
  nickname: string
  avatar?: string
  level: 'normal' | 'senior'
  rate: number
  status: 'pending' | 'active' | 'disabled'
  apply_at: string
  approved_at?: string | null
}

export type CommissionStatus = 'pending' | 'locked' | 'settled' | 'canceled' | 'suspect'

export interface CommissionRecord {
  id: string
  order_id: string
  order_no: string
  distributor_user_id: string
  level: string
  rate: number
  base_amount_cents: number
  amount_cents: number
  status: CommissionStatus
  suspect_reason?: string | null
  freeze_until: string
  settled_at?: string | null
  created_at: string
}

export type WithdrawStatus = 'pending' | 'processing' | 'success' | 'failed' | 'canceled'

export interface WithdrawOrder {
  id: string
  withdraw_no: string
  amount_cents: number
  channel: 'wx_transfer'
  status: WithdrawStatus
  wx_transfer_no?: string | null
  wx_transfer_state?: string | null
  fail_reason?: string | null
  applied_at: string
  finished_at?: string | null
}
