import { View } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { useAuthStore } from '@/stores/auth'
import ProductCard from '@/components/ProductCard'
import { getFavorites, removeFavorite } from '@/services/user'
import { Button, Skeleton } from '@/ui/nutui'
import EmptyState from '@/components/EmptyState'
import './index.scss'

export default function FavoritePage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()

  useDidShow(() => {
    if (!isLoggedIn) {
      void ensureAuth(undefined, '/pages/user/favorite/index')
    }
  })

  const { data, isLoading } = useQuery({
    queryKey: ['favorites'],
    queryFn: () => getFavorites({ page: 1, page_size: 20 }),
    enabled: isLoggedIn,
  })

  const removeMutation = useMutation({
    mutationFn: (productId: string) => removeFavorite(productId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['favorites'] })
      void Taro.showToast({ title: '已取消收藏', icon: 'none' })
    },
    onError: () => {
      void Taro.showToast({ title: '操作失败，请重试', icon: 'none' })
    },
  })

  const items = data?.list ?? []

  if (!isLoggedIn) return <View className='page-shell' />

  return (
    <View className='page-shell favorite-page'>
      {isLoading ? (
        <View className='favorite-page__list'>
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} animated rows={4} />
          ))}
        </View>
      ) : items.length ? (
        <View className='favorite-page__list'>
          {items.map((item) => (
            <ProductCard
              key={item.id}
              mode='horizontal'
              product={{
                id: item.product_id,
                title: item.title,
                main_image: item.image,
                price_cents: item.price_cents,
                market_price_cents: item.market_price_cents,
              }}
              tags={['已收藏']}
              cartButton={
                <Button
                  size='small'
                  plain
                  onClick={(event: { stopPropagation: () => void }) => {
                    event.stopPropagation()
                    removeMutation.mutate(item.product_id)
                  }}
                >
                  取消收藏
                </Button>
              }
              onClick={() =>
                void Taro.navigateTo({
                  url: `/pages/product/detail/index?id=${item.product_id}`,
                })
              }
            />
          ))}
        </View>
      ) : (
        <EmptyState
          title='还没有收藏商品'
          description='逛逛商品，把心仪的加入收藏吧。'
          actionText='去逛逛'
          onAction={() => void Taro.switchTab({ url: '/pages/home/index' })}
        />
      )}
    </View>
  )
}
