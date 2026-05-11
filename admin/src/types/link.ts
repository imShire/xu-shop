export type LinkType = 'product' | 'category' | 'article' | 'product_list' | 'custom'

export interface LinkConfig {
  type: LinkType
  target_id?: string
  target_name?: string
  url: string
}
