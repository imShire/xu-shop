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
}

export interface ShipTrack {
  time: string
  content: string
  status?: string
}

export interface AftersaleStatus {
  id: string
  order_id: string
  type: 'refund' | 'return_refund'
  status: 'pending' | 'approved' | 'rejected' | 'completed'
  reason: string
  created_at: string
}
