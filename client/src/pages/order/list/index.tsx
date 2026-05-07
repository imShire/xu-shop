import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useCallback } from 'react'
import PullList from '@/components/PullList'
import PriceText from '@/components/PriceText'
import { useAuthStore } from '@/stores/auth'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { getOrders, cancelOrder, confirmReceived } from '@/services/order'
import { usePay } from '@/hooks/usePay'
import type { Order, OrderStatus } from '@/types/biz'
import { Button, Tabs, TabPane, Tag } from '@/ui/nutui'
import './index.scss'

const PAGE_SIZE = 10

const STATUS_TABS: Array<{ label: string; value: OrderStatus | '' }> = [
  { label: '全部', value: '' },
  { label: '待付款', value: 'pending_payment' },
  { label: '待发货', value: 'paid' },
  { label: '已发货', value: 'shipped' },
  { label: '已完成', value: 'completed' },
]

const STATUS_LABEL: Record<OrderStatus, string> = {
  pending_payment: '待付款',
  paid: '待发货',
  shipped: '已发货',
  delivered: '已送达',
  completed: '已完成',
  cancelled: '已取消',
  refunding: '退款中',
  refunded: '已退款',
}

const STATUS_TYPE: Record<OrderStatus, 'primary' | 'warning' | 'success' | 'danger' | 'default'> = {
  pending_payment: 'warning',
  paid: 'primary',
  shipped: 'primary',
  delivered: 'primary',
  completed: 'success',
  cancelled: 'default',
  refunding: 'warning',
  refunded: 'default',
}

interface OrderCardProps {
  order: Order
  onCancel: (id: string) => void
  onPay: (id: string) => void
  onConfirm: (id: string) => void
}

function OrderCard({ order, onCancel, onPay, onConfirm }: OrderCardProps) {
  const items = order.items ?? []
  const totalQty = items.reduce((s, i) => s + i.qty, 0)
  const firstItem = items[0]

  const handleCardClick = () => {
    void Taro.navigateTo({ url: `/pages/order/detail/index?id=${order.id}` })
  }

  const handleCancel = (e: { stopPropagation: () => void }) => {
    e.stopPropagation()
    void Taro.showModal({
      title: '取消订单',
      content: '确定要取消此订单吗？',
      confirmText: '确定取消',
      confirmColor: '#ff4d4f',
    }).then(({ confirm }) => {
      if (confirm) onCancel(order.id)
    })
  }

  const handlePay = (e: { stopPropagation: () => void }) => {
    e.stopPropagation()
    onPay(order.id)
  }

  const handleConfirm = (e: { stopPropagation: () => void }) => {
    e.stopPropagation()
    void Taro.showModal({
      title: '确认收货',
      content: '确认已收到商品？收货后如有问题可申请售后。',
      confirmText: '确认收货',
    }).then(({ confirm }) => {
      if (confirm) onConfirm(order.id)
    })
  }

  const handleBuyAgain = (e: { stopPropagation: () => void }) => {
    e.stopPropagation()
    if (firstItem) {
      void Taro.navigateTo({ url: `/pages/product/detail/index?id=${firstItem.product_id}` })
    }
  }

  const createdAt = order.created_at ? order.created_at.slice(0, 10) : ''

  return (
    <View className='order-list-page__card' onClick={handleCardClick}>
      {/* Header row */}
      <View className='order-list-page__card-header'>
        <Text className='order-list-page__order-no'>订单号: {order.order_no}</Text>
        <Text className='order-list-page__order-date'>{createdAt}</Text>
      </View>

      {/* Items */}
      {items.map((item) => {
        const snap = item.product_snapshot
        return (
          <View key={item.id} className='order-list-page__item'>
            {snap?.main_image ? (
              <Image
                className='order-list-page__item-image'
                src={snap.main_image}
                mode='aspectFill'
              />
            ) : (
              <View className='order-list-page__item-image order-list-page__item-image--placeholder' />
            )}
            <View className='order-list-page__item-info'>
              <Text className='order-list-page__item-title'>
                {snap?.title ?? `商品 #${item.product_id}`}
              </Text>
              {snap?.attrs && Object.keys(snap.attrs).length > 0 && (
                <Text className='order-list-page__item-attrs'>
                  {Object.values(snap.attrs).join(' / ')}
                </Text>
              )}
              <View className='order-list-page__item-bottom'>
                <PriceText cents={item.price_cents} />
                <Text className='order-list-page__item-qty'>× {item.qty}</Text>
              </View>
            </View>
          </View>
        )
      })}

      {/* Footer: total + status + actions */}
      <View className='order-list-page__card-footer'>
        <View className='order-list-page__footer-left'>
          <Tag type={STATUS_TYPE[order.status]}>{STATUS_LABEL[order.status]}</Tag>
          <Text className='order-list-page__total'>
            共 {totalQty} 件  实付:{' '}
            <Text className='order-list-page__total-price'>¥{(order.pay_cents / 100).toFixed(2)}</Text>
          </Text>
        </View>
        <View className='order-list-page__actions'>
          {order.status === 'pending_payment' && (
            <>
              <Button size='small' type='default' plain onClick={handleCancel}>
                取消订单
              </Button>
              <Button size='small' type='primary' onClick={handlePay}>
                去支付
              </Button>
            </>
          )}
          {order.status === 'shipped' && (
            <Button size='small' type='primary' onClick={handleConfirm}>
              确认收货
            </Button>
          )}
          {order.status === 'completed' && (
            <Button size='small' type='default' plain onClick={handleBuyAgain}>
              再次购买
            </Button>
          )}
        </View>
      </View>
    </View>
  )
}

export default function OrderListPage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<OrderStatus | ''>('')

  useDidShow(() => {
    void queryClient.invalidateQueries({ queryKey: ['orders'] })
  })

  const {
    data: pagesData,
    isFetching,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
  } = useInfiniteQuery({
    queryKey: ['orders', activeTab],
    queryFn: ({ pageParam }) =>
      getOrders({
        status: activeTab || undefined,
        page: pageParam as number,
        page_size: PAGE_SIZE,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((acc, p) => acc + p.list.length, 0)
      if (lastPage.list.length < PAGE_SIZE) return undefined
      return Math.floor(loaded / PAGE_SIZE) + 1
    },
    enabled: isLoggedIn,
  })

  const allOrders: Order[] = pagesData?.pages.flatMap((p) => p.list) ?? []

  const handleLoadMore = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      void fetchNextPage()
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage])

  const handleTabChange = (value: string) => {
    setActiveTab(value as OrderStatus | '')
    void queryClient.removeQueries({ queryKey: ['orders', value] })
  }

  const cancelMutation = useMutation({
    mutationFn: (id: string) => cancelOrder(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void Taro.showToast({ title: '订单已取消', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '取消失败，请重试', icon: 'none' }),
  })

  const confirmMutation = useMutation({
    mutationFn: (id: string) => confirmReceived(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void Taro.showToast({ title: '确认收货成功', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '操作失败，请重试', icon: 'none' }),
  })

  const { pay } = usePay({
    onSuccess: (orderId) =>
      void Taro.navigateTo({ url: `/pages/order/detail/index?id=${orderId}` }),
  })

  // Redirect to login if not logged in — after all hooks
  if (!isLoggedIn) {
    void ensureAuth(undefined, '/pages/order/list/index')
    return <View className='page-shell order-list-page' />
  }

  return (
    <View className='page-shell order-list-page'>
      <Tabs
        value={activeTab}
        onChange={(value) => handleTabChange(String(value))}
        className='order-list-page__tabs'
      >
        {STATUS_TABS.map((tab) => (
          <TabPane key={tab.value} value={tab.value} title={tab.label}>
            <PullList
              data={allOrders}
              loading={isFetching && allOrders.length === 0}
              hasMore={hasNextPage}
              onLoadMore={handleLoadMore}
              renderItem={(order) => (
                <OrderCard
                  order={order}
                  onCancel={(id) => cancelMutation.mutate(id)}
                  onPay={(id) => void pay(id)}
                  onConfirm={(id) => confirmMutation.mutate(id)}
                />
              )}
              keyExtractor={(order) => String(order.id)}
              emptyTitle='暂无订单'
              emptyDescription='去购物，把心仪的商品带回家。'
              emptyActionText='去逛逛'
              onEmptyAction={() => void Taro.switchTab({ url: '/pages/home/index' })}
              className='order-list-page__pull-list'
            />
          </TabPane>
        ))}
      </Tabs>
    </View>
  )
}
