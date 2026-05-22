import { Text, View } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import { getMyDistributor } from '@/services/distributor'
import { Skeleton } from '@/ui/nutui'
import { formatYuan } from '@/utils/price'
import './index.scss'

export default function FunnelPage() {
  const profileQuery = useQuery({
    queryKey: ['my-distributor'],
    queryFn: getMyDistributor,
  })

  if (profileQuery.isLoading) {
    return (
      <View className='dist-funnel'>
        <Skeleton animated rows={6} />
      </View>
    )
  }
  const p = profileQuery.data
  const click = p?.share_click_count ?? 0
  const reg = p?.share_register_count ?? 0
  const order = p?.share_order_count ?? 0
  const max = Math.max(click, 1)

  const stages = [
    { label: '点击', num: click },
    { label: '注册', num: reg },
    { label: '下单', num: order },
  ]

  return (
    <View className='dist-funnel'>
      <View className='dist-funnel__card'>
        <Text className='dist-funnel__title'>分享转化漏斗</Text>
        {stages.map((s) => (
          <View key={s.label} className='dist-funnel__row'>
            <Text className='dist-funnel__label'>{s.label}</Text>
            <View className='dist-funnel__bar-wrap'>
              <View
                className='dist-funnel__bar'
                style={{ width: `${(s.num / max) * 100}%` }}
              />
            </View>
            <Text className='dist-funnel__num'>{s.num}</Text>
          </View>
        ))}
      </View>

      <View className='dist-funnel__card'>
        <Text className='dist-funnel__title'>累计数据</Text>
        <View className='dist-funnel__total-row'>
          <Text className='dist-funnel__total-label'>邀请注册</Text>
          <Text className='dist-funnel__total-value'>{p?.invitee_count ?? 0} 人</Text>
        </View>
        <View className='dist-funnel__total-row'>
          <Text className='dist-funnel__total-label'>本月新增</Text>
          <Text className='dist-funnel__total-value'>{p?.monthly_invitee_count ?? 0} 人</Text>
        </View>
        <View className='dist-funnel__total-row'>
          <Text className='dist-funnel__total-label'>本月成单</Text>
          <Text className='dist-funnel__total-value'>{p?.monthly_order_count ?? 0} 单</Text>
        </View>
        <View className='dist-funnel__total-row'>
          <Text className='dist-funnel__total-label'>分享 GMV</Text>
          <Text className='dist-funnel__total-value'>
            {formatYuan(p?.share_gmv_cents ?? 0)}
          </Text>
        </View>
      </View>
    </View>
  )
}
