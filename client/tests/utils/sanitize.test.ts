import { describe, expect, it } from 'vitest'
import { sanitizeHtml } from '@/utils/sanitize'

describe('sanitizeHtml', () => {
  it('拦截 <script>', () => {
    const out = sanitizeHtml('<p>hi</p><script>alert(1)</script>')
    expect(out).toContain('<p>hi</p>')
    expect(out.toLowerCase()).not.toContain('<script')
    expect(out.toLowerCase()).not.toContain('alert(1)')
  })

  it('拦截 onerror 属性', () => {
    const out = sanitizeHtml('<img src="x" onerror="alert(1)" />')
    expect(out.toLowerCase()).not.toContain('onerror')
    expect(out.toLowerCase()).not.toContain('alert(1)')
  })

  it('拦截 <iframe>', () => {
    const out = sanitizeHtml('<iframe src="http://evil"></iframe>')
    expect(out.toLowerCase()).not.toContain('<iframe')
  })

  it('保留白名单标签', () => {
    const out = sanitizeHtml('<p><strong>bold</strong> and <em>em</em></p>')
    expect(out).toContain('<p>')
    expect(out).toContain('<strong>bold</strong>')
    expect(out).toContain('<em>em</em>')
  })

  it('空输入返回空字符串', () => {
    expect(sanitizeHtml('')).toBe('')
  })
})
