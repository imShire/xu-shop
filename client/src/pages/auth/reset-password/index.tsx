import { useState, useRef } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { sendSmsCode, resetPassword } from '@/services/auth'
import { Button, Input } from '@/ui/nutui'
import { withAuthRedirect } from '@/utils/auth-navigation'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function ResetPasswordPage() {
  const params = Taro.getCurrentInstance().router?.params
  const redirect = typeof params?.redirect === 'string' ? params.redirect : undefined

  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [sending, setSending] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  async function handleSendCode() {
    if (!phone.trim()) {
      Taro.showToast({ title: '请输入手机号', icon: 'none' })
      return
    }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) {
      Taro.showToast({ title: '手机号格式不正确', icon: 'none' })
      return
    }

    setSending(true)
    try {
      await sendSmsCode(phone.trim(), 'reset')
      Taro.showToast({ title: '验证码已发送', icon: 'success' })
      setCountdown(60)
      timerRef.current = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) {
            if (timerRef.current) {
              clearInterval(timerRef.current)
              timerRef.current = null
            }
            return 0
          }
          return c - 1
        })
      }, 1000)
    } catch (error) {
      showErrorToast(error, '验证码发送失败')
    } finally {
      setSending(false)
    }
  }

  async function handleReset() {
    if (!phone.trim()) {
      Taro.showToast({ title: '请输入手机号', icon: 'none' })
      return
    }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) {
      Taro.showToast({ title: '手机号格式不正确', icon: 'none' })
      return
    }
    if (!code.trim()) {
      Taro.showToast({ title: '请输入验证码', icon: 'none' })
      return
    }
    if (!newPassword) {
      Taro.showToast({ title: '请输入新密码', icon: 'none' })
      return
    }
    if (newPassword.length < 8) {
      Taro.showToast({ title: '密码至少 8 位', icon: 'none' })
      return
    }
    if (newPassword !== confirmPassword) {
      Taro.showToast({ title: '两次密码不一致', icon: 'none' })
      return
    }

    setLoading(true)
    try {
      await resetPassword({ phone: phone.trim(), code: code.trim(), new_password: newPassword })
      Taro.showToast({ title: '密码重置成功，请重新登录', icon: 'success' })
      setTimeout(() => {
        void Taro.redirectTo({ url: withAuthRedirect('/pages/auth/login/index', redirect) })
      }, 1200)
    } catch (error) {
      showErrorToast(error, '密码重置失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='page-shell reset-page'>
      <View className='reset-page__header'>
        <Text className='reset-page__title'>重置密码</Text>
        <Text className='reset-page__subtitle'>通过手机验证码重新设置密码</Text>
      </View>

      <View className='reset-page__form'>
        <View className='reset-page__input-wrap'>
          <Text className='reset-page__label'>手机号</Text>
          <Input
            className='reset-page__input'
            type='tel'
            placeholder='请输入注册时的手机号'
            value={phone}
            onChange={(val) => setPhone(val)}
          />
        </View>

        <View className='reset-page__input-wrap'>
          <Text className='reset-page__label'>验证码</Text>
          <View className='reset-page__code-row'>
            <Input
              className='reset-page__input reset-page__code-input'
              type='number'
              placeholder='请输入验证码'
              value={code}
              onChange={(val) => setCode(val)}
            />
            <Button
              type='primary'
              plain
              className='reset-page__send-btn'
              loading={sending}
              disabled={sending || countdown > 0}
              onClick={() => void handleSendCode()}
            >
              {countdown > 0 ? `${countdown}s 后重发` : '获取验证码'}
            </Button>
          </View>
        </View>

        <View className='reset-page__input-wrap'>
          <Text className='reset-page__label'>新密码</Text>
          <Input
            className='reset-page__input'
            type='password'
            placeholder='请设置新密码（至少 8 位）'
            value={newPassword}
            onChange={(val) => setNewPassword(val)}
          />
        </View>

        <View className='reset-page__input-wrap'>
          <Text className='reset-page__label'>确认新密码</Text>
          <Input
            className='reset-page__input'
            type='password'
            placeholder='请再次输入新密码'
            value={confirmPassword}
            onChange={(val) => setConfirmPassword(val)}
          />
        </View>

        <Button
          type='primary'
          className='reset-page__submit'
          loading={loading}
          disabled={loading}
          onClick={() => void handleReset()}
        >
          确认重置
        </Button>

        <View className='reset-page__footer'>
          <Text className='reset-page__hint'>想起密码了？</Text>
          <Text
            className='reset-page__link'
            onClick={() => void Taro.navigateBack()}
          >
            返回登录
          </Text>
        </View>
      </View>
    </View>
  )
}
