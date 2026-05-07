import { Empty } from '@/ui/nutui'
import './index.scss'

interface EmptyStateProps {
  title?: string
  description?: string
  // Unified action object API
  action?: { text: string; onClick: () => void }
  // Legacy flat props (kept for backward compat with existing callers)
  actionText?: string
  onAction?: () => void
}

export default function EmptyState({
  title = '暂无内容',
  description = '',
  action,
  actionText,
  onAction,
}: EmptyStateProps) {
  // Merge: action object takes precedence over flat props
  const resolvedText = action?.text ?? actionText
  const resolvedHandler = action?.onClick ?? onAction

  const actions =
    resolvedText && resolvedHandler
      ? [
          {
            text: resolvedText,
            type: 'primary' as const,
            onClick: resolvedHandler as unknown as () => () => void,
          },
        ]
      : []

  return (
    <Empty
      className='empty-state'
      title={title}
      description={description}
      actions={actions}
    />
  )
}
