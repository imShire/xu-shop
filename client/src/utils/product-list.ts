import type { ProductSort } from '@/types/biz'

export interface ProductListRouteParams {
  categoryId?: string
  keyword?: string
  sort?: ProductSort
}

const PRODUCT_LIST_PAGE = '/pages/product/list/index'
const CATEGORY_PAGE = '/pages/category/index'

export function normalizeProductSort(value?: string): ProductSort {
  return value === 'popular' ? 'popular' : 'latest'
}

export function buildProductListUrl(params: ProductListRouteParams = {}) {
  const query: string[] = []

  if (params.categoryId) {
    query.push(`categoryId=${encodeURIComponent(params.categoryId)}`)
  }

  if (params.keyword) {
    query.push(`keyword=${encodeURIComponent(params.keyword.trim())}`)
  }

  if (params.sort) {
    query.push(`sort=${encodeURIComponent(params.sort)}`)
  }

  return query.length > 0 ? `${PRODUCT_LIST_PAGE}?${query.join('&')}` : PRODUCT_LIST_PAGE
}

export function mapCategoryLinkToProductListUrl(linkUrl: string) {
  if (!linkUrl.startsWith(CATEGORY_PAGE)) {
    return ''
  }

  const [, rawQuery = ''] = linkUrl.split('?')
  if (!rawQuery) {
    return ''
  }

  const queryEntries = rawQuery.split('&').reduce<Record<string, string>>((accumulator, entry) => {
    if (!entry) {
      return accumulator
    }

    const [rawKey, ...rawValueParts] = entry.split('=')
    if (!rawKey) {
      return accumulator
    }

    const value = rawValueParts.length > 0 ? rawValueParts.join('=') : ''
    accumulator[decodeURIComponent(rawKey)] = decodeURIComponent(value)
    return accumulator
  }, {})

  const categoryId = queryEntries.categoryId ?? queryEntries.category_id ?? ''
  const keyword = queryEntries.keyword ?? ''
  const sort = queryEntries.sort ?? ''

  if (!categoryId && !keyword && !sort) {
    return ''
  }

  return buildProductListUrl({
    categoryId: categoryId || undefined,
    keyword: keyword || undefined,
    sort: normalizeProductSort(sort),
  })
}
