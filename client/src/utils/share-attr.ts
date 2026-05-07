import { storage } from '@/utils/storage'

const SHARE_USER_KEY = 'share_user_id'

export function getShareUserId() {
  return storage.get<string>(SHARE_USER_KEY)
}

export function persistShareUserId(value?: string | null) {
  if (!value) {
    return
  }

  storage.set(SHARE_USER_KEY, value)
}

export function appendShareUserId(path: string) {
  const shareUserId = getShareUserId()
  if (!shareUserId) {
    return path
  }

  const connector = path.includes('?') ? '&' : '?'
  return `${path}${connector}suid=${encodeURIComponent(shareUserId)}`
}
