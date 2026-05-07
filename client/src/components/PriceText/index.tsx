import { Text, View } from '@tarojs/components'
import { formatPrice } from '@/utils/price'
import { Price } from '@/ui/nutui'
import './index.scss'

interface PriceTextProps {
  cents: number
  marketCents?: number
}

export default function PriceText({ cents, marketCents }: PriceTextProps) {
  return (
    <View className='price-text'>
      <Price className='price-text__main' price={formatPrice(cents)} color='var(--ds-color-text-price)' size='large' thousands />
      {marketCents ? <Text className='price-text__market'>¥{formatPrice(marketCents)}</Text> : null}
    </View>
  )
}
