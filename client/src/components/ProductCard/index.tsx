import { type ReactNode } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { Image } from '@/ui/nutui'
import { Cart } from '@/ui/icons'
import PriceText from '@/components/PriceText'
import './index.scss'

interface ProductCardProduct {
  id: string | number
  title: string
  subtitle?: string
  main_image: string
  price_cents: number
  market_price_cents?: number
  virtual_sales?: number
}

interface ProductCardProps {
  product: ProductCardProduct
  /** vertical: 竖版（一行2个，默认）; horizontal: 横版（一行1个） */
  mode?: 'vertical' | 'horizontal'
  /** 标签列表，最多展示前2个 */
  tags?: string[]
  /** 标题右侧自定义内容 slot */
  titleRight?: ReactNode
  /** 替换默认购物车按钮的 slot */
  cartButton?: ReactNode
  /** 点击购物车按钮时触发；未提供则跳转商品详情页选择规格 */
  onAddToCart?: () => void
  onClick?: () => void
}

export default function ProductCard({
  product,
  mode = 'vertical',
  tags,
  titleRight,
  cartButton,
  onAddToCart,
  onClick,
}: ProductCardProps) {
  const handleCardClick = () => {
    if (onClick) {
      onClick()
    } else {
      void Taro.navigateTo({ url: `/pages/product/detail/index?id=${product.id}` })
    }
  }

  const handleAddClick = (e: { stopPropagation: () => void }) => {
    e.stopPropagation()
    if (onAddToCart) {
      onAddToCart()
    } else {
      void Taro.navigateTo({ url: `/pages/product/detail/index?id=${product.id}` })
    }
  }

  const tagList = tags?.slice(0, 2) ?? []

  const defaultCartButton = (
    <View className='product-card__cart-round' onClick={handleAddClick}>
      <Cart width={16} height={16} color='#FFFFFF' />
    </View>
  )

  const cartNode = cartButton ?? defaultCartButton

  const titleRow = (
    <View className='product-card__title-row'>
      <Text className='product-card__title' numberOfLines={2}>
        {product.title}
      </Text>
      {titleRight ? <View className='product-card__title-right'>{titleRight}</View> : null}
    </View>
  )

  const tagsRow =
    tagList.length > 0 ? (
      <View className='product-card__tags'>
        {tagList.map((tag) => (
          <Text key={tag} className='product-card__tag'>
            {tag}
          </Text>
        ))}
      </View>
    ) : null

  const priceRow = (
    <View className='product-card__price-row'>
      <PriceText cents={product.price_cents} marketCents={product.market_price_cents} />
      <View className='product-card__cart'>{cartNode}</View>
    </View>
  )

  if (mode === 'horizontal') {
    return (
      <View className='product-card product-card--horizontal' onClick={handleCardClick}>
        <Image
          className='product-card__h-image'
          src={product.main_image}
          mode='aspectFill'
        />
        <View className='product-card__h-body'>
          {titleRow}
          {tagsRow}
          {priceRow}
        </View>
      </View>
    )
  }

  // vertical mode (default)
  return (
    <View className='product-card product-card--vertical' onClick={handleCardClick}>
      <Image
        className='product-card__v-image'
        src={product.main_image}
        mode='aspectFill'
      />
      <View className='product-card__v-body'>
        {titleRow}
        {tagsRow}
        {priceRow}
      </View>
    </View>
  )
}
