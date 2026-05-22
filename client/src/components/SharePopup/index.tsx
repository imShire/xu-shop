import { useState } from 'react'
import { Button as TaroButton, View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation } from '@tanstack/react-query'
import { Popup } from '@/ui/nutui'
import { createShareLink, type ShareScene } from '@/services/share'
import { showErrorToast } from '@/utils/error'
import { isH5, isWeapp } from '@/utils/platform'
import './index.scss'

interface SharePopupProps {
  visible: boolean
  onClose: () => void
  scene: ShareScene
  targetId?: string | null
  /** 海报跳转路径（不传则不显示海报按钮） */
  posterPath?: string
}

/**
 * Cross-platform share popup.
 *
 * - **weapp**: 系统分享（好友 / 朋友圈通过 openType="share"），保存海报跳转海报页。
 *   实际分享内容由当前页面 `useShareAppMessage` 决定。
 * - **h5**: 调 `/c/share/links` 拿短链 → 复制 / 浏览器原生分享 / 海报。
 */
export default function SharePopup({
  visible,
  onClose,
  scene,
  targetId,
  posterPath,
}: SharePopupProps) {
  const [shortUrl, setShortUrl] = useState('')

  const linkMutation = useMutation({
    mutationFn: () =>
      createShareLink({
        scene,
        target_id: targetId ?? null,
        channel_code: isWeapp ? 'wxapp' : 'h5',
      }),
    onSuccess: (link) => {
      setShortUrl(link.h5_url)
    },
  })

  async function ensureLink() {
    if (shortUrl) return shortUrl
    const link = await linkMutation.mutateAsync()
    return link.h5_url
  }

  async function handleCopy() {
    try {
      const url = await ensureLink()
      await Taro.setClipboardData({ data: url })
      Taro.showToast({ title: '链接已复制', icon: 'success' })
      onClose()
    } catch (error) {
      showErrorToast(error, '获取链接失败')
    }
  }

  async function handleH5Native() {
    if (!isH5 || typeof navigator === 'undefined') return
    try {
      const url = await ensureLink()
      const nav = navigator as Navigator & {
        share?: (data: { title?: string; url?: string }) => Promise<void>
      }
      if (typeof nav.share === 'function') {
        await nav.share({ title: '徐记小铺', url })
      } else {
        await Taro.setClipboardData({ data: url })
        Taro.showToast({ title: '已复制链接', icon: 'success' })
      }
      onClose()
    } catch (error) {
      // user cancelled native share — silently ignore
      if ((error as { name?: string })?.name !== 'AbortError') {
        showErrorToast(error, '分享失败')
      }
    }
  }

  function handlePoster() {
    if (!posterPath) return
    onClose()
    void Taro.navigateTo({ url: posterPath })
  }

  return (
    <Popup visible={visible} position='bottom' onClose={onClose} round>
      <View className='share-popup'>
        <View className='share-popup__title'>分享给好友</View>

        <View className='share-popup__grid'>
          {isWeapp ? (
            <>
              <TaroButton
                className='share-popup__cell share-popup__btn-reset'
                openType='share'
                onClick={() => onClose()}
              >
                <View className='share-popup__icon share-popup__icon--wechat'>微</View>
                <Text className='share-popup__label'>微信好友</Text>
              </TaroButton>
              <View className='share-popup__cell' onClick={() => void handleCopy()}>
                <View className='share-popup__icon share-popup__icon--copy'>链</View>
                <Text className='share-popup__label'>复制链接</Text>
              </View>
            </>
          ) : (
            <>
              <View className='share-popup__cell' onClick={() => void handleH5Native()}>
                <View className='share-popup__icon share-popup__icon--wechat'>分</View>
                <Text className='share-popup__label'>系统分享</Text>
              </View>
              <View className='share-popup__cell' onClick={() => void handleCopy()}>
                <View className='share-popup__icon share-popup__icon--copy'>链</View>
                <Text className='share-popup__label'>复制链接</Text>
              </View>
            </>
          )}

          {posterPath ? (
            <View className='share-popup__cell' onClick={handlePoster}>
              <View className='share-popup__icon share-popup__icon--poster'>图</View>
              <Text className='share-popup__label'>生成海报</Text>
            </View>
          ) : null}
        </View>

        {shortUrl ? (
          <View className='share-popup__url'>
            <Text className='share-popup__url-text'>{shortUrl}</Text>
          </View>
        ) : null}

        <View className='share-popup__cancel' onClick={onClose}>
          取消
        </View>
      </View>
    </Popup>
  )
}
