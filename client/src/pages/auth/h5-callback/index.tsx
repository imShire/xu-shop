import { Text, View } from '@tarojs/components'
import Taro, { useLoad } from '@tarojs/taro'
import { useAuthStore } from '@/stores/auth'
import { Button, Loading, Tag } from '@/ui/nutui'
import { navigateAfterAuth, normalizeAuthRedirect } from '@/utils/auth-navigation'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function H5CallbackPage() {
  const hydrate = useAuthStore((state) => state.hydrate)

  useLoad(() => {
    void (async () => {
      try {
        const params = Taro.getCurrentInstance().router?.params
        const redirect = normalizeAuthRedirect(
          typeof params?.redirect === 'string' ? params.redirect : undefined
        )

        await hydrate()
        navigateAfterAuth(redirect)
      } catch (error) {
        showErrorToast(error, '登录流程启动失败')
      }
    })()
  })

  return (
    <View className='page-shell auth-page'>
      <Tag plain type='primary'>登录中</Tag>
      <Text className='auth-page__title'>正在完成登录</Text>
      <Text className='auth-page__description'>请稍候，登录完成后会自动返回。</Text>
      <View className='auth-page__status'>
        <Loading />
        <Text className='auth-page__status-text'>正在验证登录状态，请稍候。</Text>
      </View>
      <Button plain type='primary' onClick={() => Taro.switchTab({ url: '/pages/home/index' })}>
        返回首页
      </Button>
    </View>
  )
}
