import { Text, View } from '@tarojs/components'
import type { CartItem } from '@/types/biz'
import { Button, Card, Checkbox, InputNumber, Tag } from '@/ui/nutui'
import { Del } from '@/ui/icons'
import { formatCartSkuAttrs, formatCartUnavailableReason } from '@/utils/cart'
import { formatPrice } from '@/utils/price'
import './index.scss'

interface CartListItemProps {
  item: CartItem
  checked: boolean
  editMode: boolean
  busy?: boolean
  displayQty?: number
  maxQty?: number
  onToggle: (checked: boolean) => void
  onQtyChange: (qty: number) => void
  onDelete: () => void
}

export default function CartListItem({
  item,
  checked,
  editMode,
  busy = false,
  displayQty,
  maxQty,
  onToggle,
  onQtyChange,
  onDelete,
}: CartListItemProps) {
  const outOfStock = item.is_available && item.available_stock === 0
  const disabled = !item.is_available || outOfStock
  const hasOriginalPrice =
    item.snapshot_price_cents > 0 &&
    item.snapshot_price_cents !== item.current_price_cents
  const coverImage = item.product?.main_image || item.sku_image
  const productTitle = item.product?.title || item.product_title
  const productSubtitle = item.product?.subtitle?.trim()
  const currentPrice = formatPrice(item.current_price_cents)
  const originalPrice = hasOriginalPrice ? formatPrice(item.snapshot_price_cents) : null
  const specText = formatCartSkuAttrs(item.sku_attrs)
  const effectiveQty = displayQty ?? item.qty
  const effectiveMaxQty = Math.max(maxQty ?? item.available_stock ?? 999, 1)
  const lowStock = !disabled && item.available_stock > 0 && item.available_stock <= 5

  return (
    <View className={`cart-list-item${disabled ? ' cart-list-item--disabled' : ''}`}>
      <View className='cart-list-item__selector'>
        <Checkbox
          checked={checked}
          disabled={disabled || busy}
          onChange={onToggle}
        />
      </View>
      <View className='cart-list-item__content'>
        {editMode ? (
          <Button
            className='cart-list-item__delete'
            fill='none'
            disabled={busy}
            onClick={onDelete}
          >
            <Del size={16} />
          </Button>
        ) : null}
        <Card
          className='cart-list-item__card'
          src={coverImage}
          title={productTitle}
          tag={disabled ? <Tag type='danger'>已失效</Tag> : null}
          description={
            <View className='cart-list-item__description'>
              {productSubtitle ? (
                <Text className='cart-list-item__subtitle'>{productSubtitle}</Text>
              ) : null}
              {specText ? <Text className='cart-list-item__spec'>{specText}</Text> : null}
              {!item.is_available ? (
                <Text className='cart-list-item__status'>
                  {formatCartUnavailableReason(item.unavailable_reason)}
                </Text>
              ) : outOfStock ? (
                <Text className='cart-list-item__status'>暂时缺货</Text>
              ) : lowStock ? (
                <Text className='cart-list-item__status'>库存仅剩 {item.available_stock} 件</Text>
              ) : null}
              <View className='cart-list-item__summary'>
                <View className='cart-list-item__price-block'>
                  <Text className='cart-list-item__price-symbol'>¥</Text>
                  <Text className='cart-list-item__price-value'>{currentPrice}</Text>
                  {originalPrice ? (
                    <Text className='cart-list-item__price-origin'>¥{originalPrice}</Text>
                  ) : null}
                </View>
                {editMode ? (
                  <View className='cart-list-item__extra-placeholder' />
                ) : (
                  <View className='cart-list-item__extra'>
                    <InputNumber
                      value={effectiveQty}
                      min={1}
                      max={effectiveMaxQty}
                      disabled={disabled || busy}
                      size='small'
                      onChange={(value: string | number) => {
                        const nextQty = Number(value)
                        if (!Number.isNaN(nextQty) && nextQty > 0 && nextQty !== effectiveQty) {
                          onQtyChange(nextQty)
                        }
                      }}
                    />
                  </View>
                )}
              </View>
            </View>
          }
        />
      </View>
    </View>
  )
}
