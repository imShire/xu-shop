import Taro from '@tarojs/taro'

const DEFAULT_REDIRECT = '/pages/user/index/index'

const TAB_PAGES = new Set([
  '/pages/home/index',
  '/pages/category/index',
  '/pages/cart/index',
  '/pages/user/index/index',
])

function safeDecode(value?: string) {
  if (!value) {
    return ''
  }

  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

export function normalizeAuthRedirect(redirect?: string) {
  const target = safeDecode(redirect) || DEFAULT_REDIRECT

  if (target.startsWith('/pages/auth/')) {
    return DEFAULT_REDIRECT
  }

  return target
}

export function withAuthRedirect(url: string, redirect?: string) {
  const target = normalizeAuthRedirect(redirect)
  const separator = url.includes('?') ? '&' : '?'
  return `${url}${separator}redirect=${encodeURIComponent(target)}`
}

export function navigateAfterAuth(redirect?: string) {
  const target = normalizeAuthRedirect(redirect)

  if (TAB_PAGES.has(target)) {
    void Taro.switchTab({ url: target })
    return
  }

  void Taro.redirectTo({ url: target })
}
