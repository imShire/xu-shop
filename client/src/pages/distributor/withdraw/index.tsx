import { useEffect, useRef, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery } from '@tanstack/react-query'
import {
  applyWithdraw,
  getMyDistributor,
  sendWithdrawSms,
} from '@/services/distributor'
import { Button, Input, Skeleton } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import { formatPrice, formatYuan } from '@/utils/price'
import './index.scss'

export default function WithdrawPage() {
  const profileQuery = useQuery({ queryKey: ['my-distributor'], queryFn: getMyDistributor })

  const [amountYuan, setAmountYuan] = useState('')
  const [smsCode, setSmsCode] = useState('')
  const [countdown, setCountdown] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [sending, setSending] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(
    () => () => {
      if (timerRef.current) clearInterval(timerRef.current)
    },
    [],
  )

  const balanceCents = profileQuery.data?.available_commission_cents ?? 0
  const amountCents = Math.round(Number(amountYuan || 0) * 100)

  async function handleSendSms() {
    if (countdown > 0 || sending) return
    setSending(true)
    try {
      await sendWithdrawSms()
      Taro.showToast({ title: '验证码已发送', icon: 'success' })
      setCountdown(60)
      timerRef.current = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) {
            if (timerRef.current) clearInterval(timerRef.current)
            return 0
          }
          return c - 1
        })
      }, 1000)
    } catch (e) {
      showErrorToast(e, '发送失败')
    } finally {
      setSending(false)
    }
  }

  async function handleSubmit() {
    if (amountCents <= 0) {
      Taro.showToast({ title: '请输入提现金额', icon: 'none' })
      return
    }
    if (amountCents < 1000) {
      Taro.showToast({ title: '最低提现金额 ¥10.00', icon: 'none' })
      return
    }
    if (amountCents > balanceCents) {
      Taro.showToast({ title: '余额不足', icon: 'none' })
      return
    }
    if (!smsCode.trim()) {
      Taro.showToast({ title: '请输入短信验证码', icon: 'none' })
      return
    }
    const ok = await Taro.showModal({
      title: '确认提现',
      content: `提现金额：${formatYuan(amountCents)}\n将通过微信零钱发放，确认提交？`,
    })
    if (!ok.confirm) return

    setSubmitting(true)
    try {
      const key = `wd-${Date.now()}-${Math.random().toString(36).slice(2)}`
      await applyWithdraw({ amount_cents: amountCents, sms_code: smsCode.trim() }, key)
      Taro.showToast({ title: '申请已提交', icon: 'success' })
      setTimeout(() => {
        void Taro.redirectTo({ url: '/pages/distributor/withdraws/index' })
      }, 1000)
    } catch (e) {
      showErrorToast(e, '提现失败')
    } finally {
      setSubmitting(false)
    }
  }

  if (profileQuery.isLoading) {
    return (
      <View className='dist-withdraw'>
        <Skeleton animated rows={4} />
      </View>
    )
  }

  return (
    <View className='dist-withdraw'>
      <View className='dist-withdraw__balance'>
        <Text className='dist-withdraw__balance-label'>可提现佣金（元）</Text>
        <Text className='dist-withdraw__balance-value'>{formatPrice(balanceCents)}</Text>
      </View>

      <View className='dist-withdraw__form'>
        <View className='dist-withdraw__field'>
          <Text className='dist-withdraw__label'>提现金额</Text>
          <View className='dist-withdraw__input'>
            <Input
              type='digit'
              placeholder='最低 10 元'
              value={amountYuan}
              onChange={(v) => setAmountYuan(String(v))}
            />
            <Text
              className='dist-withdraw__send-sms'
              onClick={() => setAmountYuan(formatPrice(balanceCents))}
            >
              全部
            </Text>
          </View>
        </View>

        <View className='dist-withdraw__field'>
          <Text className='dist-withdraw__label'>短信验证码</Text>
          <View className='dist-withdraw__input'>
            <Input
              type='number'
              placeholder='请输入 6 位验证码'
              value={smsCode}
              onChange={(v) => setSmsCode(String(v))}
              maxLength={6}
            />
            <Text
              className={`dist-withdraw__send-sms ${countdown > 0 || sending ? 'dist-withdraw__send-sms--disabled' : ''}`}
              onClick={() => void handleSendSms()}
            >
              {countdown > 0 ? `${countdown}s 后重发` : sending ? '发送中' : '获取验证码'}
            </Text>
          </View>
        </View>

        <Button type='primary' block loading={submitting} onClick={() => void handleSubmit()}>
          提交申请
        </Button>

        <View className='dist-withdraw__notice'>
          • 单笔最低 ¥10.00，单日累计最高 ¥5000{'\n'}
          • 提现通过微信零钱发放，1-3 个工作日到账{'\n'}
          • 请确保已实名认证，否则可能审核失败
        </View>
      </View>
    </View>
  )
}
