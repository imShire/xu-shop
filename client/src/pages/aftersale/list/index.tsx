import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useMemo, useState } from 'react'
import PullList from '@/components/PullList'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import {
  AFTERSALE_STATUS_LABEL,
  AFTERSALE_TYPE_LABEL,
  isAftersaleTerminal,
  listAftersales,
} from '@/services/aftersale'
import { useAuthStore } from '@/stores/auth'
import type { AftersaleOrder, AftersaleStatus } from '@/types/biz'
import { TabPane, Tabs } from '@/ui/nutui'
import './index.scss'

const PAGE_SIZE = 10

type TabKey = 'all' | 'processing' | 'completed' | 'closed'

interface TabDef {
  key: TabKey
  label: string
  statuses: AftersaleStatus[] | null
}

const TABS: TabDef[] = [
  { key: 'all', label: '全部', statuses: null },
  {
    key: 'processing',
    label: '处理中',
    statuses: ['applying', 'seller_agreed', 'buyer_returned', 'seller_received'],
  },
  { key: 'completed', label: '已完成', statuses: ['completed'] },
  { key: 'closed', label: '已关闭', statuses: ['seller_rejected', 'cancelled', 'closed'] },
]

function statusClass(status: AftersaleStatus): string {
  if (isAftersaleTerminal(status)) {
    if (status === 'completed') return 'aftersale-list-page__status'
    return 'aftersale-list-page__status aftersale-list-page__status--terminal'
  }
  if (status === 'applying') {
    return 'aftersale-list-page__status aftersale-list-page__status--warn'
  }
  return 'aftersale-list-page__status'
}

function AftersaleCard({ item }: { item: AftersaleOrder }) {
  const snap = item.item_snapshot
  const handleClick = () => {
    void Taro.navigateTo({ url: `/pages/aftersale/detail/index?id=${item.id}` })
  }

  return (
    <View className='aftersale-list-page__card' onClick={handleClick}>
      <View className='aftersale-list-page__header'>
        <Text className='aftersale-list-page__no'>售后单号 {item.aftersale_no}</Text>
        <Text className={statusClass(item.status)}>
          {AFTERSALE_STATUS_LABEL[item.status]}
        </Text>
      </View>
      <View className='aftersale-list-page__body'>
        {snap?.image ? (
          <Image className='aftersale-list-page__image' src={snap.image} mode='aspectFill' />
        ) : (
          <View className='aftersale-list-page__image' />
        )}
        <View className='aftersale-list-page__info'>
          <Text className='aftersale-list-page__title'>
            {snap?.product_name ?? `订单 ${item.order_no}`}
          </Text>
          <Text className='aftersale-list-page__sub'>
            {AFTERSALE_TYPE_LABEL[item.type]} · 申请于{' '}
            {item.applied_at ? item.applied_at.slice(0, 10) : ''}
          </Text>
        </View>
      </View>
      <View className='aftersale-list-page__footer'>
        <Text>关联订单 {item.order_no}</Text>
        <Text>
          退款{' '}
          <Text className='aftersale-list-page__amount'>
            ¥{(item.refund_amount_cents / 100).toFixed(2)}
          </Text>
        </Text>
      </View>
    </View>
  )
}

function useAftersaleListQuery(tab: TabKey, enabled: boolean) {
  return useInfiniteQuery({
    queryKey: ['aftersale-list', tab],
    queryFn: ({ pageParam }) =>
      listAftersales({
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
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<TabKey>('all')

  useDidShow(() => {
    void queryClient.invalidateQueries({ queryKey: ['aftersale-list'] })
  })

  const query = useAftersaleListQuery(activeTab, isLoggedIn)
  const { isFetching, isFetchingNextPage, hasNextPage, fetchNextPage, data } = query

  const items: AftersaleOrder[] = useMemo(() => {
    const list = data?.pages.flatMap((p) => p.list) ?? []
    const def = TABS.find((t) => t.key === activeTab)!
    if (!def.statuses) return list
    return list.filter((it) => def.statuses!.includes(it.status))
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
        onChange={(value) => setActiveTab(value as TabKey)}
        className='aftersale-list-page__tabs'
      >
        {TABS.map((tab) => (
          <TabPane key={tab.key} value={tab.key} title={tab.label}>
            <PullList
              data={items}
              loading={isFetching && items.length === 0}
              hasMore={hasNextPage}
              onLoadMore={handleLoadMore}
              renderItem={(it) => <AftersaleCard item={it} />}
              keyExtractor={(it) => String(it.id)}
              emptyTitle='暂无售后记录'
              emptyDescription='发起售后申请后将在此显示'
              className='aftersale-list-page__pull-list'
            />
          </TabPane>
        ))}
      </Tabs>
    </View>
  )
}
