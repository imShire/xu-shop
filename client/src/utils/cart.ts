export function formatCartUnavailableReason(reason?: string) {
  switch (reason) {
    case 'sku_not_found':
      return '商品规格已失效'
    case 'product_deleted':
      return '商品已删除'
    case 'product_offsale':
      return '商品已下架'
    case 'sku_disabled':
      return '规格已失效'
    case 'stock_insufficient':
      return '库存不足'
    default:
      return '当前不可购买'
  }
}

export function formatCartConflictReason(reason: string) {
  switch (reason) {
    case 'sku_not_found':
      return '商品规格已失效'
    case 'sku_disabled':
      return '商品规格已失效'
    case 'product_deleted':
      return '商品已删除'
    case 'product_offsale':
      return '商品已下架'
    case 'price_changed':
      return '价格已更新'
    case 'stock_insufficient':
      return '库存不足'
    default:
      return '商品状态已变化'
  }
}

function normalizeCartSkuAttrList(attrs: string[] | Record<string, unknown> | string | null | undefined): string[] {
  if (!attrs) {
    return []
  }

  if (Array.isArray(attrs)) {
    return attrs.map((item) => String(item)).filter(Boolean)
  }

  if (typeof attrs === 'string') {
    const trimmed = attrs.trim()
    if (!trimmed) {
      return []
    }

    if (trimmed.startsWith('[') || trimmed.startsWith('{')) {
      try {
        return normalizeCartSkuAttrList(JSON.parse(trimmed) as string[] | Record<string, unknown>)
      } catch {
        return [trimmed]
      }
    }

    return [trimmed]
  }

  return Object.values(attrs)
    .map((item) => String(item))
    .filter(Boolean)
}

export function formatCartSkuAttrs(
  attrs: string[] | Record<string, unknown> | string | null | undefined,
  separator = ' / ',
) {
  const parts = normalizeCartSkuAttrList(attrs)
  return parts.length > 0 ? parts.join(separator) : ''
}
