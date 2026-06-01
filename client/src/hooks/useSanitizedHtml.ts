// H5 走 DOMPurify 净化；weapp 直通（小程序沙箱安全，由 rich-text 处理）
import { isH5 } from '@/utils/platform'
import { sanitizeHtml } from '@/utils/sanitize'

/**
 * 返回净化后的 HTML 字符串。
 *   - H5：调用 DOMPurify 白名单
 *   - weapp：直通，调用方应当用 `<RichText nodes={raw} />`
 */
export function useSanitizedHtml(raw: string | null | undefined): string {
  if (!raw) return ''
  if (isH5) return sanitizeHtml(raw)
  return raw
}
