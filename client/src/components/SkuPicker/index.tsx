import { View, Text, Image } from '@tarojs/components'
import { useState, useEffect } from 'react'
import PriceText from '@/components/PriceText'
import type { ProductDetail, Sku } from '@/types/biz'
import { Button, InputNumber, Popup, Tag } from '@/ui/nutui'
import './index.scss'

interface SkuPickerProps {
  visible: boolean
  product: ProductDetail
  mode: 'cart' | 'buy'
  onClose: () => void
  onConfirm: (params: { skuId: string; qty: number; mode: 'cart' | 'buy' }) => void
}

export default function SkuPicker({
  visible,
  product,
  mode,
  onClose,
  onConfirm,
}: SkuPickerProps) {
  const skus = product.skus ?? []
  const [qty, setQty] = useState(1)
  const [selectedSkuId, setSelectedSkuId] = useState<string | null>(
    skus.length === 1 ? skus[0].id : null
  )

  useEffect(() => {
    if (visible) {
      setQty(1)
      setSelectedSkuId(skus.length === 1 ? skus[0].id : null)
    }
  }, [visible]) // eslint-disable-line react-hooks/exhaustive-deps

  const attrGroups = skus.reduce<Record<string, Set<string>>>((acc, sku) => {
    Object.entries(sku.attrs).forEach(([k, v]) => {
      if (!acc[k]) acc[k] = new Set()
      acc[k].add(v)
    })
    return acc
  }, {})

  const hasSku = skus.length > 0

  const selectedSku = skus.find((s: Sku) => s.id === selectedSkuId)
  const displayPrice = selectedSku?.price_cents ?? product.price_min_cents
  const displayMarket = selectedSku?.original_price_cents
  const displayImage = selectedSku?.image ?? product.main_image
  const maxQty = selectedSku?.stock ?? 999

  function handleConfirm() {
    if (hasSku && !selectedSkuId) {
      return
    }
    if (!selectedSkuId) return
    onConfirm({ skuId: selectedSkuId, qty, mode })
    onClose()
  }

  return (
    <Popup
      round
      visible={visible}
      position='bottom'
      closeable
      title='选择规格'
      onClose={onClose}
    >
      <View className='sku-picker'>
        <View className='sku-picker__header'>
          <Image className='sku-picker__image' src={displayImage} mode='aspectFill' />
          <View className='sku-picker__info'>
            <PriceText cents={displayPrice} marketCents={displayMarket} />
            <Text className='sku-picker__title'>{product.title}</Text>
            {selectedSku && (
              <Text className='sku-picker__selected'>
                已选：{Object.values(selectedSku.attrs).join(' / ')}
              </Text>
            )}
          </View>
        </View>

        {hasSku &&
          Object.entries(attrGroups).map(([attrName, values]) => (
            <View key={attrName} className='sku-picker__group'>
              <Text className='sku-picker__group-label'>{attrName}</Text>
              <View className='sku-picker__tags'>
                {Array.from(values).map((val) => {
                  const matchSku = skus.find((s: Sku) => s.attrs[attrName] === val)
                  const isSelected = matchSku ? matchSku.id === selectedSkuId : false
                  const outOfStock = matchSku ? matchSku.stock === 0 : false
                  return (
                    <Tag
                      key={val}
                      className={`sku-picker__tag ${outOfStock ? 'sku-picker__tag--disabled' : ''}`}
                      type={isSelected ? 'primary' : 'default'}
                      plain={!isSelected}
                      onClick={() => {
                        if (!outOfStock && matchSku) {
                          setSelectedSkuId(matchSku.id)
                          setQty(1)
                        }
                      }}
                    >
                      {val}
                    </Tag>
                  )
                })}
              </View>
            </View>
          ))}

        {!hasSku && (
          <View className='sku-picker__group'>
            <Text className='sku-picker__group-label'>规格</Text>
            <View className='sku-picker__tags'>
              <Tag type='primary'>默认款</Tag>
            </View>
          </View>
        )}

        <View className='sku-picker__qty'>
          <Text className='sku-picker__qty-label'>数量</Text>
          <InputNumber
            value={qty}
            min={1}
            max={maxQty}
            onChange={(v) => setQty(Number(v))}
          />
        </View>

        <View className='sku-picker__actions'>
          <Button
            type='primary'
            block
            disabled={hasSku && !selectedSkuId}
            onClick={handleConfirm}
          >
            {mode === 'cart' ? '加入购物车' : '立即购买'}
          </Button>
        </View>
      </View>
    </Popup>
  )
}

