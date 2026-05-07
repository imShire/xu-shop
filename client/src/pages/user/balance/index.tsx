import { View, Text } from '@tarojs/components'
import { useLoad } from '@tarojs/taro'
import { useState } from 'react'
import { Cell, Empty } from '@/ui/nutui'
import { getMyBalance } from '@/services/user'
import type { BalanceLog } from '@/services/user'
import './index.scss'

const TYPE_LABEL: Record<string, string> = {
  recharge: '充值',
  spend: '消费',
  refund: '退款',
}

export default function BalancePage() {
  const [balanceCents, setBalanceCents] = useState(0)
  const [logs, setLogs] = useState<BalanceLog[]>([])
  const [loading, setLoading] = useState(true)

  useLoad(async () => {
    try {
      const res = await getMyBalance({ page: 1, page_size: 50 })
      setBalanceCents(res.balance_cents)
      setLogs(res.logs ?? [])
    } finally {
      setLoading(false)
    }
  })

  return (
    <View className='balance-page'>
      <View className='balance-page__header'>
        <Text className='balance-page__header-label'>账户余额（元）</Text>
        <Text className='balance-page__header-amount'>
          ¥{(balanceCents / 100).toFixed(2)}
        </Text>
      </View>

      <View className='balance-page__list'>
        {!loading && logs.length === 0 ? (
          <Empty description='暂无流水记录' />
        ) : (
          logs.map((log) => (
            <Cell
              key={log.id}
              title={TYPE_LABEL[log.type] ?? log.type}
              description={log.created_at?.slice(0, 10)}
              extra={
                <Text
                  style={{
                    color: log.change_cents >= 0 ? '#52c41a' : '#f5222d',
                    fontWeight: 'bold',
                  }}
                >
                  {log.change_cents >= 0 ? '+' : ''}
                  {(log.change_cents / 100).toFixed(2)}
                </Text>
              }
            />
          ))
        )}
      </View>
    </View>
  )
}
