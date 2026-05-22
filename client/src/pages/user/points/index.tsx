import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { dailySignIn, getPointSummary } from '@/services/point'
import { useAuthStore } from '@/stores/auth'
import { Cell, Skeleton } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

export default function UserPointsPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const queryClient = useQueryClient()

  const summaryQuery = useQuery({
    queryKey: ['point-summary'],
    queryFn: getPointSummary,
    enabled: isLoggedIn,
  })

  const signInMutation = useMutation({
    mutationFn: dailySignIn,
    onSuccess: (res) => {
      Taro.showToast({
        title: res.change ? `签到 +${res.change}` : '签到成功',
        icon: 'success',
      })
      void queryClient.invalidateQueries({ queryKey: ['point-summary'] })
    },
    onError: (e) => showErrorToast(e, '签到失败'),
  })

  if (!isLoggedIn || summaryQuery.isLoading) {
    return (
      <View className='user-points'>
        <Skeleton animated rows={4} />
      </View>
    )
  }

  const data = summaryQuery.data
  const expiringSoon = data?.expiring_soon ?? 0

  return (
    <View className='user-points'>
      <View className='user-points__hero'>
        <Text className='user-points__hero-label'>积分余额</Text>
        <Text className='user-points__hero-balance'>{data?.balance ?? 0}</Text>
        <View className='user-points__hero-meta'>
          <View className='user-points__hero-meta-item'>
            <Text className='user-points__hero-meta-num'>{data?.total_earned ?? 0}</Text>
            <Text>累计获得</Text>
          </View>
          <View className='user-points__hero-meta-item'>
            <Text className='user-points__hero-meta-num'>{data?.total_spent ?? 0}</Text>
            <Text>累计使用</Text>
          </View>
          <View className='user-points__hero-meta-item'>
            <Text className='user-points__hero-meta-num'>{data?.locked ?? 0}</Text>
            <Text>冻结中</Text>
          </View>
        </View>
      </View>

      {expiringSoon > 0 ? (
        <View className='user-points__alert'>
          您有 {expiringSoon} 积分即将过期
          {data?.expiring_at ? `（${data.expiring_at.slice(0, 10)} 前）` : ''}，请尽快使用。
        </View>
      ) : null}

      <View className='user-points__actions'>
        <Cell
          title='每日签到 +5'
          extra='立即签到'
          clickable
          onClick={() => signInMutation.mutate()}
        />
        <Cell
          title='积分流水'
          extra='查看 →'
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/user/points/transactions/index' })}
        />
        <Cell
          title='会员中心'
          extra='查看等级 →'
          clickable
          onClick={() => void Taro.navigateTo({ url: '/pages/user/member/index' })}
        />
      </View>

      <View className='user-points__rules'>
        说明：{'\n'}1. 下单成功并确认收货后获得积分{'\n'}
        2. 积分可在结算时抵扣订单金额{'\n'}
        3. 部分商品不参与积分抵扣
      </View>
    </View>
  )
}
