import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query'
import { useState, useCallback, useMemo } from 'react'
import PullList from '@/components/PullList'
import { useAuthStore } from '@/stores/auth'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { getOrders } from '@/services/order'
import type { Order } from '@/types/biz'
import { Tabs, TabPane } from '@/ui/nutui'
import './index.scss'

const PAGE_SIZE = 10

type AftersaleTab = 'pending' | 'refunded' | 'closed'

const TABS: Array<{ key: AftersaleTab; label: string; serverStatus: string }> = [
  { key: 'pending', label: '申请中', serverStatus: 'paid' },
  { key: 'refunded', label: '已退款', serverStatus: 'refunded' },
  { key: 'closed', label: '已关闭', serverStatus: 'cancelled' },
]

function formatCardItem(tab: AftersaleTab) {
  if (tab === 'pending') {
    return {
      statusText: '取消申请审核中',
      showReason: true,
    }
  }
  if (tab === 'refunded') {
    return { statusText: '已退款', showReason: false }
  }
  return { statusText: '订单已关闭', showReason: false }
}

interface AftersaleCardProps {
  order: Order
  tab: AftersaleTab
}

function AftersaleCard({ order, tab }: AftersaleCardProps) {
  const items = order.items ?? []
  const { statusText, showReason } = formatCardItem(tab)
  const createdAt = order.created_at ? order.created_at.slice(0, 10) : ''

  const handleClick = () => {
    void Taro.navigateTo({ url: `/pages/order/detail/index?id=${order.id}` })
  }

  return (
    <View className='aftersale-list-page__card' onClick={handleClick}>
      <View className='aftersale-list-page__card-header'>
        <Text className='aftersale-list-page__order-no'>订单号: {order.order_no}</Text>
        <Text>{createdAt}</Text>
      </View>

      {items.map((item) => {
        const snap = item.product_snapshot
        return (
          <View key={item.id} className='aftersale-list-page__item'>
            {snap?.main_image ? (
              <Image
                className='aftersale-list-page__item-image'
                src={snap.main_image}
                mode='aspectFill'
              />
            ) : (
              <View className='aftersale-list-page__item-image aftersale-list-page__item-image--placeholder' />
            )}
            <View className='aftersale-list-page__item-info'>
              <Text className='aftersale-list-page__item-title'>
                {snap?.title ?? `商品 #${item.product_id}`}
              </Text>
              <Text className='aftersale-list-page__item-meta'>× {item.qty}</Text>
            </View>
          </View>
        )
      })}

      {showReason && order.cancel_request_reason && (
        <Text className='aftersale-list-page__reason'>
          申请原因: {order.cancel_request_reason}
        </Text>
      )}

      <View className='aftersale-list-page__card-footer'>
        <Text className='aftersale-list-page__status-text'>{statusText}</Text>
        <Text className='aftersale-list-page__total'>
          实付:{' '}
          <Text className='aftersale-list-page__total-price'>
            ¥{(order.pay_cents / 100).toFixed(2)}
          </Text>
        </Text>
      </View>
    </View>
  )
}

function useAftersaleQuery(tab: AftersaleTab, enabled: boolean) {
  const serverStatus = TABS.find((t) => t.key === tab)!.serverStatus

  return useInfiniteQuery({
    queryKey: ['aftersale-list', tab],
    queryFn: ({ pageParam }) =>
      getOrders({
        status: serverStatus,
        page: pageParam as number,
        page_size: PAGE_SIZE,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((acc, p) => acc + p.list.length, 0)
      if (lastPage.list.length < PAGE_SIZE) return undefined
      return Math.floor(loaded / PAGE_SIZE) + 1
    },
    enabled,
  })
}

export default function AftersaleListPage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<AftersaleTab>('pending')

  useDidShow(() => {
    void queryClient.invalidateQueries({ queryKey: ['aftersale-list'] })
  })

  const query = useAftersaleQuery(activeTab, isLoggedIn)
  const { isFetching, isFetchingNextPage, hasNextPage, fetchNextPage, data } = query

  // 申请中 tab 需在客户端按 cancel_request_pending 二次过滤
  const allOrders: Order[] = useMemo(() => {
    const list = data?.pages.flatMap((p) => p.list) ?? []
    if (activeTab === 'pending') {
      return list.filter((o) => o.cancel_request_pending === true)
    }
    return list
  }, [data, activeTab])

  const handleLoadMore = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      void fetchNextPage()
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  if (!isLoggedIn) {
    void ensureAuth(undefined, '/pages/aftersale/list/index')
    return <View className='page-shell aftersale-list-page' />
  }

  return (
    <View className='page-shell aftersale-list-page'>
      <Tabs
        value={activeTab}
        onChange={(value) => setActiveTab(value as AftersaleTab)}
        className='aftersale-list-page__tabs'
      >
        {TABS.map((tab) => (
          <TabPane key={tab.key} value={tab.key} title={tab.label}>
            <PullList
              data={allOrders}
              loading={isFetching && allOrders.length === 0}
              hasMore={hasNextPage}
              onLoadMore={handleLoadMore}
              renderItem={(order) => <AftersaleCard order={order} tab={activeTab} />}
              keyExtractor={(order) => String(order.id)}
              emptyTitle='暂无相关订单'
              emptyDescription={
                activeTab === 'pending'
                  ? '你还没有取消申请'
                  : activeTab === 'refunded'
                    ? '暂无已退款订单'
                    : '暂无已关闭订单'
              }
              className='aftersale-list-page__pull-list'
            />
          </TabPane>
        ))}
      </Tabs>
    </View>
  )
}
