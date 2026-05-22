import { View, Text, ScrollView } from '@tarojs/components'
import CouponItem from '@/components/CouponItem'
import EmptyState from '@/components/EmptyState'
import type { UserCoupon } from '@/types/biz'
import { Popup } from '@/ui/nutui'
import './index.scss'

interface CouponPickerProps {
  visible: boolean
  coupons: UserCoupon[]
  selectedId?: string | null
  loading?: boolean
  onClose: () => void
  onSelect: (couponId: string | null) => void
}

export default function CouponPicker({
  visible,
  coupons,
  selectedId,
  loading,
  onClose,
  onSelect,
}: CouponPickerProps) {
  return (
    <Popup
      visible={visible}
      position='bottom'
      onClose={onClose}
      style={{ height: '70vh' }}
      round
    >
      <View className='coupon-picker'>
        <View className='coupon-picker__head'>
          <Text className='coupon-picker__title'>选择优惠券</Text>
          <Text className='coupon-picker__close' onClick={onClose}>
            关闭
          </Text>
        </View>

        <ScrollView scrollY className='coupon-picker__body'>
          <View
            className={`coupon-picker__none ${selectedId === null || selectedId === '' ? 'coupon-picker__none--active' : ''}`}
            onClick={() => onSelect(null)}
          >
            <Text>不使用优惠券</Text>
          </View>

          {loading ? (
            <Text className='coupon-picker__loading'>加载中...</Text>
          ) : coupons.length === 0 ? (
            <EmptyState title='暂无可用优惠券' description='下次再来看看吧' />
          ) : (
            coupons.map((c) => {
              const usable = c.applicable !== false
              return (
                <CouponItem
                  key={c.id}
                  userCoupon={c}
                  status={c.status}
                  selected={selectedId === c.id}
                  disabled={!usable}
                  onClick={() => usable && onSelect(c.id)}
                />
              )
            })
          )}
        </ScrollView>
      </View>
    </Popup>
  )
}
