import type { components } from './api'

type Schemas = components['schemas']

export type Response = Schemas['Response']
export type PagedResult = Schemas['PagedResult']
export type AdminResp = Schemas['AdminResp']
export type UserResp = Schemas['UserResp']
export type ProductListItem = Schemas['ProductListItem']
export type ProductDetail = Schemas['ProductDetail']
export type SkuDetail = Schemas['SkuDetail']
export type SkuInput = Schemas['SkuInput']
export type AddressResp = Schemas['AddressResp']
export type CategoryResp = Schemas['CategoryResp']
export type RoleResp = Schemas['RoleResp']
export type PermissionResp = Schemas['PermissionResp']
