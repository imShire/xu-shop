import { Text, View } from '@tarojs/components'
import { useQuery } from '@tanstack/react-query'
import { getMyMemberLevel } from '@/services/member'
import { useAuthStore } from '@/stores/auth'
import { Skeleton } from '@/ui/nutui'
import { formatYuan } from '@/utils/price'
import './index.scss'

const DEFAULT_BENEFITS = ['专属会员价', '生日礼券', '积分倍率加成', '专属客服通道']

export default function MemberCenterPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const levelQuery = useQuery({
    queryKey: ['my-level'],
    queryFn: getMyMemberLevel,
    enabled: isLoggedIn,
  })

  if (!isLoggedIn || levelQuery.isLoading) {
    return (
      <View className='member-page'>
        <Skeleton animated rows={6} />
      </View>
    )
  }

  const data = levelQuery.data
  const current = data?.current
  const next = data?.next
  const progress = Math.max(0, Math.min(1, data?.progress ?? 0))

  const benefitsObj = (current?.benefits ?? {}) as Record<string, unknown>
  const benefits =
    Array.isArray(benefitsObj.descriptions)
      ? (benefitsObj.descriptions as string[])
      : DEFAULT_BENEFITS

  return (
    <View className='member-page'>
      <View className='member-page__hero'>
        <Text className='member-page__level'>当前等级</Text>
        <Text className='member-page__name'>{current?.name ?? '普通会员'}</Text>
        <Text>积分倍率 {(current?.points_multiplier ?? 1).toFixed(1)}x</Text>

        <View className='member-page__progress'>
          <View className='member-page__progress-bar'>
            <View
              className='member-page__progress-fill'
              style={{ width: `${(progress * 100).toFixed(1)}%` }}
            />
          </View>
          <View className='member-page__progress-meta'>
            <Text>已累计 {formatYuan(data?.cumulative_amount_cents ?? 0)}</Text>
            <Text>
              {next
                ? `还差 ${formatYuan(data?.to_next_cents ?? 0)} 升 ${next.name}`
                : '已达最高等级'}
            </Text>
          </View>
        </View>
      </View>

      <View className='member-page__benefits'>
        <Text className='member-page__section-title'>会员专属权益</Text>
        {benefits.map((b, idx) => (
          <View key={`${b}-${idx}`} className='member-page__benefit-row'>
            <View className='member-page__benefit-dot' />
            <Text className='member-page__benefit-text'>{b}</Text>
          </View>
        ))}
      </View>
    </View>
  )
}
