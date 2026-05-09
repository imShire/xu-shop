import { View, Text } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import ProductCard from '@/components/ProductCard'
import { getProducts } from '@/services/product'
import type { ProductListData } from '@/services/page-config'
import './ProductListModule.scss'

interface Props {
  data: ProductListData
}

export default function ProductListModule({ data }: Props) {
  const { title = '推荐商品', sort = 'latest', limit = 4 } = data

  const { data: result } = useQuery({
    queryKey: ['page-module-products', sort, limit],
    queryFn: () => getProducts({ sort, page: 1, page_size: limit }),
    staleTime: 2 * 60 * 1000,
  })

  const products = result?.list ?? []
  if (products.length === 0) return null

  return (
    <View className='product-list-module'>
      <View className='product-list-module__head'>
        <Text className='product-list-module__title'>{title}</Text>
      </View>
      <View className='product-list-module__grid'>
        {products.map((product) => (
          <View key={product.id} className='product-list-module__item'>
            <ProductCard
              mode='vertical'
              product={{
                id: product.id,
                title: product.title,
                main_image: product.main_image,
                price_cents: product.price_min_cents,
              }}
            />
          </View>
        ))}
      </View>
    </View>
  )
}
