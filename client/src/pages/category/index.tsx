import { Image, ScrollView, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useRef, useState } from 'react'
import EmptyState from '@/components/EmptyState'
import { getCategories } from '@/services/product'
import type { Category } from '@/types/biz'
import { buildProductListUrl, normalizeProductSort } from '@/utils/product-list'
import { Search } from '@/ui/icons'
import './index.scss'

function SubItem({ item, onTap }: { item: Category; onTap: (id: string) => void }) {
  return (
    <View className='category-page__sub-item' onClick={() => onTap(item.id)}>
      {item.icon ? (
        <Image className='category-page__sub-img' src={item.icon} mode='aspectFill' />
      ) : (
        <View className='category-page__sub-img category-page__sub-img--placeholder' />
      )}
      <Text className='category-page__sub-name'>{item.name}</Text>
    </View>
  )
}

export default function CategoryPage() {
  const routerParams = Taro.getCurrentInstance().router?.params ?? {}
  const legacyCategoryId = typeof routerParams.categoryId === 'string' ? routerParams.categoryId : ''
  const legacyKeyword = typeof routerParams.keyword === 'string' ? routerParams.keyword : ''
  const legacySort = typeof routerParams.sort === 'string' ? routerParams.sort : ''
  const shouldRedirectToList = Boolean(legacyCategoryId || legacyKeyword || legacySort)

  const [activeCatId, setActiveCatId] = useState('')
  const [rightScrollId, setRightScrollId] = useState('')
  const [leftScrollId, setLeftScrollId] = useState('')
  const sectionOffsetsRef = useRef<Map<string, number>>(new Map())
  const isClickScrollingRef = useRef(false)

  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: getCategories,
    staleTime: 5 * 60 * 1000,
  })

  useEffect(() => {
    if (!shouldRedirectToList) {
      return
    }

    void Taro.redirectTo({
      url: buildProductListUrl({
        categoryId: legacyCategoryId || undefined,
        keyword: legacyKeyword || undefined,
        sort: legacySort ? normalizeProductSort(legacySort) : undefined,
      }),
    })
  }, [legacyCategoryId, legacyKeyword, legacySort, shouldRedirectToList])

  useEffect(() => {
    if (categories.length > 0 && activeCatId === '') {
      setActiveCatId(categories[0].id)
    }
  }, [categories, activeCatId])

  useEffect(() => {
    if (categories.length === 0 || shouldRedirectToList) {
      return
    }

    const timer = setTimeout(() => {
      const query = Taro.createSelectorQuery()
      query.select('.category-page__browse-right').boundingClientRect()
      query.exec((res: Array<{ top?: number } | null>) => {
        const containerTop = res[0]?.top ?? 0
        const query2 = Taro.createSelectorQuery()
        categories.forEach((cat) => {
          query2.select(`#cat-section-${cat.id}`).boundingClientRect()
        })
        query2.exec((results: Array<{ top?: number } | null>) => {
          categories.forEach((cat, index) => {
            const top = (results[index]?.top ?? 0) - containerTop
            sectionOffsetsRef.current.set(cat.id, top)
          })
        })
      })
    }, 300)

    return () => clearTimeout(timer)
  }, [categories, shouldRedirectToList])

  const openProductList = useCallback((params?: { categoryId?: string }) => {
    void Taro.navigateTo({
      url: buildProductListUrl({
        categoryId: params?.categoryId,
        sort: 'latest',
      }),
    })
  }, [])

  const handleLeftTap = useCallback((catId: string) => {
    setActiveCatId(catId)
    isClickScrollingRef.current = true
    setRightScrollId(`cat-section-${catId}`)
    setTimeout(() => {
      isClickScrollingRef.current = false
    }, 500)
  }, [])

  const handleRightScroll = useCallback(
    (event: { detail?: { scrollTop?: number } }) => {
      if (isClickScrollingRef.current) {
        return
      }

      const scrollTop = event.detail?.scrollTop ?? 0
      let newActiveId = categories[0]?.id ?? ''

      for (const cat of categories) {
        const offset = sectionOffsetsRef.current.get(cat.id) ?? 0
        if (offset <= scrollTop + 20) {
          newActiveId = cat.id
        }
      }

      if (newActiveId !== activeCatId) {
        setActiveCatId(newActiveId)
        setLeftScrollId(`left-item-${newActiveId}`)
      }
    },
    [activeCatId, categories],
  )

  const renderSubCategoriesFor = (cat: Category) => {
    const children = cat.children ?? []
    if (children.length === 0) {
      return <EmptyState title='暂无子分类' />
    }

    const hasSubGroups = children.some((child) => (child.children ?? []).length > 0)
    if (hasSubGroups) {
      return (
        <>
          {children.map((group) => (
            <View key={group.id} className='category-page__section'>
              <Text className='category-page__section-title'>{group.name}</Text>
              <View className='category-page__sub-grid'>
                {(group.children ?? []).map((item) => (
                  <SubItem key={item.id} item={item} onTap={(id) => openProductList({ categoryId: id })} />
                ))}
              </View>
            </View>
          ))}
        </>
      )
    }

    return (
      <View className='category-page__section'>
        <View className='category-page__sub-grid'>
          {children.map((item) => (
            <SubItem key={item.id} item={item} onTap={(id) => openProductList({ categoryId: id })} />
          ))}
        </View>
      </View>
    )
  }

  if (shouldRedirectToList) {
    return <View className='category-page' />
  }

  return (
    <View className='category-page'>
      <View className='category-page__topbar'>
        <View className='category-page__search-entry' onClick={() => openProductList()}>
          <View className='category-page__search-icon'>
            <Search width={16} height={16} />
          </View>
          <Text className='category-page__search-placeholder'>搜索商品 / 查看热卖</Text>
        </View>
      </View>

      {categories.length > 0 ? (
        <View className='category-page__body'>
          <ScrollView scrollY className='category-page__sidebar' scrollIntoView={leftScrollId}>
            {categories.map((cat) => (
              <View
                key={cat.id}
                id={`left-item-${cat.id}`}
                className={`category-page__sidebar-item${activeCatId === cat.id ? ' category-page__sidebar-item--active' : ''}`}
                onClick={() => handleLeftTap(cat.id)}
              >
                <Text className='category-page__sidebar-label'>{cat.name}</Text>
              </View>
            ))}
          </ScrollView>

          <ScrollView
            scrollY
            className='category-page__browse-right'
            scrollIntoView={rightScrollId}
            scrollWithAnimation
            onScroll={handleRightScroll}
          >
            {categories.map((cat) => (
              <View key={cat.id} id={`cat-section-${cat.id}`}>
                <Text className='category-page__cat-header'>{cat.name}</Text>
                {renderSubCategoriesFor(cat)}
              </View>
            ))}
          </ScrollView>
        </View>
      ) : (
        <EmptyState title='暂无分类' description='请稍后再试' />
      )}
    </View>
  )
}
