import { View, ScrollView } from '@tarojs/components'
import { type ReactNode, useRef, useCallback } from 'react'
import { Skeleton } from '@/ui/nutui'
import EmptyState from '@/components/EmptyState'
import './index.scss'

interface PullListProps<T> {
  data: T[]
  loading: boolean
  hasMore?: boolean
  onLoadMore?: () => void
  renderItem: (item: T, index: number) => ReactNode
  keyExtractor?: (item: T, index: number) => string
  emptyTitle?: string
  emptyDescription?: string
  emptyActionText?: string
  onEmptyAction?: () => void
  skeletonRows?: number
  className?: string
}

export default function PullList<T>({
  data,
  loading,
  hasMore = false,
  onLoadMore,
  renderItem,
  keyExtractor,
  emptyTitle = '暂无内容',
  emptyDescription,
  emptyActionText,
  onEmptyAction,
  skeletonRows = 3,
  className = '',
}: PullListProps<T>) {
  const loadingRef = useRef(false)

  const handleScrollToLower = useCallback(() => {
    if (loadingRef.current || !hasMore || !onLoadMore) return
    loadingRef.current = true
    onLoadMore()
    setTimeout(() => { loadingRef.current = false }, 1000)
  }, [hasMore, onLoadMore])

  if (loading && data.length === 0) {
    return (
      <View className={`pull-list ${className}`}>
        {Array.from({ length: skeletonRows }).map((_, i) => (
          <Skeleton key={i} animated rows={3} />
        ))}
      </View>
    )
  }

  if (!loading && data.length === 0) {
    return (
      <EmptyState
        title={emptyTitle}
        description={emptyDescription ?? ''}
        actionText={emptyActionText}
        onAction={onEmptyAction}
      />
    )
  }

  return (
    <ScrollView
      scrollY
      className={`pull-list ${className}`}
      onScrollToLower={handleScrollToLower}
      lowerThreshold={80}
    >
      {data.map((item, index) => (
        <View key={keyExtractor ? keyExtractor(item, index) : `item-${index}`}>
          {renderItem(item, index)}
        </View>
      ))}
      {hasMore && loading && <Skeleton animated rows={2} />}
    </ScrollView>
  )
}
