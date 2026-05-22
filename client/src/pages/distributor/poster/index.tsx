import { useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation } from '@tanstack/react-query'
import { createShareLink, generatePoster } from '@/services/share'
import type { ShareScene } from '@/services/share'
import { Button } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import { isWeapp } from '@/utils/platform'
import './index.scss'

export default function PosterPage() {
  const params = Taro.getCurrentInstance().router?.params ?? {}
  const scene = ((params as Record<string, string>).scene as ShareScene) || 'invite_register'
  const targetId = (params as Record<string, string>).target_id || null

  const [posterUrl, setPosterUrl] = useState('')
  const [generating, setGenerating] = useState(false)

  async function handleGenerate() {
    setGenerating(true)
    try {
      const link = await createShareLink({
        scene,
        target_id: targetId,
        channel_code: isWeapp ? 'wxapp' : 'h5',
      })
      const poster = await generatePoster(link.id)
      setPosterUrl(poster.poster_url)
    } catch (e) {
      showErrorToast(e, '生成失败')
    } finally {
      setGenerating(false)
    }
  }

  const saveMutation = useMutation({
    mutationFn: async () => {
      if (!posterUrl) throw new Error('请先生成海报')
      if (isWeapp) {
        const dl = await Taro.downloadFile({ url: posterUrl })
        if (dl.statusCode !== 200) throw new Error('下载失败')
        await Taro.saveImageToPhotosAlbum({ filePath: dl.tempFilePath })
      } else {
        // H5: open in new tab so user can long-press to save.
        if (typeof window !== 'undefined') {
          window.open(posterUrl, '_blank')
        }
      }
    },
    onSuccess: () => Taro.showToast({ title: '操作成功', icon: 'success' }),
    onError: (e) => showErrorToast(e, '保存失败'),
  })

  return (
    <View className='dist-poster'>
      {posterUrl ? (
        <Image className='dist-poster__image' src={posterUrl} mode='widthFix' />
      ) : (
        <View className='dist-poster__placeholder'>
          <Text>点击下方按钮生成海报</Text>
        </View>
      )}

      <Text className='dist-poster__hint'>
        {isWeapp ? '生成后长按或点击「保存到相册」' : '生成后在新窗口长按图片即可保存'}
      </Text>

      <View className='dist-poster__actions'>
        <Button block loading={generating} onClick={() => void handleGenerate()}>
          {posterUrl ? '重新生成' : '生成海报'}
        </Button>
        {posterUrl ? (
          <Button
            type='primary'
            block
            loading={saveMutation.isPending}
            onClick={() => saveMutation.mutate()}
          >
            {isWeapp ? '保存到相册' : '在新窗口打开'}
          </Button>
        ) : null}
      </View>
    </View>
  )
}
