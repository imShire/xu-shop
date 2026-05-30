import DOMPurify from 'isomorphic-dompurify'

const ALLOWED_TAGS = [
  'p', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'ul', 'ol', 'li',
  'strong', 'em', 'br',
  'a', 'blockquote', 'span', 'div',
  'table', 'thead', 'tbody', 'tr', 'td', 'th',
  'img',
]

const ALLOWED_ATTR = ['href', 'target', 'rel', 'src', 'alt', 'title', 'class']

const FORBID_TAGS = ['script', 'iframe', 'object', 'embed', 'form', 'svg', 'math', 'style', 'link']

export function sanitizeHtml(html: string): string {
  if (!html) return ''

  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS,
    ALLOWED_ATTR,
    FORBID_TAGS,
    ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel|data:image\/(?:png|jpeg|gif|webp));|[^a-z]|[a-z+.-]+(?:[^a-z+.\-:]|$))/i,
    KEEP_CONTENT: true,
  })
}
