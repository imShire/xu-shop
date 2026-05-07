import { View, Text, Image } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery } from '@tanstack/react-query'
import { getProductDetail } from '@/services/product'
import { Button, Skeleton, SafeArea } from '@/ui/nutui'
import PriceText from '@/components/PriceText'
import { demoProducts } from '@/utils/demo'
import './index.scss'

export default function SharePosterPage() {
  const params = Taro.getCurrentInstance().router?.params
  const productId = params?.product_id ?? demoProducts[0].id

  const { data: product, isLoading } = useQuery({
    queryKey: ['product-detail', productId],
    queryFn: () => getProductDetail(productId),
  })

  const item = product ?? demoProducts[0]

  return (
    <View className='page-shell share-poster-page'>
      {isLoading ? (
        <Skeleton animated rows={6} />
      ) : (
        <>
          <View className='share-poster-page__card page-card'>
            <Image className='share-poster-page__image' src={item.main_image} mode='aspectFill' />
            <View className='share-poster-page__content'>
              <Text className='share-poster-page__shop-name'>徐记小铺</Text>
              <Text className='share-poster-page__title'>{item.title}</Text>
              {item.subtitle && (
                <Text className='share-poster-page__subtitle'>{item.subtitle}</Text>
              )}
              <PriceText cents={item.price_min_cents} />
              <View className='share-poster-page__qrcode-area'>
                <View className='share-poster-page__qrcode-placeholder'>
                  <Text className='share-poster-page__qrcode-text'>扫码查看</Text>
                </View>
                <View className='share-poster-page__qrcode-hint'>
                  <Text className='share-poster-page__qrcode-hint-text'>微信扫一扫</Text>
                  <Text className='share-poster-page__qrcode-hint-text'>进店选购</Text>
                </View>
              </View>
            </View>
          </View>

          <View className='share-poster-page__actions'>
            <Button
              type='primary'
              block
              onClick={() => void Taro.showToast({ title: '长按图片保存', icon: 'none' })}
            >
              保存海报
            </Button>
          </View>
        </>
      )}
      <SafeArea position='bottom' />
    </View>
  )
}
