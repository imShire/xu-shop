import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useEffect, useMemo, useState } from 'react'
import EvidenceUploader from '@/components/EvidenceUploader'
import { applyAftersale } from '@/services/aftersale'
import { getOrderDetail } from '@/services/order'
import type { AftersaleType } from '@/types/biz'
import { Button, Input, SafeArea, Skeleton, TextArea } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

interface TypeOption {
  value: AftersaleType
  label: string
  desc: string
}

const TYPE_OPTIONS: TypeOption[] = [
  { value: 'refund_only', label: '仅退款', desc: '未收到货 / 商家协商一致退款' },
  { value: 'refund_return', label: '退货退款', desc: '已收到货，需寄回商品' },
  { value: 'exchange', label: '换货', desc: '已收到货，希望更换商品' },
]

function parseYuan(value: string): number {
  const num = Number(value)
  if (Number.isNaN(num) || num < 0) return 0
  return Math.round(num * 100)
}

export default function AftersaleApplyPage() {
  const params = Taro.getCurrentInstance().router?.params ?? {}
  const orderId = params.order_id ?? ''
  const orderItemId = params.order_item_id ?? ''

  const orderQuery = useQuery({
    queryKey: ['order', orderId],
    queryFn: () => getOrderDetail(orderId),
    enabled: orderId.length > 0,
  })

  const order = orderQuery.data
  const targetItem = useMemo(() => {
    if (!order || !orderItemId) return undefined
    return order.items?.find((it) => String(it.id) === orderItemId)
  }, [order, orderItemId])

  // 金额上限
  const maxCents = useMemo(() => {
    if (targetItem) return targetItem.price_cents * targetItem.qty
    return order?.pay_cents ?? 0
  }, [order, targetItem])

  const [type, setType] = useState<AftersaleType>('refund_return')
  const [reason, setReason] = useState('')
  const [amountStr, setAmountStr] = useState('')
  const [evidence, setEvidence] = useState<string[]>([])

  // 初始化金额为上限
  useEffect(() => {
    if (maxCents > 0 && amountStr === '') {
      setAmountStr((maxCents / 100).toFixed(2))
    }
  }, [maxCents, amountStr])

  const amountCents = type === 'exchange' ? 0 : parseYuan(amountStr)

  const submitMutation = useMutation({
    mutationFn: () =>
      applyAftersale({
        order_id: orderId,
        order_item_id: orderItemId || null,
        type,
        reason: reason.trim(),
        refund_amount_cents: amountCents,
        evidence: evidence.length > 0 ? evidence : undefined,
      }),
    onSuccess: (res) => {
      void Taro.showToast({ title: '已提交申请', icon: 'success' })
      setTimeout(() => {
        void Taro.redirectTo({ url: `/pages/aftersale/detail/index?id=${res.id}` })
      }, 500)
    },
    onError: (err) => showErrorToast(err, '提交失败，请稍后再试'),
  })

  function handleSubmit() {
    if (!orderId) {
      void Taro.showToast({ title: '缺少订单参数', icon: 'none' })
      return
    }
    const r = reason.trim()
    if (r.length < 2) {
      void Taro.showToast({ title: '请填写申请原因（至少 2 字）', icon: 'none' })
      return
    }
    if (r.length > 200) {
      void Taro.showToast({ title: '原因不超过 200 字', icon: 'none' })
      return
    }
    if (type !== 'exchange') {
      if (amountCents <= 0) {
        void Taro.showToast({ title: '请填写退款金额', icon: 'none' })
        return
      }
      if (amountCents > maxCents) {
        void Taro.showToast({ title: '金额超过可退上限', icon: 'none' })
        return
      }
    }
    submitMutation.mutate()
  }

  if (orderQuery.isLoading || !order) {
    return (
      <View className='page-shell aftersale-apply-page'>
        <Skeleton animated rows={6} />
      </View>
    )
  }

  return (
    <View className='page-shell aftersale-apply-page'>
      {/* 售后类型 */}
      <View className='aftersale-apply-page__section'>
        <Text className='aftersale-apply-page__section-title'>选择售后类型</Text>
        {TYPE_OPTIONS.map((opt) => {
          const active = opt.value === type
          return (
            <View
              key={opt.value}
              className='aftersale-apply-page__radio-row'
              onClick={() => setType(opt.value)}
            >
              <View
                className={`aftersale-apply-page__radio-dot${
                  active ? ' aftersale-apply-page__radio-dot--active' : ''
                }`}
              />
              <View className='aftersale-apply-page__radio-label'>
                <Text>{opt.label}</Text>
                <View>
                  <Text className='aftersale-apply-page__radio-desc'>{opt.desc}</Text>
                </View>
              </View>
            </View>
          )
        })}
      </View>

      {/* 退款金额 */}
      {type !== 'exchange' && (
        <View className='aftersale-apply-page__section'>
          <View className='aftersale-apply-page__amount-row'>
            <Text>退款金额（元）</Text>
            <View className='aftersale-apply-page__amount-input'>
              <Input
                type='digit'
                placeholder='0.00'
                value={amountStr}
                onChange={(val) => setAmountStr(String(val))}
                align='right'
              />
            </View>
          </View>
          <Text className='aftersale-apply-page__amount-cap'>
            可退上限 ¥{(maxCents / 100).toFixed(2)}
          </Text>
        </View>
      )}

      {/* 申请原因 */}
      <View className='aftersale-apply-page__section'>
        <View className='aftersale-apply-page__field'>
          <Text className='aftersale-apply-page__field-label'>申请原因（必填）</Text>
          <TextArea
            placeholder='请描述问题，便于商家快速处理'
            value={reason}
            onChange={(v) => setReason(String(v))}
            maxLength={200}
            rows={4}
          />
        </View>
      </View>

      {/* 凭证 */}
      <View className='aftersale-apply-page__section'>
        <View className='aftersale-apply-page__field'>
          <Text className='aftersale-apply-page__field-label'>上传凭证（选填，最多 6 张）</Text>
          <EvidenceUploader value={evidence} onChange={setEvidence} max={6} />
        </View>
      </View>

      {/* 提交 */}
      <View className='aftersale-apply-page__submit-bar'>
        <Button
          type='primary'
          block
          loading={submitMutation.isPending}
          onClick={handleSubmit}
        >
          提交申请
        </Button>
        <SafeArea position='bottom' />
      </View>
    </View>
  )
}
