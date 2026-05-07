import { PropsWithChildren } from 'react'
import Taro from '@tarojs/taro'
import { View } from '@tarojs/components'
import { ArrowLeft, Cart } from '@/ui/icons'
import { Badge, NavBar } from '@/ui/nutui'
import { useCartStore } from '@/stores/cart'
import { isH5 } from '@/utils/platform'
import './index.scss'

const pageMeta: Record<string, { title: string; isTab?: boolean; rightAction?: 'cart'; hideShellNav?: boolean }> = {
  'pages/home/index': { title: '首页', isTab: true },
  'pages/category/index': { title: '分类', isTab: true },
  'pages/cart/index': { title: '购物车', isTab: true },
  'pages/user/index/index': { title: '我的', isTab: true },
  'pages/product/detail/index': { title: '商品详情', hideShellNav: true },
  'pages/product/list/index': { title: '商品列表' },
  'pages/order/confirm/index': { title: '填写订单' },
  'pages/user/favorite/index': { title: '我的收藏' },
  'pages/user/history/index': { title: '浏览历史' },
  'pages/user/settings/index': { title: '账号设置' },
  'pages/auth/h5-callback/index': { title: '微信登录' },
  'pages/auth/login/index': { title: '登录' },
  'pages/auth/register/index': { title: '注册' },
  'pages/auth/reset-password/index': { title: '重置密码' },
  'pages/order/list/index': { title: '我的订单' },
  'pages/order/detail/index': { title: '订单详情' },
  'pages/address/list/index': { title: '收货地址' },
  'pages/address/edit/index': { title: '编辑地址' },
}

interface H5ShellProps extends PropsWithChildren {
  currentRoute: string
  stackLength: number
}

export default function H5Shell({ currentRoute, stackLength, children }: H5ShellProps) {
  const count = useCartStore((state) => state.count)

  if (!isH5) {
    return <>{children}</>
  }

  const meta = pageMeta[currentRoute] ?? { title: '徐记小铺' }

  // Tab pages: no NavBar, Taro native tabBar handles navigation
  if (meta.isTab || meta.hideShellNav) {
    return <>{children}</>
  }

  const rightAction =
    meta.rightAction === 'cart' ? (
      <View
        className='h5-shell__nav-action'
        onClick={() => void Taro.switchTab({ url: '/pages/cart/index' })}
      >
        <Badge value={count > 0 ? count : ''}>
          <Cart size='20' />
        </Badge>
      </View>
    ) : null

  return (
    <View className='h5-shell'>
      <NavBar
        fixed
        safeAreaInsetTop
        placeholder
        title={meta.title}
        back={<ArrowLeft size='18' />}
        right={rightAction}
        onBackClick={() => {
          if (stackLength > 1) {
            void Taro.navigateBack()
            return
          }
          void Taro.switchTab({ url: '/pages/home/index' })
        }}
      />
      <View className='h5-shell__body h5-shell__body--nav'>{children}</View>
    </View>
  )
}
