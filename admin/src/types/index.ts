export interface PageParams {
  page?: number
  page_size?: number
  [key: string]: any
}

export interface PageResult<T> {
  list: T[]
  page: number
  page_size: number
  total: number
}

export interface ApiResult<T = any> {
  code: number
  message: string
  data: T
  request_id: string
}

export interface AdminUser {
  id: string
  username: string
  real_name?: string
  avatar?: string
  roles: string[]
  perms: string[]
  status: number
  created_at: string
}

export interface Role {
  id: string
  name: string
  code: string
  perms: string[]
  created_at: string
}

export interface Permission {
  code: string
  name: string
  group: string
}

export interface UploadSettings {
  driver: 'local' | 's3'
  public_base_url: string
  max_size_mb: number
  local_dir: string
  s3_vendor: string
  s3_endpoint: string
  s3_region: string
  s3_bucket: string
  s3_prefix: string
  s3_force_path_style: boolean
  s3_access_key?: string
  s3_secret_key?: string
  s3_access_key_set?: boolean
  s3_secret_key_set?: boolean
}

export interface Category {
  id: string
  name: string
  parent_id: string
  sort: number
  icon?: string
  children?: Category[]
}

export interface Sku {
  id?: string
  spec_values: string[]
  price: number
  original_price: number
  stock: number
  sku_code: string
  image?: string | null
  weight_g: number
  barcode?: string | null
  low_stock_threshold?: number
  status?: 'active' | 'disabled'
}

export interface Product {
  id: string
  title: string
  subtitle?: string
  category_id: string
  category_name?: string
  cover: string
  images: string[]
  specs: { name: string; values: string[] }[]
  skus: Sku[]
  status: 'draft' | 'onsale' | 'offsale'
  tags: string[]
  description: string
  min_price: number
  max_price: number
  total_stock: number
  created_at: string
  updated_at: string
  unit: string
  is_virtual: boolean
  freight_template_id?: string | null
  freight_template_name?: string | null
  virtual_sales: number
  on_sale_at?: string | null
  sort: number
  video_url?: string | null
  detail_html?: string | null
  sales?: number
}

export interface ProductSpec {
  id?: string
  name: string
  sort?: number
  values: ProductSpecValue[]
}

export interface ProductSpecValue {
  id?: string
  value: string
  sort?: number
}

export interface FreightTemplateSimple {
  id: string
  name: string
}

export interface Order {
  id: string
  order_no: string
  user_id: string
  user_nickname: string
  user_phone: string
  items: OrderItem[]
  total_amount: number
  discount_amount: number
  shipping_fee: number
  pay_amount: number
  status: string
  pay_status: string
  shipping_status: string
  address: ShippingAddress
  remark?: string
  created_at: string
  updated_at: string
  paid_at?: string
  shipped_at?: string
  completed_at?: string
  cancel_reason?: string
  cancel_request_pending?: boolean
}

export interface OrderItem {
  id: string
  product_id: string
  product_title: string
  sku_id: string
  spec_values: string[]
  price: number
  quantity: number
  subtotal: number
  cover: string
}

export interface ShippingAddress {
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

export interface OrderLog {
  id: string
  order_id: string
  action: string
  content: string
  operator: string
  created_at: string
}

export interface Shipment {
  id: string
  order_id: string
  order_no: string
  carrier: string
  tracking_no: string
  status: string
  shipped_at: string
  tracks?: ShipmentTrack[]
}

export interface ShipmentTrack {
  time: string
  content: string
  location?: string
}

export interface Aftersale {
  id: string
  order_id: string
  order_no: string
  user_id: string
  user_nickname: string
  type: 'cancel' | 'refund' | 'return'
  status: string
  reason: string
  amount: number
  created_at: string
}

export interface PaymentRecord {
  id: string
  order_no: string
  pay_no: string
  channel: string
  amount: number
  status: string
  created_at: string
}

export interface RefundRecord {
  id: string
  order_no: string
  refund_no: string
  channel: string
  amount: number
  status: string
  reason: string
  created_at: string
}

export interface InventoryAlert {
  id: string
  product_id: string
  product_title: string
  sku_id: string
  sku_code: string
  spec_values: string[]
  stock: number
  threshold: number
  status: 'unread' | 'read'
  created_at: string
}

export interface InventoryLog {
  id: string
  product_id: string
  product_title: string
  sku_id: string
  sku_code: string
  change_type: string
  change_qty: number
  before_qty: number
  after_qty: number
  remark?: string
  operator?: string
  created_at: string
}

export interface User {
  id: string
  openid: string
  nickname: string
  avatar?: string
  phone?: string
  status: number
  created_at: string
  last_login_at?: string
}

export interface StatsOverview {
  order_count: number
  order_amount: number
  refund_amount: number
  net_income: number
  order_count_ratio?: number
  order_amount_ratio?: number
}

export interface StatsTrend {
  date: string
  order_count: number
  order_amount: number
}

export interface Notification {
  id: string
  template_id: string
  template_name: string
  target: string
  status: string
  created_at: string
}

export interface NotificationTemplate {
  id: string
  name: string
  type: string
  external_id?: string
  content?: string
  enabled: boolean
  updated_at: string
}

export interface Carrier {
  id: string
  name: string
  code: string
  enabled: boolean
}

export interface SenderAddress {
  id: string
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

export interface FreightTemplate {
  id: string
  name: string
  is_default: boolean
  free_threshold?: number
  base_fee: number
  rules: FreightRule[]
  created_at: string
}

export interface FreightRule {
  regions: string[]
  extra_fee: number
}

export interface ChannelCode {
  id: string
  name: string
  qrcode_url: string
  scan_count: number
  follow_count: number
  order_count: number
  created_at: string
}

export interface Tag {
  id: string
  name: string
  color: string
  user_count: number
  created_at: string
}

export interface AuditLog {
  id: string
  module: string
  action: string
  operator: string
  operator_id: string
  ip: string
  detail?: string
  created_at: string
}
