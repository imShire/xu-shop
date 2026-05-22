import { useState } from 'react'
import { Text, View } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'
import { getMyCommissions, type CommissionStatus } from '@/services/distributor'
import { useAuthStore } from '@/stores/auth'
import { Skeleton, Tabs, TabPane } from '@/ui/nutui'
import { formatYuan } from '@/utils/price'
import './index.scss'

const TABS: Array<{ value: CommissionStatus | 'all'; label: string }> = [
  { value: 'all', label: '全部' },
  { value: 'pending', label: '待结算' },
  { value: 'locked', label: '冻结中' },
  { value: 'settled', label: '已到账' },
  { value: 'canceled', label: '已取消' },
]

const STATUS_LABEL: Record<CommissionStatus, string> = {
  pending: '待结算',
  locked: '冻结中',
  settled: '已到账',
  canceled: '已取消',
  suspect: '风控核查',
}

export default function CommissionsPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const [tab, setTab] = useState<CommissionStatus | 'all'>('all')

  const listQuery = useQuery({
    queryKey: ['commissions', tab],
    queryFn: () =>
      getMyCommissions({
        status: tab === 'all' ? undefined : tab,
        page: 1,
        page_size: 50,
      }),
    enabled: isLoggedIn,
  })

  return (
    <View className='dist-comm'>
      <Tabs value={tab} onChange={(v) => setTab(v as CommissionStatus | 'all')}>
        {TABS.map((t) => (
          <TabPane key={t.value} title={t.label} value={t.value} />
        ))}
      </Tabs>

      {listQuery.isLoading ? (
        <Skeleton animated rows={6} />
      ) : (listQuery.data?.list ?? []).length === 0 ? (
        <EmptyState title='暂无佣金记录' />
      ) : (
        listQuery.data!.list.map((rec) => (
          <View key={rec.id} className='dist-comm__row'>
            <View className='dist-comm__main'>
              <Text className='dist-comm__order'>订单 {rec.order_no}</Text>
              <Text className='dist-comm__base'>
                按 {formatYuan(rec.base_amount_cents)} × {(rec.rate * 100).toFixed(1)}%
              </Text>
              <Text className='dist-comm__time'>
                {rec.created_at?.replace('T', ' ').slice(0, 16)}
              </Text>
            </View>
            <View>
              <Text className='dist-comm__amount'>+{formatYuan(rec.amount_cents)}</Text>
              <Text className='dist-comm__status'>{STATUS_LABEL[rec.status] ?? rec.status}</Text>
            </View>
          </View>
        ))
      )}
    </View>
  )
}
