import { useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import CouponItem from '@/components/CouponItem'
import EmptyState from '@/components/EmptyState'
import { claimCoupon, getAvailableCoupons } from '@/services/coupon'
import { useAuthStore } from '@/stores/auth'
import { Skeleton, Tabs, TabPane } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function CouponCenterPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'all' | 'claimed'>('all')

  const listQuery = useQuery({
    queryKey: ['coupon-center'],
    queryFn: () => getAvailableCoupons({ page: 1, page_size: 50 }),
    enabled: isLoggedIn,
  })

  const claimMutation = useMutation({
    mutationFn: (templateId: string) => {
      const key = `coupon-${templateId}-${Date.now()}`
      return claimCoupon(templateId, key)
    },
    onSuccess: () => {
      Taro.showToast({ title: '领取成功', icon: 'success' })
      void queryClient.invalidateQueries({ queryKey: ['coupon-center'] })
      void queryClient.invalidateQueries({ queryKey: ['my-coupons'] })
    },
    onError: (err) => showErrorToast(err, '领取失败'),
  })

  const list = listQuery.data?.list ?? []
  const filtered = tab === 'claimed' ? list.filter((c) => c.is_claimed) : list

  if (!isLoggedIn) {
    return (
      <View className='coupon-center'>
        <EmptyState
          title='请先登录'
          description='登录后查看可领取的优惠券'
          action={{
            text: '去登录',
            onClick: () =>
              void Taro.navigateTo({ url: '/pages/auth/login/index?redirect=/pages/coupon/center/index' }),
          }}
        />
      </View>
    )
  }

  return (
    <View className='coupon-center'>
      <View
        className='coupon-center__redeem-bar'
        onClick={() => void Taro.navigateTo({ url: '/pages/coupon/redeem/index' })}
      >
        <Text className='coupon-center__redeem-title'>有兑换码？</Text>
        <Text className='coupon-center__redeem-link'>立即兑换 →</Text>
      </View>

      <Tabs value={tab} onChange={(v) => setTab(v as 'all' | 'claimed')}>
        <TabPane title='全部' value='all' />
        <TabPane title='已领取' value='claimed' />
      </Tabs>

      {listQuery.isLoading ? (
        <Skeleton animated rows={6} />
      ) : filtered.length === 0 ? (
        <EmptyState title='暂无可领取的优惠券' />
      ) : (
        filtered.map((tpl) => {
          const claimed = !!tpl.is_claimed
          const soldOut =
            tpl.total_quota !== undefined &&
            tpl.total_quota > 0 &&
            (tpl.claimed_count ?? 0) >= tpl.total_quota
          return (
            <CouponItem
              key={tpl.id}
              template={tpl}
              status={claimed ? 'claimed' : 'claimable'}
              action={{
                text: claimed ? '已领取' : soldOut ? '已抢光' : '立即领取',
                disabled: claimed || soldOut || claimMutation.isPending,
                onClick: () => claimMutation.mutate(tpl.id),
              }}
            />
          )
        })
      )}
    </View>
  )
}
