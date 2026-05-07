import { Image, RichText, Text, View } from '@tarojs/components'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import Taro from '@tarojs/taro'
import { useEffect, useMemo, useRef, useState } from 'react'
import SkuPicker from '@/components/SkuPicker'
import { useShare } from '@/hooks/useShare'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { addToCart } from '@/services/cart'
import { getProductDetail } from '@/services/product'
import { addFavorite, removeFavorite } from '@/services/user'
import type { ProductDetail, Sku } from '@/types/biz'
import { ArrowLeft, ArrowRight, Cart, HeartFill, Location, Service, Share } from '@/ui/icons'
import { Button, SafeArea, Skeleton, Swiper, SwiperItem } from '@/ui/nutui'
import { formatPrice, formatSales } from '@/utils/price'
import './index.scss'

const FALLBACK_ID = '101'

const serviceTags = [
  '平台严选',
  '售后支持',
  '极速发货',
  '商品保障',
]

export default function ProductDetailPage() {
  const id = Taro.getCurrentInstance().router?.params?.id ?? FALLBACK_ID
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const refreshCount = useCartStore((s) => s.refreshCount)
  const queryClient = useQueryClient()

  const [skuPickerMode, setSkuPickerMode] = useState<'cart' | 'buy' | null>(null)
  const [isFavorite, setIsFavorite] = useState(false)
  const [favoriteInited, setFavoriteInited] = useState(false)
  const [activeImage, setActiveImage] = useState(0)

  const detailQuery = useQuery({
    queryKey: ['product', id],
    queryFn: () => getProductDetail(id),
  })

  const product = detailQuery.data

  useShare({
    title: product?.title ?? '商品详情',
    path: `/pages/product/detail/index?id=${id}`,
  })

  if (product && !favoriteInited) {
    setIsFavorite(product.is_favorite ?? false)
    setFavoriteInited(true)
  }

  const richContent = useMemo(
    () => normalizeDetailHtml(product?.detail_html || ''),
    [product?.detail_html],
  )
  const infoCards = useMemo(
    () => (product ? buildInfoCards(product) : []),
    [product],
  )

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

  function handleBack() {
    const pages = Taro.getCurrentPages()
    if (pages.length > 1) {
      void Taro.navigateBack()
      return
    }
    void Taro.switchTab({ url: '/pages/home/index' })
  }

  function handleShareClick() {
    void Taro.showToast({ title: '请使用系统分享能力', icon: 'none' })
  }

  function handleToggleFavorite() {
    if (!isLoggedIn) {
      void Taro.navigateTo({
        url: `/pages/auth/login/index?redirect=${encodeURIComponent(`/pages/product/detail/index?id=${id}`)}`,
      })
      return
    }
    if (isFavorite) {
      setIsFavorite(false)
      removeFavorite(id).catch(() => setIsFavorite(true))
    } else {
      setIsFavorite(true)
      addFavorite(id).catch(() => setIsFavorite(false))
    }
  }

  function handleSkuConfirm(params: { skuId: string; qty: number; mode: 'cart' | 'buy' }) {
    if (params.mode === 'cart') {
      addCartMutation.mutate({ skuId: params.skuId, qty: params.qty })
      return
    }

    Taro.setStorageSync('checkout-direct-sku', { sku_id: params.skuId, qty: params.qty })
    void Taro.navigateTo({ url: '/pages/order/confirm/index' })
  }

  function handleAddToCart() {
    if (!isLoggedIn) {
      void Taro.navigateTo({
        url: `/pages/auth/login/index?redirect=${encodeURIComponent(`/pages/product/detail/index?id=${id}`)}`,
      })
      return
    }
    setSkuPickerMode('cart')
  }

  function handleBuyNow() {
    if (!isLoggedIn) {
      void Taro.navigateTo({
        url: `/pages/auth/login/index?redirect=${encodeURIComponent(`/pages/product/detail/index?id=${id}`)}`,
      })
      return
    }
    setSkuPickerMode('buy')
  }

  if (detailQuery.isLoading) {
    return (
      <View className='detail-page detail-page--loading'>
        <Skeleton animated rows={10} />
      </View>
    )
  }

  if (detailQuery.isError || (!detailQuery.isLoading && !product)) {
    return <ProductNotFound onBack={handleBack} />
  }

  const images =
    product.images && product.images.length > 0 ? product.images : [product.main_image]
  const selectedSkuDesc = getSelectedSkuDesc(product)
  const summaryLine = getSummaryLine(product)
  return (
    <View className='detail-page'>
      <View className='detail-page__nav'>
        <View className='detail-page__nav-left' onClick={handleBack}>
          <ArrowLeft width={18} height={18} />
        </View>
        <Text className='detail-page__nav-title'>商品详情</Text>
        <View className='detail-page__nav-actions'>
          <View className='detail-page__nav-action' onClick={handleShareClick}>
            <Share width={16} height={16} />
          </View>
          <View
            className='detail-page__nav-action'
            onClick={() => void Taro.switchTab({ url: '/pages/cart/index' })}
          >
            <Cart width={16} height={16} />
          </View>
        </View>
      </View>

      <View className='detail-page__scroll'>
        <View className='detail-page__hero'>
          <Swiper
            className='detail-page__swiper'
            height='320px'
            circular
            autoplay
            interval={4200}
            duration={320}
            onChange={(event) => setActiveImage(event.detail.current)}
          >
            {images.map((src, index) => (
              <SwiperItem key={`${src}-${index}`} className='detail-page__swiper-item'>
                <Image className='detail-page__swiper-image' src={src} mode='aspectFill' />
              </SwiperItem>
            ))}
          </Swiper>
          {images.length > 1 ? (
            <View className='detail-page__hero-dots'>
              {images.map((src, index) => (
                <View
                  key={`${src}-${index}-dot`}
                  className={`detail-page__hero-dot${index === activeImage ? ' detail-page__hero-dot--active' : ''}`}
                />
              ))}
            </View>
          ) : null}
        </View>

        <View className='detail-page__summary'>
          <View className='detail-page__summary-main'>
            <View className='detail-page__summary-top'>
              <View className='detail-page__summary-copy'>
                <Text className='detail-page__title'>{product.title}</Text>
                {summaryLine ? (
                  <Text className='detail-page__subtitle'>{summaryLine}</Text>
                ) : null}
              </View>
            </View>

            <View className='detail-page__summary-bottom'>
              <View className='detail-page__price-block'>
                <View className='detail-page__price-row'>
                  <Text className='detail-page__price'>¥{formatPrice(product.price_min_cents)}</Text>
                  {product.price_max_cents > product.price_min_cents ? (
                    <Text className='detail-page__price-range'>起</Text>
                  ) : null}
                </View>
              </View>

              <View className='detail-page__summary-side'>
                <Text className='detail-page__summary-sales'>
                  已售 {formatSales(product.sales ?? 0)}
                </Text>
                <View className='detail-page__summary-actions'>
                  <View className='detail-page__summary-action' onClick={handleToggleFavorite}>
                    <HeartFill width={18} height={18} color={isFavorite ? '#d38b43' : '#8d8d8d'} />
                    <Text className='detail-page__summary-action-text'>
                      {isFavorite ? '已收藏' : '收藏'}
                    </Text>
                  </View>
                  <View className='detail-page__summary-action' onClick={handleShareClick}>
                    <Share width={18} height={18} color='#8d8d8d' />
                    <Text className='detail-page__summary-action-text'>分享</Text>
                  </View>
                </View>
              </View>
            </View>
          </View>

          <View className='detail-page__service-row'>
            {serviceTags.map((item) => (
              <View key={item} className='detail-page__service-item'>
                <View className='detail-page__service-icon'>
                  <Service width={12} height={12} />
                </View>
                <Text className='detail-page__service-text'>{item}</Text>
              </View>
            ))}
          </View>
        </View>

        <View className='detail-page__panel-group'>
          <View className='detail-page__info-row' onClick={() => setSkuPickerMode('cart')}>
            <Text className='detail-page__info-label'>已选</Text>
            <View className='detail-page__info-content'>
              <Text className='detail-page__info-value'>{selectedSkuDesc}, 1件</Text>
              <ArrowRight width={14} height={14} />
            </View>
          </View>

          <View className='detail-page__info-row' onClick={() => setSkuPickerMode('cart')}>
            <Text className='detail-page__info-label'>规格</Text>
            <View className='detail-page__info-content'>
              <Text className='detail-page__info-value'>
                {product.specs?.map((item) => item.name).join(' / ') || '默认规格'}
              </Text>
              <ArrowRight width={14} height={14} />
            </View>
          </View>

          <View className='detail-page__info-row'>
            <View className='detail-page__info-logistics'>
              <View className='detail-page__info-logistics-icon'>
                <Location width={16} height={16} />
              </View>
              <View className='detail-page__info-logistics-copy'>
                <Text className='detail-page__info-logistics-label'>配送至</Text>
                <Text className='detail-page__info-logistics-desc'>全国大部分地区可配送，预计 1-3 天发出</Text>
              </View>
            </View>
            <Text className='detail-page__info-value'>全国可送</Text>
          </View>
        </View>

        <View className='detail-page__section'>
          <View className='detail-page__section-head'>
            <Text className='detail-page__section-title'>商品信息</Text>
          </View>
          <View className='detail-page__info-cards'>
            {infoCards.map((card) => (
              <View key={card.label} className='detail-page__info-card'>
                <Text className='detail-page__info-card-label'>{card.label}</Text>
                <Text className='detail-page__info-card-value'>{card.value}</Text>
              </View>
            ))}
          </View>
        </View>

        <View className='detail-page__section'>
          <View className='detail-page__section-head'>
            <Text className='detail-page__section-title'>用户评价</Text>
            <Text className='detail-page__section-side'>暂无评价</Text>
          </View>
          <View className='detail-page__review-empty'>
            <Text className='detail-page__review-empty-text'>当前商品还没有公开评价，购买后可留下真实体验。</Text>
          </View>
        </View>

        <View className='detail-page__section'>
          <View className='detail-page__section-head'>
            <Text className='detail-page__section-title'>商品介绍</Text>
          </View>
          <View className='detail-page__detail-content'>
            {richContent ? (
              <RichText nodes={richContent} />
            ) : (
              <Text className='detail-page__detail-empty'>暂无图文详情</Text>
            )}
          </View>
        </View>

        <View className='detail-page__bottom-spacer' />
      </View>

      {skuPickerMode !== null ? (
        <SkuPicker
          visible={skuPickerMode !== null}
          product={product as ProductDetail}
          mode={skuPickerMode}
          onClose={() => setSkuPickerMode(null)}
          onConfirm={handleSkuConfirm}
        />
      ) : null}

      <View className='detail-page__bar'>
        <View className='detail-page__bar-inner'>
          <View className='detail-page__bar-side' onClick={() => void Taro.switchTab({ url: '/pages/cart/index' })}>
            <Cart width={20} height={20} />
            <Text className='detail-page__bar-side-text'>购物车</Text>
          </View>
          <View className='detail-page__bar-side' onClick={handleToggleFavorite}>
            <HeartFill width={20} height={20} color={isFavorite ? '#d38b43' : '#8d8d8d'} />
            <Text className='detail-page__bar-side-text'>{isFavorite ? '已收藏' : '收藏'}</Text>
          </View>
          <View className='detail-page__bar-actions'>
            <Button className='detail-page__bar-cart' type='primary' onClick={handleAddToCart}>
              加入购物车
            </Button>
            <Button className='detail-page__bar-buy' type='primary' onClick={handleBuyNow}>
              立即购买
            </Button>
          </View>
        </View>
        <SafeArea position='bottom' />
      </View>
    </View>
  )
}

function getSelectedSkuDesc(product: ProductDetail): string {
  const skus = product.skus ?? []
  if (skus.length === 0) return '默认款'
  if (skus.length === 1) {
    const attrs = Object.values(skus[0].attrs)
    return attrs.length > 0 ? attrs.join(' / ') : '默认款'
  }
  return '请选择规格'
}

function getTotalStock(product: ProductDetail): number | null {
  const skus: Sku[] = product.skus ?? []
  if (skus.length === 0) return null
  return skus.reduce((sum, sku) => sum + sku.stock, 0)
}

function getSummaryLine(product: ProductDetail) {
  if (product.tags && product.tags.length > 0) {
    return product.tags.join('  ')
  }

  const raw = (product.subtitle || '').trim()
  if (!raw) {
    return ''
  }

  return raw
    .split(/[·|｜]/)
    .map((item) => item.trim())
    .filter((item) => item && !item.includes('已售'))
    .join('  ')
}

function buildInfoCards(product: ProductDetail) {
  const totalStock = getTotalStock(product)
  const priceRange = product.price_max_cents > product.price_min_cents
    ? `¥${formatPrice(product.price_min_cents)}-${formatPrice(product.price_max_cents)}`
    : `¥${formatPrice(product.price_min_cents)}`

  return [
    {
      label: '价格区间',
      value: priceRange,
    },
    {
      label: '在售状态',
      value: product.status === 'onsale' ? '在售' : '暂不可售',
    },
    {
      label: '总库存',
      value: totalStock != null ? `${totalStock}` : '以实际为准',
    },
    {
      label: '销量',
      value: `${formatSales(product.sales ?? 0)}`,
    },
  ]
}

function normalizeDetailHtml(html: string) {
  if (!html.trim()) {
    return ''
  }

  const wrappers = 'max-width:100%;box-sizing:border-box;word-break:break-word;overflow-wrap:anywhere;'
  const blockSpacing = 'margin:0 0 16px;'
  const imgStyle = 'max-width:100%!important;width:100%!important;height:auto!important;display:block;margin:0 auto 16px;object-fit:contain;'
  const videoStyle = 'max-width:100%!important;width:100%!important;height:auto!important;display:block;margin:0 auto 16px;'
  const tableStyle = 'display:block;max-width:100%!important;width:100%!important;overflow-x:auto;box-sizing:border-box;border-collapse:collapse;'
  const cellStyle = 'border:1px solid #f0e4d3;padding:8px;word-break:break-word;'
  const iframeStyle = 'max-width:100%!important;width:100%!important;min-height:220px;display:block;'

  return `<div style="${wrappers}overflow:hidden;">${
    injectInlineStyle(
      injectInlineStyle(
        injectInlineStyle(
          injectInlineStyle(
            injectInlineStyle(
              injectInlineStyle(
                html,
                /<(p|div|section|article|h1|h2|h3|h4|h5|h6|ul|ol|li|blockquote)([^>]*)>/gi,
                `${wrappers}${blockSpacing}`,
              ),
              /<(iframe)([^>]*)>/gi,
              iframeStyle,
            ),
            /<(td|th)([^>]*)>/gi,
            cellStyle,
          ),
          /<(table)([^>]*)>/gi,
          tableStyle,
        ),
        /<(video)([^>]*)>/gi,
        videoStyle,
      ),
      /<(img)([^>]*)>/gi,
      imgStyle,
    )
  }</div>`
}

function injectInlineStyle(source: string, pattern: RegExp, style: string) {
  return source.replace(pattern, (_match, tagName: string, attrs: string) => {
    const styleMatch = attrs.match(/\sstyle=(["'])(.*?)\1/i)
    if (styleMatch) {
      const mergedStyle = `${styleMatch[2]};${style}`
      return `<${tagName}${attrs.replace(styleMatch[0], ` style="${mergedStyle}"`)}>`
    }
    return `<${tagName}${attrs} style="${style}">`
  })
}

function ProductNotFound({ onBack }: { onBack: () => void }) {
  const [countdown, setCountdown] = useState(3)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    timerRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timerRef.current!)
          void Taro.switchTab({ url: '/pages/home/index' })
          return 0
        }
        return prev - 1
      })
    }, 1000)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [])

  return (
    <View className='detail-page detail-page--not-found'>
      <View className='detail-page__not-found-body'>
        <Text className='detail-page__not-found-emoji'>🔍</Text>
        <Text className='detail-page__not-found-title'>商品不存在</Text>
        <Text className='detail-page__not-found-desc'>该商品已下架或链接有误</Text>
        <Text className='detail-page__not-found-countdown'>{countdown}s 后自动返回首页</Text>
        <Button
          className='detail-page__not-found-btn'
          type='primary'
          onClick={() => {
            if (timerRef.current) clearInterval(timerRef.current)
            void Taro.switchTab({ url: '/pages/home/index' })
          }}
        >
          立即返回首页
        </Button>
        <Button
          className='detail-page__not-found-back'
          fill='none'
          onClick={() => {
            if (timerRef.current) clearInterval(timerRef.current)
            onBack()
          }}
        >
          返回上一页
        </Button>
      </View>
    </View>
  )
}
