import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { getOrders } from '@/services/order'
import { getMyBalance, getFavorites, getHistory } from '@/services/user'
import { useAuthStore } from '@/stores/auth'
import { Avatar } from '@/ui/nutui'
import {
  ArrowRight,
  List,
  Location,
  Order,
  Setting,
  User,
  HeartFill,
  Footprint,
} from '@/ui/icons'
import {
  ToReceive,
  Wallet,
} from '@nutui/icons-react-taro'
import './index.scss'

interface IconProps {
  size?: number | string
  color?: string
}

interface OrderShortcut {
  label: string
  tab: string
  icon: React.ComponentType<IconProps>
  queryKey: 'pending_payment' | 'paid' | 'shipped' | null
}

const ORDER_SHORTCUTS: OrderShortcut[] = [
  { label: '待付款', tab: 'pending_payment', icon: Wallet, queryKey: 'pending_payment' },
  { label: '待发货', tab: 'paid', icon: Order, queryKey: 'paid' },
  { label: '待收货', tab: 'shipped', icon: ToReceive, queryKey: 'shipped' },
  { label: '全部订单', tab: '', icon: List, queryKey: null },
]

interface ServiceItem {
  label: string
  url: string
  icon: React.ComponentType<IconProps>
}

const MY_SERVICES: ServiceItem[] = [
  { label: '收货地址', url: '/pages/address/list/index', icon: Location },
  { label: '我的收藏', url: '/pages/user/favorite/index', icon: HeartFill },
  { label: '浏览记录', url: '/pages/user/history/index', icon: Footprint },
]

const MORE_SERVICES: ServiceItem[] = [
  { label: '账号设置', url: '/pages/user/settings/index', icon: Setting },
]

function maskPhone(phone: string | null | undefined): string {
  if (!phone) {
    return '暂未绑定手机号'
  }

  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

function formatCount(value: number): string {
  if (value > 99) {
    return '99+'
  }

  return String(value)
}

export default function UserPage() {
  const queryClient = useQueryClient()
  const ensureAuth = useAuthGuard()
  const { isLoggedIn, user } = useAuthStore()

  useDidShow(() => {
    if (isLoggedIn) {
      void queryClient.invalidateQueries({ queryKey: ['user-center'] })
    }
  })

  const pendingOrdersQuery = useQuery({
    queryKey: ['user-center', 'orders', 'pending_payment'],
    queryFn: () => getOrders({ status: 'pending_payment', page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
  })

  const paidOrdersQuery = useQuery({
    queryKey: ['user-center', 'orders', 'paid'],
    queryFn: () => getOrders({ status: 'paid', page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
  })

  const shippedOrdersQuery = useQuery({
    queryKey: ['user-center', 'orders', 'shipped'],
    queryFn: () => getOrders({ status: 'shipped', page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
  })

  const balanceQuery = useQuery({
    queryKey: ['user-center', 'balance'],
    queryFn: () => getMyBalance({ page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
  })

  const favoritesQuery = useQuery({
    queryKey: ['user-center', 'favorites'],
    queryFn: () => getFavorites({ page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
  })

  const historyQuery = useQuery({
    queryKey: ['user-center', 'history'],
    queryFn: () => getHistory({ page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
  })

  const handleLogin = () => {
    void ensureAuth(undefined, '/pages/user/index/index')
  }

  const navigateProtected = (url: string) => {
    void ensureAuth(() => {
      void Taro.navigateTo({ url })
    }, url)
  }

  const openOrder = (tab: string) => {
    const url = tab ? `/pages/order/list/index?tab=${tab}` : '/pages/order/list/index'

    void ensureAuth(() => {
      void Taro.navigateTo({ url })
    }, url)
  }

  const orderCounts: Record<string, number> = {
    pending_payment: pendingOrdersQuery.data?.total ?? 0,
    paid: paidOrdersQuery.data?.total ?? 0,
    shipped: shippedOrdersQuery.data?.total ?? 0,
  }

  const stats = [
    { label: '收藏', value: isLoggedIn ? String(favoritesQuery.data?.total ?? 0) : '0', url: '/pages/user/favorite/index' },
    { label: '足迹', value: isLoggedIn ? String(historyQuery.data?.total ?? 0) : '0', url: '/pages/user/history/index' },
    { label: '余额', value: isLoggedIn ? `¥${((balanceQuery.data?.balance_cents ?? 0) / 100).toFixed(2)}` : '¥0.00', url: '/pages/user/balance/index' },
    { label: '优惠券', value: '0', url: '' },
  ]

  return (
    <View className='user-page'>
      {/* Green header banner */}
      <View className='user-page__banner'>
        <View
          className='user-page__setting-btn'
          onClick={() => navigateProtected('/pages/user/settings/index')}
        >
          <Setting size={22} color='#fff' />
        </View>

        <View
          className='user-page__profile-row'
          onClick={() => { if (!isLoggedIn) { handleLogin() } }}
        >
          <Avatar
            size='large'
            src={isLoggedIn ? (user?.avatar ?? undefined) : undefined}
            className='user-page__avatar'
          >
            {isLoggedIn
              ? (user?.nickname?.slice(0, 1) ?? '我')
              : <User size={26} color='#07C160' />}
          </Avatar>

          <View className='user-page__profile-info'>
            {isLoggedIn ? (
              <>
                <Text className='user-page__nickname'>{user?.nickname ?? '微信用户'}</Text>
                <Text className='user-page__phone'>{maskPhone(user?.phone)}</Text>
              </>
            ) : (
              <>
                <Text className='user-page__nickname'>登录后查看完整账户</Text>
                <Text className='user-page__phone'>点击立即登录</Text>
              </>
            )}
          </View>
        </View>

        {/* Stats row */}
        <View className='user-page__stats'>
          {stats.map((s) => (
            <View
              key={s.label}
              className='user-page__stat-item'
              onClick={() => { if (s.url) { navigateProtected(s.url) } }}
            >
              <Text className='user-page__stat-count'>{s.value}</Text>
              <Text className='user-page__stat-label'>{s.label}</Text>
            </View>
          ))}
        </View>

        {/* Balance row */}
        {isLoggedIn && (
          <View
            className='user-page__balance-row'
            onClick={() => navigateProtected('/pages/user/balance/index')}
          >
            <Text className='user-page__balance-label'>可用余额</Text>
            <Text className='user-page__balance-value'>
              ¥{((balanceQuery.data?.balance_cents ?? 0) / 100).toFixed(2)}
            </Text>
            <ArrowRight width={14} height={14} color='rgba(255,255,255,0.7)' />
          </View>
        )}
      </View>

      {/* Orders section */}
      <View className='user-page__section'>
        <View className='user-page__section-head'>
          <Text className='user-page__section-title'>我的订单</Text>
          <Text className='user-page__section-link' onClick={() => openOrder('')}>
            查看全部 &gt;
          </Text>
        </View>

        <View className='user-page__icon-grid'>
          {ORDER_SHORTCUTS.map((item) => {
            const IconComp = item.icon
            const count = item.queryKey ? (orderCounts[item.queryKey] ?? 0) : 0
            const showBadge = isLoggedIn && count > 0

            return (
              <View
                key={item.label}
                className='user-page__icon-item'
                onClick={() => openOrder(item.tab)}
              >
                <View className='user-page__icon-wrap'>
                  <IconComp size={28} color='#333' />
                  {showBadge && (
                    <Text className='user-page__badge'>{formatCount(count)}</Text>
                  )}
                </View>
                <Text className='user-page__icon-label'>{item.label}</Text>
              </View>
            )
          })}
        </View>
      </View>

      {/* My services section */}
      <View className='user-page__section'>
        <View className='user-page__section-head'>
          <Text className='user-page__section-title'>我的服务</Text>
        </View>

        <View className='user-page__icon-grid'>
          {MY_SERVICES.map((item) => {
            const IconComp = item.icon

            return (
              <View
                key={item.label}
                className='user-page__icon-item'
                onClick={() => navigateProtected(item.url)}
              >
                <View className='user-page__icon-wrap'>
                  <IconComp size={28} color='#333' />
                </View>
                <Text className='user-page__icon-label'>{item.label}</Text>
              </View>
            )
          })}
        </View>
      </View>

      {/* More services section */}
      <View className='user-page__section'>
        <View className='user-page__section-head'>
          <Text className='user-page__section-title'>更多服务</Text>
        </View>

        <View className='user-page__icon-grid'>
          {MORE_SERVICES.map((item) => {
            const IconComp = item.icon

            return (
              <View
                key={item.label}
                className='user-page__icon-item'
                onClick={() => navigateProtected(item.url)}
              >
                <View className='user-page__icon-wrap'>
                  <IconComp size={28} color='#333' />
                </View>
                <Text className='user-page__icon-label'>{item.label}</Text>
              </View>
            )
          })}
        </View>
      </View>
    </View>
  )
}
