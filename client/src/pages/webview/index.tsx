import { WebView } from '@tarojs/components'
import Taro from '@tarojs/taro'
import './index.scss'

export default function WebviewPage() {
  const params = Taro.getCurrentInstance().router?.params
  const src = params?.src ? decodeURIComponent(String(params.src)) : ''

  if (!src) return null

  return <WebView src={src} />
}
