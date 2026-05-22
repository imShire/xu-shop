import { Text, View } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'
import { getPointTransactions } from '@/services/point'
import { useAuthStore } from '@/stores/auth'
import type { PointTransaction } from '@/types/biz'
import { Skeleton } from '@/ui/nutui'
import './index.scss'

const TYPE_LABEL: Record<PointTransaction['type'], string> = {
  earn: '获得',
  spend: '消费抵扣',
  expire: '过期清零',
  refund: '退款返还',
  admin_adjust: '人工调整',
  freeze: '冻结',
  unfreeze: '解冻',
}

export default function PointTransactionsPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const listQuery = useQuery({
    queryKey: ['point-tx'],
    queryFn: () => getPointTransactions({ page: 1, page_size: 50 }),
    enabled: isLoggedIn,
  })

  if (listQuery.isLoading) {
    return (
      <View className='point-tx'>
        <Skeleton animated rows={6} />
      </View>
    )
  }

  const list = listQuery.data?.list ?? []
  if (list.length === 0) {
    return (
      <View className='point-tx'>
        <EmptyState title='暂无积分流水' />
      </View>
    )
  }

  return (
    <View className='point-tx'>
      {list.map((tx) => (
        <View key={tx.id} className='point-tx__row'>
          <View className='point-tx__main'>
            <Text className='point-tx__type'>{TYPE_LABEL[tx.type] ?? tx.type}</Text>
            <Text className='point-tx__reason'>{tx.reason}</Text>
            <Text className='point-tx__time'>{tx.created_at?.replace('T', ' ').slice(0, 16)}</Text>
          </View>
          <View>
            <Text
              className={`point-tx__amount ${tx.change >= 0 ? 'point-tx__amount--pos' : 'point-tx__amount--neg'}`}
            >
              {tx.change >= 0 ? '+' : ''}
              {tx.change}
            </Text>
            <Text className='point-tx__balance'>余额 {tx.balance_after}</Text>
          </View>
        </View>
      ))}
    </View>
  )
}
