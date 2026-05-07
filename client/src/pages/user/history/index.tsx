import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { useAuthStore } from '@/stores/auth'
import ProductCard from '@/components/ProductCard'
import { clearHistory, getHistory, type HistoryItem } from '@/services/user'
import { ArrowRight } from '@/ui/icons'
import { Skeleton } from '@/ui/nutui'
import EmptyState from '@/components/EmptyState'
import dayjs from 'dayjs'
import './index.scss'

function getDateLabel(dateStr: string): string {
  const date = dayjs(dateStr)
  const today = dayjs().startOf('day')
  const yesterday = today.subtract(1, 'day')

  if (date.isAfter(today)) return '今天'
  if (date.isAfter(yesterday)) return '昨天'
  return date.format('MM月DD日')
}

interface HistoryGroup {
  label: string
  items: HistoryItem[]
}

function groupByDate(items: HistoryItem[]): HistoryGroup[] {
  const map = new Map<string, HistoryItem[]>()
  for (const item of items) {
    const label = getDateLabel(item.viewed_at)
    if (!map.has(label)) {
      map.set(label, [])
    }
    map.get(label)!.push(item)
  }
  return Array.from(map.entries()).map(([label, groupItems]) => ({ label, items: groupItems }))
}

export default function HistoryPage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()

  useDidShow(() => {
    if (!isLoggedIn) {
      void ensureAuth(undefined, '/pages/user/history/index')
    }
  })

  const { data, isLoading } = useQuery({
    queryKey: ['history'],
    queryFn: () => getHistory({ page: 1, page_size: 20 }),
    enabled: isLoggedIn,
  })

  const clearMutation = useMutation({
    mutationFn: () => clearHistory(),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['history'] })
      void Taro.showToast({ title: '已清空浏览记录', icon: 'none' })
    },
    onError: () => {
      void Taro.showToast({ title: '清空失败，请重试', icon: 'none' })
    },
  })

  const handleClear = () => {
    void Taro.showModal({
      title: '清空浏览记录',
      content: '确认清空全部浏览记录？',
      confirmText: '清空',
      cancelText: '取消',
    }).then(({ confirm }) => {
      if (confirm) {
        clearMutation.mutate()
      }
    })
  }

  const items = data?.list ?? []
  const groups = groupByDate(items)

  if (!isLoggedIn) return <View className='page-shell' />

  return (
    <View className='page-shell history-page'>
      {isLoading ? (
        <View className='history-page__list'>
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} animated rows={4} />
          ))}
        </View>
      ) : items.length ? (
        <View className='history-page__groups'>
          {groups.map((group) => (
            <View key={group.label} className='history-page__group'>
              <View className='history-page__group-head'>
                <Text className='history-page__group-label'>{group.label}</Text>
                {group === groups[0] ? (
                  <Text className='history-page__clear-action' onClick={handleClear}>
                    清空
                  </Text>
                ) : null}
              </View>
              <View className='history-page__panel'>
                {group.items.map((item) => (
                  <ProductCard
                    key={item.id}
                    mode='horizontal'
                    product={{
                      id: item.product_id,
                      title: item.title,
                      main_image: item.image,
                      price_cents: item.price_cents,
                    }}
                    cartButton={<ArrowRight width={14} height={14} />}
                    onClick={() =>
                      void Taro.navigateTo({
                        url: `/pages/product/detail/index?id=${item.product_id}`,
                      })
                    }
                  />
                ))}
              </View>
            </View>
          ))}
        </View>
      ) : (
        <EmptyState
          title='暂无浏览记录'
          description='逛一逛，浏览记录会显示在这里。'
          actionText='去逛逛'
          onAction={() => void Taro.switchTab({ url: '/pages/home/index' })}
        />
      )}
    </View>
  )
}
