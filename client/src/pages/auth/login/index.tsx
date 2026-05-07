import { useState } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { phoneLogin } from '@/services/auth'
import { useAuthStore } from '@/stores/auth'
import { Button, Input } from '@/ui/nutui'
import { navigateAfterAuth, withAuthRedirect } from '@/utils/auth-navigation'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function LoginPage() {
  const params = Taro.getCurrentInstance().router?.params
  const redirect = typeof params?.redirect === 'string' ? params.redirect : undefined

  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  const { setAuth } = useAuthStore()

  async function handlePhoneLogin() {
    if (!phone.trim()) {
      Taro.showToast({ title: '请输入手机号', icon: 'none' })
      return
    }
    if (!/^1[3-9]\d{9}$/.test(phone.trim())) {
      Taro.showToast({ title: '手机号格式不正确', icon: 'none' })
      return
    }
    if (!password) {
      Taro.showToast({ title: '请输入密码', icon: 'none' })
      return
    }

    setLoading(true)
    try {
      const res = await phoneLogin({ phone: phone.trim(), password })
      setAuth(res.token, res.user)
      Taro.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => navigateAfterAuth(redirect), 800)
    } catch (error) {
      showErrorToast(error, '登录失败，请检查手机号和密码')
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='page-shell login-page'>
      <View className='login-page__logo'>
        <Text className='login-page__brand'>徐记小铺</Text>
        <Text className='login-page__tagline'>登录账号，享受专属服务</Text>
      </View>

      <View className='login-page__form'>
        <View className='login-page__input-wrap'>
          <Text className='login-page__label'>手机号</Text>
          <Input
            className='login-page__input'
            type='tel'
            placeholder='请输入手机号'
            value={phone}
            onChange={(val) => setPhone(val)}
          />
        </View>

        <View className='login-page__input-wrap'>
          <Text className='login-page__label'>密码</Text>
          <Input
            className='login-page__input'
            type='password'
            placeholder='请输入密码'
            value={password}
            onChange={(val) => setPassword(val)}
          />
        </View>

        <View className='login-page__links'>
          <Text
            className='login-page__link'
            onClick={() => void Taro.navigateTo({ url: withAuthRedirect('/pages/auth/register/index', redirect) })}
          >
            没有账号？立即注册
          </Text>
          <Text
            className='login-page__link'
            onClick={() => void Taro.navigateTo({ url: withAuthRedirect('/pages/auth/reset-password/index', redirect) })}
          >
            忘记密码
          </Text>
        </View>

        <Button
          type='primary'
          className='login-page__submit'
          loading={loading}
          disabled={loading}
          onClick={() => void handlePhoneLogin()}
        >
          登录
        </Button>
      </View>
    </View>
  )
}
