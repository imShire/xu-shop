import { Image, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import type { CategoryEntryData } from '@/services/page-config'
import './CategoryEntryModule.scss'

interface Props {
  data: CategoryEntryData
}

function navigate(linkUrl: string) {
  if (!linkUrl) return
  if (linkUrl.startsWith('/pages')) {
    void Taro.navigateTo({ url: linkUrl })
  } else {
    void Taro.navigateTo({ url: `/pages/webview/index?url=${encodeURIComponent(linkUrl)}` })
  }
}

export default function CategoryEntryModule({ data }: Props) {
  const items = data.items ?? []
  if (items.length === 0) return null

  return (
    <View className='category-entry-module'>
      {items.map((item, index) => (
        <View
          key={index}
          className='category-entry-module__item'
          onClick={() => navigate(item.link_url)}
        >
          <View className='category-entry-module__circle'>
            <Image
              className='category-entry-module__image'
              src={item.image_url}
              mode='aspectFill'
            />
          </View>
          <Text className='category-entry-module__label'>{item.title}</Text>
        </View>
      ))}
    </View>
  )
}
