import { Input, ScrollView, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useInfiniteQuery } from '@tanstack/react-query'
import { useCallback, useRef, useState } from 'react'
import ProductCard from '@/components/ProductCard'
import { getProducts } from '@/services/product'
import type { ProductListPage } from '@/types/biz'
import './index.scss'

const PAGE_SIZE = 20
const HISTORY_KEY = 'search_history'
const MAX_HISTORY = 10

function getHistory(): string[] {
  try {
    return (Taro.getStorageSync(HISTORY_KEY) as string[]) || []
  } catch {
    return []
  }
}

function saveHistory(keyword: string) {
  try {
    const prev = getHistory().filter((k) => k !== keyword)
    const next = [keyword, ...prev].slice(0, MAX_HISTORY)
    Taro.setStorageSync(HISTORY_KEY, next)
  } catch {
    // ignore
  }
}

function removeHistory(keyword: string) {
  try {
    const next = getHistory().filter((k) => k !== keyword)
    Taro.setStorageSync(HISTORY_KEY, next)
  } catch {
    // ignore
  }
}

function clearHistory() {
  try {
    Taro.setStorageSync(HISTORY_KEY, [])
  } catch {
    // ignore
  }
}

export default function SearchPage() {
  const loadingRef = useRef(false)
  const [draft, setDraft] = useState('')
  const [keyword, setKeyword] = useState('')
  const [history, setHistory] = useState<string[]>(() => getHistory())

  const hasResults = keyword.trim().length > 0

  const { data, isFetching, isFetchingNextPage, hasNextPage, fetchNextPage } = useInfiniteQuery({
    queryKey: ['products', 'search', keyword],
    queryFn: ({ pageParam = 1 }) =>
      getProducts({ keyword, page: pageParam as number, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    enabled: hasResults,
    getNextPageParam: (lastPage) => {
      const loaded = lastPage.page * lastPage.page_size
      return loaded >= lastPage.total ? undefined : lastPage.page + 1
    },
  })

  const allProducts = data?.pages.flatMap((page: ProductListPage) => page.list) ?? []

  function commitSearch(kw: string) {
    const trimmed = kw.trim()
    if (!trimmed) return
    setDraft(trimmed)
    setKeyword(trimmed)
    saveHistory(trimmed)
    setHistory(getHistory())
  }

  function handleHistoryTap(kw: string) {
    commitSearch(kw)
  }

  function handleHistoryLongPress(kw: string) {
    removeHistory(kw)
    setHistory(getHistory())
  }

  function handleClearAll() {
    clearHistory()
    setHistory([])
  }

  const handleScrollToLower = useCallback(() => {
    if (loadingRef.current || !hasNextPage || isFetchingNextPage) return
    loadingRef.current = true
    void fetchNextPage().finally(() => {
      loadingRef.current = false
    })
  }, [fetchNextPage, hasNextPage, isFetchingNextPage])

  return (
    <View className='search-page'>
      {/* Search bar */}
      <View className='search-page__bar'>
        <View className='search-page__input-wrap'>
          <Input
            className='search-page__input'
            value={draft}
            placeholder='搜索商品'
            focus
            confirmType='search'
            onInput={(e) => setDraft(e.detail.value)}
            onConfirm={() => commitSearch(draft)}
          />
        </View>
        <View className='search-page__btn' onClick={() => commitSearch(draft)}>
          <Text className='search-page__btn-text'>搜索</Text>
        </View>
      </View>

      {/* Content area */}
      {!hasResults ? (
        <View className='search-page__history'>
          {history.length > 0 ? (
            <>
              <View className='search-page__history-head'>
                <Text className='search-page__history-title'>最近搜索</Text>
                <Text className='search-page__history-clear' onClick={handleClearAll}>
                  清空
                </Text>
              </View>
              <View className='search-page__chips'>
                {history.map((kw) => (
                  <View
                    key={kw}
                    className='search-page__chip'
                    onClick={() => handleHistoryTap(kw)}
                    onLongPress={() => handleHistoryLongPress(kw)}
                  >
                    <Text className='search-page__chip-text'>{kw}</Text>
                  </View>
                ))}
              </View>
            </>
          ) : (
            <View className='search-page__history-empty'>
              <Text className='search-page__history-empty-text'>暂无搜索记录</Text>
            </View>
          )}
        </View>
      ) : (
        <ScrollView
          className='search-page__results'
          scrollY
          onScrollToLower={handleScrollToLower}
          lowerThreshold={120}
        >
          {isFetching && allProducts.length === 0 ? (
            <View className='search-page__loading'>
              <Text>搜索中...</Text>
            </View>
          ) : allProducts.length === 0 ? (
            <View className='search-page__empty'>
              <Text className='search-page__empty-text'>暂无相关商品</Text>
            </View>
          ) : (
            <View className='search-page__grid'>
              {allProducts.map((product) => (
                <View key={product.id} className='search-page__grid-item'>
                  <ProductCard
                    product={{
                      id: product.id,
                      title: product.title,
                      subtitle: product.subtitle,
                      main_image: product.main_image,
                      price_cents: product.price_min_cents,
                    }}
                  />
                </View>
              ))}
            </View>
          )}
          {isFetchingNextPage && (
            <View className='search-page__loadmore'>
              <Text className='search-page__loadmore-text'>加载中...</Text>
            </View>
          )}
        </ScrollView>
      )}
    </View>
  )
}
