import { useState } from 'react'
import { View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery } from '@tanstack/react-query'
import CouponItem from '@/components/CouponItem'
import EmptyState from '@/components/EmptyState'
import { getMyCoupons, type UserCouponStatus } from '@/services/coupon'
import { useAuthStore } from '@/stores/auth'
import { Cell, Skeleton, Tabs, TabPane } from '@/ui/nutui'
import './index.scss'

const TABS: Array<{ value: UserCouponStatus; label: string }> = [
  { value: 'unused', label: '未使用' },
  { value: 'used', label: '已使用' },
  { value: 'expired', label: '已过期' },
]

export default function UserCouponsPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const [tab, setTab] = useState<UserCouponStatus>('unused')

  const listQuery = useQuery({
    queryKey: ['my-coupons', tab],
    queryFn: () => getMyCoupons({ status: tab, page: 1, page_size: 50 }),
    enabled: isLoggedIn,
  })

  return (
    <View className='user-coupons'>
      <View className='user-coupons__entries'>
        <Cell
          title='领券中心'
          extra='去领券 →'
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/coupon/center/index' })}
        />
        <Cell
          title='输入兑换码'
          extra='去兑换 →'
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/coupon/redeem/index' })}
        />
      </View>

      <Tabs value={tab} onChange={(v) => setTab(v as UserCouponStatus)}>
        {TABS.map((t) => (
          <TabPane key={t.value} title={t.label} value={t.value} />
        ))}
      </Tabs>

      <View className='user-coupons__list'>
        {listQuery.isLoading ? (
          <Skeleton animated rows={4} />
        ) : (listQuery.data?.list ?? []).length === 0 ? (
          <EmptyState title='暂无优惠券' />
        ) : (
          listQuery.data!.list.map((c) => (
            <CouponItem key={c.id} userCoupon={c} status={c.status} disabled={tab !== 'unused'} />
          ))
        )}
      </View>
    </View>
  )
}
