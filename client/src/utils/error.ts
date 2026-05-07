import Taro from '@tarojs/taro'

interface ErrorLike {
  message?: string
}

export function showErrorToast(error: unknown, fallback = '请求失败，请稍后再试') {
  const message = typeof error === 'object' && error && 'message' in error
    ? (error as ErrorLike).message ?? fallback
    : fallback

  Taro.showToast({
    title: message,
    icon: 'none',
  })
}
