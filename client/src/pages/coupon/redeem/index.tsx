import { useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { redeemCoupon } from '@/services/coupon'
import { Button, Input } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function CouponRedeemPage() {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleRedeem() {
    const trimmed = code.trim()
    if (!trimmed) {
      Taro.showToast({ title: '请输入兑换码', icon: 'none' })
      return
    }

    const confirmed = await Taro.showModal({
      title: '确认兑换',
      content: `兑换码：${trimmed}\n兑换后无法撤销，确认继续？`,
    })
    if (!confirmed.confirm) return

    setLoading(true)
    try {
      await redeemCoupon(trimmed)
      Taro.showToast({ title: '兑换成功', icon: 'success' })
      setCode('')
      setTimeout(() => {
        void Taro.redirectTo({ url: '/pages/user/coupons/index' })
      }, 800)
    } catch (error) {
      showErrorToast(error, '兑换失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='coupon-redeem'>
      <View className='coupon-redeem__card'>
        <Text className='coupon-redeem__title'>输入兑换码</Text>
        <Text className='coupon-redeem__hint'>每个兑换码只能使用一次</Text>

        <View className='coupon-redeem__input'>
          <Input
            placeholder='请输入您的兑换码'
            value={code}
            onChange={(v) => setCode(String(v).toUpperCase())}
            maxLength={32}
          />
        </View>

        <Button type='primary' block loading={loading} onClick={() => void handleRedeem()}>
          立即兑换
        </Button>

        <View className='coupon-redeem__notice'>
          <Text>
            • 兑换码区分大小写{'\n'}• 兑换成功后券会自动入账{'\n'}• 如有疑问请联系客服
          </Text>
        </View>
      </View>
    </View>
  )
}
