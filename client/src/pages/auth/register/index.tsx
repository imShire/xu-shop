import { useState, useRef } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { sendSmsCode, phoneRegister } from '@/services/auth'
import { useAuthStore } from '@/stores/auth'
import { Button, Input } from '@/ui/nutui'
import { navigateAfterAuth, withAuthRedirect } from '@/utils/auth-navigation'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function RegisterPage() {
  const params = Taro.getCurrentInstance().router?.params
  const redirect = typeof params?.redirect === 'string' ? params.redirect : undefined

  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [sending, setSending] = useState(false)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const { setAuth } = useAuthStore()

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
      await sendSmsCode(phone.trim(), 'register')
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

  async function handleRegister() {
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
    if (!password) {
      Taro.showToast({ title: '请输入密码', icon: 'none' })
      return
    }
    if (password.length < 8) {
      Taro.showToast({ title: '密码至少 8 位', icon: 'none' })
      return
    }
    if (password !== confirmPassword) {
      Taro.showToast({ title: '两次密码不一致', icon: 'none' })
      return
    }

    setLoading(true)
    try {
      const res = await phoneRegister({ phone: phone.trim(), code: code.trim(), password })
      setAuth(res.token, res.user)
      Taro.showToast({ title: '注册成功', icon: 'success' })
      setTimeout(() => navigateAfterAuth(redirect), 800)
    } catch (error) {
      showErrorToast(error, '注册失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='page-shell register-page'>
      <View className='register-page__header'>
        <Text className='register-page__title'>创建账号</Text>
        <Text className='register-page__subtitle'>注册后即可享受完整购物体验</Text>
      </View>

      <View className='register-page__form'>
        <View className='register-page__input-wrap'>
          <Text className='register-page__label'>手机号</Text>
          <Input
            className='register-page__input'
            type='tel'
            placeholder='请输入手机号'
            value={phone}
            onChange={(val) => setPhone(val)}
          />
        </View>

        <View className='register-page__input-wrap'>
          <Text className='register-page__label'>验证码</Text>
          <View className='register-page__code-row'>
            <Input
              className='register-page__input register-page__code-input'
              type='number'
              placeholder='请输入验证码'
              value={code}
              onChange={(val) => setCode(val)}
            />
            <Button
              type='primary'
              plain
              className='register-page__send-btn'
              loading={sending}
              disabled={sending || countdown > 0}
              onClick={() => void handleSendCode()}
            >
              {countdown > 0 ? `${countdown}s 后重发` : '获取验证码'}
            </Button>
          </View>
        </View>

        <View className='register-page__input-wrap'>
          <Text className='register-page__label'>密码</Text>
          <Input
            className='register-page__input'
            type='password'
            placeholder='请设置密码（至少 8 位）'
            value={password}
            onChange={(val) => setPassword(val)}
          />
        </View>

        <View className='register-page__input-wrap'>
          <Text className='register-page__label'>确认密码</Text>
          <Input
            className='register-page__input'
            type='password'
            placeholder='请再次输入密码'
            value={confirmPassword}
            onChange={(val) => setConfirmPassword(val)}
          />
        </View>

        <Button
          type='primary'
          className='register-page__submit'
          loading={loading}
          disabled={loading}
          onClick={() => void handleRegister()}
        >
          注册
        </Button>

        <View className='register-page__footer'>
          <Text className='register-page__hint'>已有账号？</Text>
          <Text
            className='register-page__link'
            onClick={() => void Taro.redirectTo({ url: withAuthRedirect('/pages/auth/login/index', redirect) })}
          >
            返回登录
          </Text>
        </View>
      </View>
    </View>
  )
}
