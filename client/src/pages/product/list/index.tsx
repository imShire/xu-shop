import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useState } from 'react'
import ProductCard from '@/components/ProductCard'
import PullList from '@/components/PullList'
import { getCategories, getProducts } from '@/services/product'
import type { Category, ProductListPage, ProductSort } from '@/types/biz'
import { normalizeProductSort } from '@/utils/product-list'
import { SearchBar } from '@/ui/nutui'
import './index.scss'

const PAGE_SIZE = 20
const SORT_OPTIONS: Array<{ label: string; value: ProductSort }> = [
  { label: '最新上架', value: 'latest' },
  { label: '销量优先', value: 'popular' },
  { label: '价格从低', value: 'price_asc' as ProductSort },
  { label: '价格从高', value: 'price_desc' as ProductSort },
]

function findCategoryName(categories: Category[], targetId: string): string {
  for (const category of categories) {
    if (category.id === targetId) {
      return category.name
    }

    const childName = findCategoryName(category.children ?? [], targetId)
    if (childName) {
      return childName
    }
  }

  return ''
}

export default function ProductListPage() {
  const routerParams = Taro.getCurrentInstance().router?.params ?? {}
  const categoryId = typeof routerParams.categoryId === 'string' ? routerParams.categoryId : ''
  const initialKeyword = typeof routerParams.keyword === 'string' ? routerParams.keyword : ''
  const initialSort = normalizeProductSort(
    typeof routerParams.sort === 'string' ? routerParams.sort : undefined,
  )

  const [keyword, setKeyword] = useState(initialKeyword)
  const [searchKeyword, setSearchKeyword] = useState(initialKeyword.trim())
  const [sort, setSort] = useState<ProductSort>(initialSort)

  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: getCategories,
    staleTime: 5 * 60 * 1000,
  })

  const queryKey = useMemo(
    () => ['products', 'list', categoryId, searchKeyword, sort] as const,
    [categoryId, searchKeyword, sort],
  )

  const {
    data,
    isFetching,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useInfiniteQuery({
    queryKey,
    queryFn: ({ pageParam = 1 }) =>
      getProducts({
        categoryId: categoryId || undefined,
        keyword: searchKeyword || undefined,
        sort,
        page: pageParam as number,
        page_size: PAGE_SIZE,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce(
        (sum, page) => sum + (page as ProductListPage).list.length,
        0,
      )
      return loaded >= (lastPage as ProductListPage).total
        ? undefined
        : (lastPage as ProductListPage).page + 1
    },
  })

  const products = useMemo(
    () => (data?.pages as ProductListPage[] | undefined)?.flatMap((page) => page.list) ?? [],
    [data],
  )
  // Group products into rows of 2 for 2-column grid (works with PullList's inner View wrappers)
  type Product = (typeof products)[0]
  const productRows = useMemo<[Product, Product | null][]>(() => {
    const rows: [Product, Product | null][] = []
    for (let i = 0; i < products.length; i += 2) {
      rows.push([products[i], products[i + 1] ?? null])
    }
    return rows
  }, [products])
  const total = (data?.pages?.[0] as ProductListPage | undefined)?.total ?? 0
  const categoryName = useMemo(
    () => findCategoryName(categories as Category[], categoryId),
    [categories, categoryId],
  )
  const isFirstLoad = isFetching && products.length === 0

  useEffect(() => {
    if (keyword.trim() === '' && searchKeyword !== '') {
      setSearchKeyword('')
    }
  }, [keyword, searchKeyword])

  const handleLoadMore = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      void fetchNextPage()
    }
  }, [fetchNextPage, hasNextPage, isFetchingNextPage])

  const commitKeyword = useCallback(() => {
    setSearchKeyword(keyword.trim())
  }, [keyword])

  const summaryTitle = categoryName || '全部商品'
  const summaryDesc = searchKeyword
    ? `搜索 “${searchKeyword}” 的结果`
    : sort === 'popular'
      ? '当前按热度从高到低展示'
      : '当前按上新时间排序'

  return (
    <View className='product-list-page'>
      <View className='product-list-page__topbar'>
        <View className='product-list-page__search-row'>
          <View className='product-list-page__search'>
            <SearchBar
              shape='round'
              value={keyword}
              placeholder='搜索商品 / 品牌'
              onChange={(value) => setKeyword(String(value))}
              onSearch={commitKeyword}
            />
          </View>
          <View className='product-list-page__search-button' onClick={commitKeyword}>
            搜索
          </View>
        </View>

        <View className='product-list-page__summary'>
          <Text className='product-list-page__summary-title'>{summaryTitle}</Text>
          <View className='product-list-page__summary-meta'>
            <Text className='product-list-page__summary-desc'>{summaryDesc}</Text>
            <Text className='product-list-page__summary-count'>{total} 件</Text>
          </View>
        </View>

        <View className='product-list-page__sort-bar'>
          {SORT_OPTIONS.map((option) => (
            <View
              key={option.value}
              className={`product-list-page__sort-pill${sort === option.value ? ' product-list-page__sort-pill--active' : ''}`}
              onClick={() => setSort(option.value)}
            >
              {option.label}
            </View>
          ))}
        </View>
      </View>

      <View className='product-list-page__content'>
        <PullList
          key={`${categoryId}-${searchKeyword}-${sort}`}
          data={productRows}
          loading={isFirstLoad}
          hasMore={hasNextPage ?? false}
          onLoadMore={handleLoadMore}
          emptyTitle='暂无商品'
          emptyDescription='换个关键词，或者切换排序试试'
          className='product-list-page__grid-scroll'
          keyExtractor={(row) => String(row[0].id)}
          renderItem={(row) => (
            <View className='product-list-page__grid-row'>
              {row.map((item) => item && (
                <View key={String(item.id)} className='product-list-page__grid-cell'>
                  <ProductCard
                    product={{
                      id: item.id,
                      title: item.title,
                      subtitle: item.subtitle,
                      main_image: item.main_image,
                      price_cents: item.price_min_cents,
                      virtual_sales: item.sales,
                    }}
                  />
                </View>
              ))}
              {row[1] === null && <View className='product-list-page__grid-cell' />}
            </View>
          )}
        />
      </View>
    </View>
  )
}
