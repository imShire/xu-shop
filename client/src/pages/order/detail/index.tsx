import { View, Text, Image } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect, useRef } from 'react'
import PriceText from '@/components/PriceText'
import { getOrderDetail, confirmReceived, getShipTracks, cancelOrder } from '@/services/order'
import { usePay } from '@/hooks/usePay'
import type { OrderStatus } from '@/types/biz'
import { Skeleton, Cell, CellGroup, Button, SafeArea } from '@/ui/nutui'
import { formatAddress } from '@/utils/address'
import './index.scss'

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

function formatDateTime(iso: string): string {
  if (!iso) return ''
  return iso.slice(0, 16).replace('T', ' ')
}

function useCountdown(expireAt: string | undefined) {
  const [remaining, setRemaining] = useState<number>(0)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    if (!expireAt) return

    const calc = () => {
      const diff = Math.floor((new Date(expireAt).getTime() - Date.now()) / 1000)
      setRemaining(diff > 0 ? diff : 0)
    }

    calc()
    timerRef.current = setInterval(calc, 1000)

    return () => {
      if (timerRef.current !== null) clearInterval(timerRef.current)
    }
  }, [expireAt])

  return remaining
}

function formatCountdown(seconds: number): string {
  if (seconds <= 0) return '00:00'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  const mm = String(m).padStart(2, '0')
  const ss = String(s).padStart(2, '0')
  if (h > 0) return `${String(h).padStart(2, '0')}:${mm}:${ss}`
  return `${mm}:${ss}`
}

export default function OrderDetailPage() {
  const id = Taro.getCurrentInstance().router?.params?.id ?? ''
  const queryClient = useQueryClient()
  const [tracksExpanded, setTracksExpanded] = useState(false)

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrderDetail(id),
    enabled: id.length > 0,
  })

  const { data: tracks } = useQuery({
    queryKey: ['order-tracks', id],
    queryFn: () => getShipTracks(id),
    enabled: order?.status === 'shipped' || order?.status === 'delivered',
  })

  const remaining = useCountdown(
    order?.status === 'pending_payment' ? order.expire_at : undefined,
  )
  const isExpired = order?.status === 'pending_payment' && remaining === 0

  const confirmMutation = useMutation({
    mutationFn: () => confirmReceived(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['order', id] })
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void Taro.showToast({ title: '确认收货成功', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '操作失败，请重试', icon: 'none' }),
  })

  const cancelMutation = useMutation({
    mutationFn: () => cancelOrder(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['order', id] })
      void queryClient.invalidateQueries({ queryKey: ['orders'] })
      void Taro.showToast({ title: '订单已取消', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '取消失败，请重试', icon: 'none' }),
  })

  const { pay, loading: payLoading } = usePay({
    onSuccess: (orderId) => void queryClient.invalidateQueries({ queryKey: ['order', orderId] }),
  })

  const handleConfirm = () => {
    void Taro.showModal({
      title: '确认收货',
      content: '确认已收到商品？收货后如有问题可申请售后。',
      confirmText: '确认收货',
    }).then(({ confirm }) => {
      if (confirm) confirmMutation.mutate()
    })
  }

  const handleCancel = () => {
    void Taro.showModal({
      title: '取消订单',
      content: '确定要取消此订单吗？',
      confirmText: '确定取消',
      confirmColor: '#ff4d4f',
    }).then(({ confirm }) => {
      if (confirm) cancelMutation.mutate()
    })
  }

  const handleBuyAgain = () => {
    const firstItem = order?.items?.[0]
    if (firstItem) {
      void Taro.navigateTo({ url: `/pages/product/detail/index?id=${firstItem.product_id}` })
    }
  }

  if (isLoading || !order) {
    return (
      <View className='page-shell order-detail-page'>
        <Skeleton animated rows={8} />
      </View>
    )
  }

  const address = order.address_snapshot
  const items = order.items ?? []
  const trackList = tracks ?? []
  const visibleTracks = tracksExpanded ? trackList : trackList.slice(0, 2)

  const hasActionBar =
    order.status === 'pending_payment' ||
    order.status === 'shipped' ||
    order.status === 'completed'

  return (
    <View className='page-shell order-detail-page'>
      {/* Status card */}
      <View className='order-detail-page__status-card page-card'>
        <Text className='order-detail-page__status-label'>
          {STATUS_LABEL[order.status]}
        </Text>
        {order.status === 'pending_payment' && (
          <View className='order-detail-page__countdown-row'>
            {isExpired ? (
              <Text className='order-detail-page__countdown order-detail-page__countdown--expired'>
                已超时自动取消
              </Text>
            ) : (
              <>
                <Text className='order-detail-page__countdown-hint'>还剩</Text>
                <Text className='order-detail-page__countdown'>
                  {formatCountdown(remaining)}
                </Text>
                <Text className='order-detail-page__countdown-hint'>后自动取消</Text>
              </>
            )}
          </View>
        )}
      </View>

      {/* Address */}
      {address && (
        <CellGroup className='order-detail-page__section'>
          <Cell
            title={
              <View className='order-detail-page__address-header'>
                <Text className='order-detail-page__address-name'>{address.name}</Text>
                <Text className='order-detail-page__address-phone'>{address.phone}</Text>
              </View>
            }
            description={formatAddress(address)}
          />
        </CellGroup>
      )}

      {/* Items */}
      {items.length > 0 && (
        <View className='order-detail-page__section order-detail-page__section--items'>
          {items.map((item) => {
            const snap = item.product_snapshot
            return (
              <View key={item.id} className='order-detail-page__item'>
                {snap?.main_image ? (
                  <Image
                    className='order-detail-page__item-image'
                    src={snap.main_image}
                    mode='aspectFill'
                  />
                ) : (
                  <View className='order-detail-page__item-image order-detail-page__item-image--placeholder' />
                )}
                <View className='order-detail-page__item-info'>
                  <Text className='order-detail-page__item-title'>
                    {snap?.title ?? `商品 #${item.product_id}`}
                  </Text>
                  {snap?.attrs && Object.keys(snap.attrs).length > 0 && (
                    <Text className='order-detail-page__item-attrs'>
                      {Object.values(snap.attrs).join(' / ')}
                    </Text>
                  )}
                  <View className='order-detail-page__item-bottom'>
                    <PriceText cents={item.price_cents} />
                    <Text className='order-detail-page__item-qty'>× {item.qty}</Text>
                  </View>
                </View>
              </View>
            )
          })}
        </View>
      )}

      {/* Price summary */}
      <CellGroup className='order-detail-page__section'>
        <Cell title='商品金额' extra={`¥${(order.goods_cents / 100).toFixed(2)}`} />
        <Cell title='运费' extra={`¥${(order.freight_cents / 100).toFixed(2)}`} />
        {(order.coupon_discount_cents ?? 0) > 0 && (
          <Cell
            title='优惠券'
            extra={`-¥${((order.coupon_discount_cents ?? 0) / 100).toFixed(2)}`}
          />
        )}
        <Cell
          title='实付'
          extra={
            <Text className='order-detail-page__pay-amount'>
              ¥{(order.pay_cents / 100).toFixed(2)}
            </Text>
          }
        />
      </CellGroup>

      {/* Order meta */}
      <CellGroup className='order-detail-page__section'>
        <Cell title='订单号' extra={order.order_no} />
        <Cell title='下单时间' extra={formatDateTime(order.created_at)} />
        {order.buyer_remark && <Cell title='备注' extra={order.buyer_remark} />}
      </CellGroup>

      {/* Logistics tracks */}
      {trackList.length > 0 && (
        <CellGroup className='order-detail-page__section' title='物流轨迹'>
          {visibleTracks.map((track, i) => (
            <Cell
              key={i}
              title={track.content}
              description={track.time}
              className={i === 0 ? 'order-detail-page__track--latest' : ''}
            />
          ))}
          {trackList.length > 2 && (
            <Cell
              title={tracksExpanded ? '收起' : `展开查看全部 ${trackList.length} 条`}
              className='order-detail-page__track-toggle'
              clickable
              onClick={() => setTracksExpanded((v) => !v)}
            />
          )}
        </CellGroup>
      )}

      {/* Fixed action bar */}
      {hasActionBar && (
        <View className='order-detail-page__fixed-bar'>
          <View className='order-detail-page__fixed-panel'>
            {order.status === 'pending_payment' && (
              <>
                <Button
                  plain
                  type='default'
                  loading={cancelMutation.isPending}
                  onClick={handleCancel}
                >
                  取消订单
                </Button>
                <Button
                  type='primary'
                  loading={payLoading}
                  disabled={isExpired}
                  onClick={() => void pay(order.id)}
                >
                  去支付
                </Button>
              </>
            )}
            {order.status === 'shipped' && (
              <Button
                type='primary'
                block
                loading={confirmMutation.isPending}
                onClick={handleConfirm}
              >
                确认收货
              </Button>
            )}
            {order.status === 'completed' && (
              <Button type='default' plain block onClick={handleBuyAgain}>
                再次购买
              </Button>
            )}
          </View>
          <SafeArea position='bottom' />
        </View>
      )}
    </View>
  )
}
