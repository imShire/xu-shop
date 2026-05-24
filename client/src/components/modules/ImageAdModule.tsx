import { Image, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import type { ImageAdData } from '@/services/page-config'
import './ImageAdModule.scss'

interface Props {
  data: ImageAdData
}

function navigate(url: string) {
  if (!url) return
  if (url.startsWith('/pages')) {
    void Taro.navigateTo({ url })
  } else {
    void Taro.navigateTo({ url: `/pages/webview/index?url=${encodeURIComponent(url)}` })
  }
}

export default function ImageAdModule({ data }: Props) {
  const { image_url, link_config } = data
  if (!image_url) return null
  const linkUrl = link_config?.url ?? ''
  return (
    <View
      className='image-ad-module'
      onClick={() => navigate(linkUrl)}
    >
      <Image
        className='image-ad-module__img'
        src={image_url}
        mode='widthFix'
        style={{ width: '100%' }}
      />
    </View>
  )
}
