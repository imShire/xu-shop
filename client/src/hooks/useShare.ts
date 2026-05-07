import Taro, { useLoad, useShareAppMessage, useShareTimeline } from '@tarojs/taro'
import { appendShareUserId, persistShareUserId } from '@/utils/share-attr'

interface ShareOptions {
  title: string
  path: string
}

export function useShare(options: ShareOptions) {
  useLoad(() => {
    const params = Taro.getCurrentInstance().router?.params
    if (params?.suid) {
      persistShareUserId(String(params.suid))
    }
  })

  useShareAppMessage(() => ({
    title: options.title,
    path: appendShareUserId(options.path),
  }))

  useShareTimeline(() => ({
    title: options.title,
    query: appendShareUserId('').replace(/^\?/, ''),
  }))
}
