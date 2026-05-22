import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useQuery } from '@tanstack/react-query'
import EmptyState from '@/components/EmptyState'
import { getMyDistributor } from '@/services/distributor'
import { useAuthStore } from '@/stores/auth'
import { Cell, Skeleton } from '@/ui/nutui'
import { formatYuan } from '@/utils/price'
import './index.scss'

export default function DistributorCenterPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const profileQuery = useQuery({
    queryKey: ['my-distributor'],
    queryFn: getMyDistributor,
    enabled: isLoggedIn,
  })

  if (!isLoggedIn) {
    return (
      <View className='dist-center'>
        <EmptyState
          title='请先登录'
          action={{
            text: '去登录',
            onClick: () =>
              void Taro.navigateTo({ url: '/pages/auth/login/index?redirect=/pages/distributor/center/index' }),
          }}
        />
      </View>
    )
  }

  if (profileQuery.isLoading) {
    return (
      <View className='dist-center'>
        <Skeleton animated rows={6} />
      </View>
    )
  }

  const profile = profileQuery.data
  if (!profile) {
    return (
      <View className='dist-center'>
        <EmptyState
          title='您还不是分销员'
          description='申请成为分销员，开启分享赚佣金'
          action={{
            text: '立即申请',
            onClick: () => void Taro.navigateTo({ url: '/pages/distributor/apply/index' }),
          }}
        />
      </View>
    )
  }

  if (profile.status === 'pending') {
    return (
      <View className='dist-center'>
        <View className='dist-center__pending'>
          您的分销员申请正在审核中，通常 1-3 个工作日。
        </View>
      </View>
    )
  }

  if (profile.status === 'disabled') {
    return (
      <View className='dist-center'>
        <View className='dist-center__pending'>
          您的分销员资格已停用，如有疑问请联系客服。
        </View>
      </View>
    )
  }

  return (
    <View className='dist-center'>
      <View className='dist-center__hero'>
        <View className='dist-center__hero-row'>
          <View>
            <Text className='dist-center__nickname'>{profile.nickname}</Text>
            <Text className='dist-center__level'>{profile.level === 'senior' ? '高级' : '普通'}分销员</Text>
          </View>
        </View>
        <Text className='dist-center__balance-label'>可提现佣金</Text>
        <Text className='dist-center__balance-value'>
          {formatYuan(profile.available_commission_cents)}
        </Text>
        <View
          className='dist-center__withdraw-btn'
          onClick={() => void Taro.navigateTo({ url: '/pages/distributor/withdraw/index' })}
        >
          立即提现
        </View>
      </View>

      <View className='dist-center__metrics'>
        <View className='dist-center__metric'>
          <Text className='dist-center__metric-num'>{formatYuan(profile.total_commission_cents)}</Text>
          <Text className='dist-center__metric-label'>累计佣金</Text>
        </View>
        <View className='dist-center__metric'>
          <Text className='dist-center__metric-num'>{formatYuan(profile.withdrawing_cents)}</Text>
          <Text className='dist-center__metric-label'>提现中</Text>
        </View>
        <View className='dist-center__metric'>
          <Text className='dist-center__metric-num'>{formatYuan(profile.withdrawn_cents)}</Text>
          <Text className='dist-center__metric-label'>已到账</Text>
        </View>
      </View>

      <View className='dist-center__entries'>
        <Cell
          title='佣金记录'
          extra={`查看 →`}
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/distributor/commissions/index' })}
        />
        <Cell
          title='提现记录'
          extra='查看 →'
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/distributor/withdraws/index' })}
        />
        <Cell
          title='推广数据'
          extra={`点击 ${profile.share_click_count} / 注册 ${profile.share_register_count} / 订单 ${profile.share_order_count}`}
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/distributor/funnel/index' })}
        />
        <Cell
          title='生成推广海报'
          extra='去生成 →'
          clickable
          onClick={() =>
            void Taro.navigateTo({ url: '/pages/distributor/poster/index?scene=invite_register' })
          }
        />
      </View>
    </View>
  )
}
