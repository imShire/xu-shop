import { Image, ScrollView, Swiper, SwiperItem, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useRef, useState } from 'react'
import ProductCard from '@/components/ProductCard'
import SkuPicker from '@/components/SkuPicker'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { useShare } from '@/hooks/useShare'
import { getBanners } from '@/services/banner'
import { addToCart } from '@/services/cart'
import { getNavIcons, type NavIcon } from '@/services/nav-icon'
import { getCategories, getProductDetail, getProducts } from '@/services/product'
import { useCartStore } from '@/stores/cart'
import type { Category, Product, ProductDetail, ProductListPage } from '@/types/biz'
import { buildProductListUrl, mapCategoryLinkToProductListUrl } from '@/utils/product-list'
import { ArrowRight, Cart, Search } from '@/ui/icons'
import { Skeleton } from '@/ui/nutui'
import './index.scss'

const PAGE_SIZE = 8
const FEATURE_LIMIT = 8
const HOME_REDIRECT = '/pages/home/index'

type HomeEntry = {
  id: string
  title: string
  imageUrl: string
  linkUrl: string
}

function navigateByLink(linkUrl: string) {
  if (!linkUrl) return

  const productListUrl = mapCategoryLinkToProductListUrl(linkUrl)
  if (productListUrl) {
    void Taro.navigateTo({ url: productListUrl })
    return
  }

  if (linkUrl.startsWith('/pages')) {
    const tabUrl = linkUrl.split('?')[0]

    if (linkUrl.includes('/pages/home/index')) {
      void Taro.switchTab({ url: '/pages/home/index' })
      return
    }

    if (
      linkUrl.includes('/pages/cart/index') ||
      linkUrl.includes('/pages/category/index') ||
      linkUrl.includes('/pages/user/index/index')
    ) {
      void Taro.switchTab({ url: tabUrl })
      return
    }

    void Taro.navigateTo({ url: linkUrl })
    return
  }

  void Taro.navigateTo({ url: `/pages/webview/index?url=${encodeURIComponent(linkUrl)}` })
}

function buildFeatureEntries(navIcons: NavIcon[]) {
  return navIcons.slice(0, FEATURE_LIMIT).map((item) => ({
    id: `nav-${item.id}`,
    title: item.title,
    imageUrl: item.icon_url,
    linkUrl: item.link_url,
  }))
}

function buildProductMeta(product: Product, categories: Category[]) {
  const labels = product.tags?.filter(Boolean).slice(0, 2) ?? []

  if (labels.length > 0) {
    return labels
  }

  if (product.subtitle) {
    return [product.subtitle]
  }

  if (product.category_id) {
    const category = categories.find((item) => item.id === product.category_id)
    if (category) {
      return [category.name]
    }
  }

  return []
}

export default function HomePage() {
  const loadingRef = useRef(false)
  const guard = useAuthGuard()
  const queryClient = useQueryClient()
  const refreshCount = useCartStore((state) => state.refreshCount)

  const [activeBanner, setActiveBanner] = useState(0)
  const [quickAddId, setQuickAddId] = useState('')
  const [skuPickerProduct, setSkuPickerProduct] = useState<ProductDetail | null>(null)

  useShare({
    title: '徐记小铺精选好物',
    path: '/pages/home/index',
  })

  const { data: banners = [], isLoading: isBannerLoading } = useQuery({
    queryKey: ['banners'],
    queryFn: getBanners,
  })

  const { data: navIcons = [], isLoading: isNavLoading } = useQuery({
    queryKey: ['nav-icons'],
    queryFn: getNavIcons,
    staleTime: 5 * 60 * 1000,
  })

  const { data: categories = [] } = useQuery({
    queryKey: ['categories'],
    queryFn: getCategories,
    staleTime: 5 * 60 * 1000,
  })

  const {
    data: productsData,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isLoading: isProductsLoading,
  } = useInfiniteQuery({
    queryKey: ['products', 'home-latest'],
    queryFn: ({ pageParam = 1 }) =>
      getProducts({ sort: 'latest', page: pageParam as number, page_size: PAGE_SIZE }),
    getNextPageParam: (lastPage) => {
      const loaded = lastPage.page * lastPage.page_size
      return loaded >= lastPage.total ? undefined : lastPage.page + 1
    },
    initialPageParam: 1,
  })

  const addCartMutation = useMutation({
    mutationFn: (params: { skuId: string; qty: number }) =>
      addToCart({ sku_id: params.skuId, qty: params.qty }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['cart'] })
      void refreshCount()
      void Taro.showToast({ title: '已加入购物车', icon: 'success' })
    },
    onError: () => {
      void Taro.showToast({ title: '加购失败，请重试', icon: 'none' })
    },
  })

  const allProducts: Product[] =
    productsData?.pages.flatMap((page: ProductListPage) => page.list) ?? []
  const displayBanners = banners.slice(0, 5)
  const featureEntries = buildFeatureEntries(navIcons)

  const handleScrollToLower = useCallback(() => {
    if (loadingRef.current || !hasNextPage || isFetchingNextPage) return

    loadingRef.current = true
    void fetchNextPage().finally(() => {
      loadingRef.current = false
    })
  }, [fetchNextPage, hasNextPage, isFetchingNextPage])

  const handleFeatureClick = (item: HomeEntry) => {
    navigateByLink(item.linkUrl)
  }

  const handleQuickAdd = (productId: string) => {
    void guard(() => {
      setQuickAddId(productId)
      void queryClient
        .fetchQuery({
          queryKey: ['product', productId],
          queryFn: () => getProductDetail(productId),
        })
        .then((detail) => {
          if ((detail.skus ?? []).length === 1) {
            addCartMutation.mutate({ skuId: detail.skus[0].id, qty: 1 })
            return
          }

          if ((detail.skus ?? []).length === 0) {
            void Taro.navigateTo({ url: `/pages/product/detail/index?id=${productId}` })
            return
          }

          setSkuPickerProduct(detail)
        })
        .catch(() => {
          void Taro.showToast({ title: '商品信息加载失败', icon: 'none' })
        })
        .finally(() => {
          setQuickAddId('')
        })
    }, HOME_REDIRECT)
  }

  const handleSkuConfirm = (params: { skuId: string; qty: number }) => {
    addCartMutation.mutate({ skuId: params.skuId, qty: params.qty })
  }

  return (
    <ScrollView
      className='home-page'
      scrollY
      onScrollToLower={handleScrollToLower}
      lowerThreshold={120}
    >
      <View
        className='home-page__search-shell'
        onClick={() => void Taro.navigateTo({ url: buildProductListUrl() })}
      >
        <View className='home-page__search-field'>
          <View className='home-page__search-icon'>
            <Search width={18} height={18} />
          </View>
          <Text className='home-page__search-placeholder'>搜索商品 / 品牌</Text>
        </View>
        <View className='home-page__search-button'>搜索</View>
      </View>

      {isBannerLoading && displayBanners.length === 0 ? (
        <View className='home-page__hero home-page__hero--skeleton'>
          <Skeleton animated rows={3} />
        </View>
      ) : displayBanners.length > 0 ? (
        <View className='home-page__hero'>
          <Swiper
            className='home-page__hero-swiper'
            circular
            autoplay
            interval={4200}
            duration={360}
            onChange={(event) => setActiveBanner(event.detail.current)}
          >
            {displayBanners.map((banner) => (
              <SwiperItem key={banner.id}>
                <View
                  className='home-page__hero-slide'
                  onClick={() => navigateByLink(banner.link_url)}
                >
                  <Image
                    className='home-page__hero-image'
                    src={banner.image_url}
                    mode='aspectFill'
                  />
                  {banner.title ? (
                    <View className='home-page__hero-copy'>
                      <Text className='home-page__hero-title'>{banner.title}</Text>
                    </View>
                  ) : null}
                </View>
              </SwiperItem>
            ))}
          </Swiper>

          {displayBanners.length > 1 && (
            <View className='home-page__hero-dots'>
              {displayBanners.map((banner, index) => (
                <View
                  key={banner.id}
                  className={`home-page__hero-dot${index === activeBanner ? ' home-page__hero-dot--active' : ''}`}
                />
              ))}
            </View>
          )}
        </View>
      ) : null}

      {isNavLoading && featureEntries.length === 0 ? (
        <View className='home-page__feature-panel'>
          <View className='home-page__feature-grid'>
            {Array.from({ length: FEATURE_LIMIT }).map((_, index) => (
              <View key={index} className='home-page__feature-item home-page__feature-item--loading'>
                <View className='home-page__feature-circle' />
                <View className='home-page__feature-text-skeleton' />
              </View>
            ))}
          </View>
        </View>
      ) : featureEntries.length > 0 ? (
        <View className='home-page__feature-panel'>
          <View className='home-page__feature-grid'>
            {featureEntries.map((item) => (
              <View
                key={item.id}
                className='home-page__feature-item'
                onClick={() => handleFeatureClick(item)}
              >
                <View className='home-page__feature-circle'>
                  <Image
                    className='home-page__feature-image'
                    src={item.imageUrl}
                    mode='aspectFill'
                  />
                </View>
                <Text className='home-page__feature-title'>{item.title}</Text>
              </View>
            ))}
          </View>
        </View>
      ) : null}

      <View className='home-page__section-head'>
        <Text className='home-page__section-title'>最新商品</Text>
        <View
          className='home-page__section-more'
          onClick={() => void Taro.navigateTo({ url: buildProductListUrl({ sort: 'latest' }) })}
        >
          <Text className='home-page__section-more-text'>查看更多</Text>
          <ArrowRight width={14} height={14} />
        </View>
      </View>

      {isProductsLoading && allProducts.length === 0 ? (
        <View className='home-page__product-grid'>
          {Array.from({ length: 4 }).map((_, index) => (
            <View key={index} className='home-page__product-card home-page__product-card--skeleton'>
              <Skeleton animated rows={5} />
            </View>
          ))}
        </View>
      ) : (
        <>
          <View className='home-page__product-grid'>
            {allProducts.map((product) => {
              const meta = buildProductMeta(product, categories)
              return (
                <View key={product.id} className='home-page__product-card'>
                  <ProductCard
                    mode='vertical'
                    product={{
                      id: product.id,
                      title: product.title,
                      main_image: product.main_image,
                      price_cents: product.price_min_cents,
                    }}
                    tags={meta}
                  />
                </View>
              )
            })}
          </View>

          {isFetchingNextPage && (
            <View className='home-page__loading-more'>
              <Skeleton animated rows={2} />
            </View>
          )}

          {!hasNextPage && allProducts.length > 0 && (
            <View className='home-page__end'>
              <Text className='home-page__end-text'>没有更多商品了</Text>
            </View>
          )}
        </>
      )}

      {skuPickerProduct ? (
        <SkuPicker
          visible={!!skuPickerProduct}
          product={skuPickerProduct}
          mode='cart'
          onClose={() => setSkuPickerProduct(null)}
          onConfirm={handleSkuConfirm}
        />
      ) : null}
    </ScrollView>
  )
}
