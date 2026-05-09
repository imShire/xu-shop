import { RichText, View } from '@tarojs/components'
import type { RichTextData } from '@/services/page-config'
import './RichTextModule.scss'

interface Props {
  data: RichTextData
}

export default function RichTextModule({ data }: Props) {
  const content = data.content ?? ''
  if (!content) return null

  return (
    <View className='rich-text-module'>
      <RichText nodes={content} />
    </View>
  )
}
