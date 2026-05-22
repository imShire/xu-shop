import { useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { applyDistributor } from '@/services/distributor'
import { Button, Input, TextArea } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function DistributorApplyPage() {
  const [realName, setRealName] = useState('')
  const [phone, setPhone] = useState('')
  const [reason, setReason] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit() {
    if (!realName.trim()) {
      Taro.showToast({ title: '请输入真实姓名', icon: 'none' })
      return
    }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) {
      Taro.showToast({ title: '请输入正确的手机号', icon: 'none' })
      return
    }
    const ok = await Taro.showModal({
      title: '确认提交申请',
      content: '提交后将进入审核流程，请确认您的资料无误。',
    })
    if (!ok.confirm) return

    setLoading(true)
    try {
      await applyDistributor({
        real_name: realName.trim(),
        phone: phone.trim(),
        reason: reason.trim() || undefined,
      })
      Taro.showToast({ title: '申请已提交，等待审核', icon: 'success' })
      setTimeout(() => {
        void Taro.redirectTo({ url: '/pages/distributor/center/index' })
      }, 1000)
    } catch (e) {
      showErrorToast(e, '提交失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='dist-apply'>
      <View className='dist-apply__intro'>
        <Text className='dist-apply__title'>成为分销员</Text>
        <Text className='dist-apply__sub'>
          分享好物，成单即得佣金。加入后可在分销中心查看推广数据。
        </Text>
      </View>

      <View className='dist-apply__form'>
        <View className='dist-apply__field'>
          <Text className='dist-apply__label'>真实姓名</Text>
          <View className='dist-apply__input'>
            <Input placeholder='请输入真实姓名' value={realName} onChange={(v) => setRealName(String(v))} />
          </View>
        </View>
        <View className='dist-apply__field'>
          <Text className='dist-apply__label'>联系手机号</Text>
          <View className='dist-apply__input'>
            <Input
              type='tel'
              placeholder='用于结算和审核联络'
              value={phone}
              onChange={(v) => setPhone(String(v))}
              maxLength={11}
            />
          </View>
        </View>
        <View className='dist-apply__field'>
          <Text className='dist-apply__label'>申请理由（选填）</Text>
          <TextArea
            placeholder='简单介绍一下您的推广渠道或想法'
            value={reason}
            onChange={(v) => setReason(String(v))}
            maxLength={200}
          />
        </View>

        <View className='dist-apply__terms'>
          提交即视为同意《分销员服务协议》。审核通常 1-3 个工作日。
        </View>

        <Button type='primary' block loading={loading} onClick={() => void handleSubmit()}>
          提交申请
        </Button>
      </View>
    </View>
  )
}
