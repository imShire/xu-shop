import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import {
  CARRIER_OPTIONS,
  submitAftersaleExpress,
} from '@/services/aftersale'
import { Button, Input, SafeArea } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function AftersaleExpressPage() {
  const id = Taro.getCurrentInstance().router?.params?.id ?? ''
  const [carrier, setCarrier] = useState<string>('sf')
  const [waybill, setWaybill] = useState('')

  const mutation = useMutation({
    mutationFn: () =>
      submitAftersaleExpress(id, {
        carrier_code: carrier,
        waybill_no: waybill.trim(),
      }),
    onSuccess: () => {
      void Taro.showToast({ title: '已提交运单', icon: 'success' })
      setTimeout(() => {
        void Taro.redirectTo({ url: `/pages/aftersale/detail/index?id=${id}` })
      }, 500)
    },
    onError: (err) => showErrorToast(err, '提交失败，请稍后再试'),
  })

  function handleSubmit() {
    if (!id) {
      void Taro.showToast({ title: '缺少售后单参数', icon: 'none' })
      return
    }
    if (!carrier) {
      void Taro.showToast({ title: '请选择快递公司', icon: 'none' })
      return
    }
    const no = waybill.trim()
    if (!no || no.length < 4) {
      void Taro.showToast({ title: '请填写正确的运单号', icon: 'none' })
      return
    }
    mutation.mutate()
  }

  return (
    <View className='page-shell aftersale-express-page'>
      <View className='aftersale-express-page__section'>
        <Text className='aftersale-express-page__title'>选择快递公司</Text>
        <View className='aftersale-express-page__carrier-list'>
          {CARRIER_OPTIONS.map((c) => {
            const active = c.code === carrier
            return (
              <View
                key={c.code}
                className={`aftersale-express-page__carrier-chip${
                  active ? ' aftersale-express-page__carrier-chip--active' : ''
                }`}
                onClick={() => setCarrier(c.code)}
              >
                {c.name}
              </View>
            )
          })}
        </View>
      </View>

      <View className='aftersale-express-page__section'>
        <Text className='aftersale-express-page__title'>运单号</Text>
        <Input
          placeholder='请输入快递单号'
          value={waybill}
          onChange={(v) => setWaybill(String(v).trim())}
        />
        <Text className='aftersale-express-page__hint'>
          请确保信息准确，错误信息可能导致退款失败
        </Text>
      </View>

      <View className='aftersale-express-page__submit-bar'>
        <Button
          type='primary'
          block
          loading={mutation.isPending}
          onClick={handleSubmit}
        >
          提交
        </Button>
        <SafeArea position='bottom' />
      </View>
    </View>
  )
}
