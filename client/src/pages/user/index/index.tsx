import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { getOrders } from '@/services/order'
import { getMyBalance } from '@/services/user'
import { useAuthStore } from '@/stores/auth'
import { Avatar, Button, Cell } from '@/ui/nutui'
import {
  ArrowRight,
  List,
  Location,
  Setting,
  User,
  HeartFill,
  Footprint,
  Order,
} from '@/ui/icons'
import {
  ToPay,
  ToReceive,
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
  queryKey: 'all' | 'pending_payment' | 'paid' | 'shipped'
}

const ORDER_SHORTCUTS: OrderShortcut[] = [
  { label: '待付款', tab: 'pending_payment', icon: ToPay, queryKey: 'pending_payment' },
  { label: '待发货', tab: 'paid', icon: Order, queryKey: 'paid' },
  { label: '待收货', tab: 'shipped', icon: ToReceive, queryKey: 'shipped' },
  { label: '全部订单', tab: '', icon: List, queryKey: 'all' },
]

interface ServiceShortcut {
  label: string
  description: string
  url: string
  icon: React.ComponentType<IconProps>
}

const SERVICE_SHORTCUTS: ServiceShortcut[] = [
  {
    label: '收货地址',
    description: '管理常用收货地址',
    url: '/pages/address/list/index',
    icon: Location,
  },
  {
    label: '我的收藏',
    description: '同步心仪商品清单',
    url: '/pages/user/favorite/index',
    icon: HeartFill,
  },
  {
    label: '浏览记录',
    description: '查看最近浏览商品',
    url: '/pages/user/history/index',
    icon: Footprint,
  },
  {
    label: '账号设置',
    description: '头像、昵称与密码管理',
    url: '/pages/user/settings/index',
    icon: Setting,
  },
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
  const { isLoggedIn, logout, user } = useAuthStore()

  useDidShow(() => {
    if (isLoggedIn) {
      void queryClient.invalidateQueries({ queryKey: ['user-center'] })
    }
  })

  const allOrdersQuery = useQuery({
    queryKey: ['user-center', 'orders', 'all'],
    queryFn: () => getOrders({ page: 1, page_size: 1 }),
    enabled: isLoggedIn,
    staleTime: 60 * 1000,
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

  const handleLogout = () => {
    void Taro.showModal({
      title: '退出登录',
      content: '确认退出当前账号？',
      confirmText: '退出',
      cancelText: '取消',
    }).then(({ confirm }) => {
      if (confirm) {
        void logout().then(() => {
          queryClient.clear()
          void Taro.switchTab({ url: '/pages/home/index' })
        })
      }
    })
  }

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

  const orderCounts = {
    all: allOrdersQuery.data?.total ?? 0,
    pending_payment: pendingOrdersQuery.data?.total ?? 0,
    paid: paidOrdersQuery.data?.total ?? 0,
    shipped: shippedOrdersQuery.data?.total ?? 0,
  }

  return (
    <View className='page-shell page-shell--tab user-page'>
      <View className='user-page__profile-card'>
        <View
          className='user-page__profile-setting'
          onClick={() => navigateProtected('/pages/user/settings/index')}
        >
          <Setting size={20} color='#7c5a31' />
        </View>

        <View
          className='user-page__profile-main'
          onClick={() => {
            if (!isLoggedIn) {
              handleLogin()
            }
          }}
        >
          <Avatar
            size='large'
            src={isLoggedIn ? (user?.avatar ?? undefined) : undefined}
            className='user-page__profile-avatar'
          >
            {isLoggedIn
              ? (user?.nickname?.slice(0, 1) ?? '我')
              : <User size={26} color='#8f6a41' />}
          </Avatar>

          <View className='user-page__profile-copy'>
            {isLoggedIn ? (
              <>
                <Text className='user-page__profile-name'>{user?.nickname ?? '微信用户'}</Text>
                <Text className='user-page__profile-phone'>{maskPhone(user?.phone)}</Text>
              </>
            ) : (
              <>
                <Text className='user-page__profile-name'>登录后查看完整账户</Text>
                <Text className='user-page__profile-phone'>订单、收藏和收货地址会同步到这里</Text>
              </>
            )}
          </View>

          {!isLoggedIn && (
            <View className='user-page__profile-arrow'>
              <ArrowRight width={16} height={16} color='#b17b45' />
            </View>
          )}
        </View>

        {isLoggedIn && (
          <View
            className='user-page__profile-balance'
            onClick={() => navigateProtected('/pages/user/balance/index')}
          >
            <Text className='user-page__profile-balance-label'>可用余额</Text>
            <Text className='user-page__profile-balance-value'>
              ¥{((balanceQuery.data?.balance_cents ?? 0) / 100).toFixed(2)}
            </Text>
            <ArrowRight width={14} height={14} color='#b17b45' />
          </View>
        )}

        {!isLoggedIn && (
          <Button type='primary' className='user-page__login-button' onClick={handleLogin}>
            立即登录
          </Button>
        )}
      </View>

      <View className='user-page__section'>
        <View className='user-page__section-head'>
          <View>
            <Text className='user-page__section-title'>订单中心</Text>
            <Text className='user-page__section-subtitle'>按状态快速查看处理中的订单</Text>
          </View>
          <Text
            className='user-page__section-link'
            onClick={() => openOrder('')}
          >
            查看全部
          </Text>
        </View>

        <View className='user-page__order-grid'>
          {ORDER_SHORTCUTS.map((item) => {
            const IconComp = item.icon
            const count = orderCounts[item.queryKey]
            const shouldShowBadge = isLoggedIn && count > 0

            return (
              <View
                key={item.label}
                className='user-page__order-item'
                onClick={() => openOrder(item.tab)}
              >
                <View className='user-page__order-icon-wrap'>
                  <IconComp size={22} color='#9a6936' />
                  {shouldShowBadge && (
                    <Text className='user-page__order-badge'>{formatCount(count)}</Text>
                  )}
                </View>
                <Text className='user-page__order-label'>{item.label}</Text>
              </View>
            )
          })}
        </View>
      </View>

      <View className='user-page__section user-page__section--tools'>
        <View className='user-page__section-head'>
          <View>
            <Text className='user-page__section-title'>常用功能</Text>
            <Text className='user-page__section-subtitle'>保留当前系统里真实可用的高频入口</Text>
          </View>
        </View>

        <View className='user-page__tool-grid'>
          {SERVICE_SHORTCUTS.map((item) => {
            const IconComp = item.icon

            return (
              <View
                key={item.label}
                className='user-page__tool-card'
                onClick={() => navigateProtected(item.url)}
              >
                <View className='user-page__tool-icon'>
                  <IconComp size={20} color='#7d5a2d' />
                </View>
                <View className='user-page__tool-copy'>
                  <Text className='user-page__tool-label'>{item.label}</Text>
                  <Text className='user-page__tool-description'>{item.description}</Text>
                </View>
                <ArrowRight width={14} height={14} color='#b89a74' />
              </View>
            )
          })}
        </View>
      </View>

      {isLoggedIn && (
        <Button
          plain
          type='primary'
          className='user-page__logout'
          onClick={handleLogout}
        >
          退出登录
        </Button>
      )}
    </View>
  )
}
