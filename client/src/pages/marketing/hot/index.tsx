import { ScrollView, Text, View } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import ProductCard from '@/components/ProductCard'
import { getHotProducts } from '@/services/product'
import type { Product } from '@/types/biz'
import { Skeleton } from '@/ui/nutui'
import './index.scss'

const HOT_LIMIT = 50

export default function HotPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['hot-products', HOT_LIMIT],
    queryFn: () => getHotProducts(HOT_LIMIT),
    staleTime: 2 * 60 * 1000,
  })

  const list: Product[] = data?.list ?? []

  return (
    <ScrollView className='hot-page' scrollY>
      <View className='hot-page__header'>
        <Text className='hot-page__title'>热门商品</Text>
        <Text className='hot-page__subtitle'>大家都在买的好物</Text>
      </View>

      {isLoading && list.length === 0 ? (
        <View className='hot-page__grid'>
          {Array.from({ length: 4 }).map((_, index) => (
            <View key={index} className='hot-page__card hot-page__card--skeleton'>
              <Skeleton animated rows={5} />
            </View>
          ))}
        </View>
      ) : list.length === 0 ? (
        <View className='hot-page__empty'>暂无数据</View>
      ) : (
        <View className='hot-page__grid'>
          {list.map((product) => {
            const tags = product.tags?.filter(Boolean).slice(0, 2) ?? []
            return (
              <View key={product.id} className='hot-page__card'>
                <ProductCard
                  mode='vertical'
                  product={{
                    id: product.id,
                    title: product.title,
                    main_image: product.main_image,
                    price_cents: product.price_min_cents,
                  }}
                  tags={tags}
                />
              </View>
            )
          })}
        </View>
      )}
    </ScrollView>
  )
}
