import { Text, View } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'
import { getMyWithdraws } from '@/services/distributor'
import type { WithdrawStatus } from '@/types/biz'
import { Skeleton } from '@/ui/nutui'
import { formatYuan } from '@/utils/price'
import './index.scss'

const STATUS_LABEL: Record<WithdrawStatus, string> = {
  pending: '审核中',
  processing: '打款中',
  success: '已到账',
  failed: '已失败',
  canceled: '已取消',
}

export default function WithdrawListPage() {
  const listQuery = useQuery({
    queryKey: ['my-withdraws'],
    queryFn: () => getMyWithdraws({ page: 1, page_size: 50 }),
  })

  if (listQuery.isLoading) {
    return (
      <View className='dist-wd-list'>
        <Skeleton animated rows={6} />
      </View>
    )
  }
  const list = listQuery.data?.list ?? []
  if (list.length === 0) {
    return (
      <View className='dist-wd-list'>
        <EmptyState title='暂无提现记录' />
      </View>
    )
  }

  return (
    <View className='dist-wd-list'>
      {list.map((w) => (
        <View key={w.id} className='dist-wd-list__row'>
          <View className='dist-wd-list__main'>
            <Text className='dist-wd-list__no'>编号 {w.withdraw_no}</Text>
            <Text className='dist-wd-list__time'>
              {w.applied_at?.replace('T', ' ').slice(0, 16)}
            </Text>
            {w.status === 'failed' && w.fail_reason ? (
              <Text className='dist-wd-list__fail'>失败原因：{w.fail_reason}</Text>
            ) : null}
          </View>
          <View>
            <Text className='dist-wd-list__amount'>{formatYuan(w.amount_cents)}</Text>
            <Text className={`dist-wd-list__status dist-wd-list__status--${w.status}`}>
              {STATUS_LABEL[w.status] ?? w.status}
            </Text>
          </View>
        </View>
      ))}
    </View>
  )
}
