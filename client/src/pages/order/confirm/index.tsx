import { Image, Text, View } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useEffect, useMemo, useState } from 'react'
import PriceText from '@/components/PriceText'
import AddressCard from '@/components/AddressCard'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { usePay } from '@/hooks/usePay'
import { useAuthStore } from '@/stores/auth'
import { createOrder } from '@/services/order'
import { getCart, precheckCartItems } from '@/services/cart'
import { getAddresses } from '@/services/address'
import { getMyBalance } from '@/services/user'
import type { Address, CartItem } from '@/types/biz'
import { Button, Cell, CellGroup, Input, SafeArea, Skeleton, Switch } from '@/ui/nutui'
import {
  formatCartConflictReason,
  formatCartSkuAttrs,
  formatCartUnavailableReason,
} from '@/utils/cart'
import './index.scss'

// ─── Entry mode types ────────────────────────────────────────────────────────

interface DirectSkuItem {
  sku_id: string
  qty: number
  title?: string
  image?: string
  price_cents?: number
}

type CheckoutMode =
  | { type: 'direct'; item: DirectSkuItem }
  | { type: 'cart'; ids: string[] }
  | { type: 'empty' }

/** Reads and clears checkout storage keys. Called once on mount. */
function readCheckoutMode(): CheckoutMode {
  const direct = Taro.getStorageSync('checkout-direct-sku') as DirectSkuItem | null
  if (direct && direct.sku_id) {
    Taro.removeStorageSync('checkout-direct-sku')
    Taro.removeStorageSync('checkout-cart-item-ids')
    return { type: 'direct', item: direct }
  }

  const cartIds = Taro.getStorageSync('checkout-cart-item-ids') as string[] | string | null
  if (cartIds) {
    Taro.removeStorageSync('checkout-cart-item-ids')
    Taro.removeStorageSync('checkout-direct-sku')
    const ids: string[] = Array.isArray(cartIds)
      ? cartIds
      : (JSON.parse(cartIds as string) as string[])
    if (ids.length > 0) return { type: 'cart', ids }
  }

  return { type: 'empty' }
}

// ─── Normalised display item ─────────────────────────────────────────────────

interface DisplayItem {
  /** Unique key for React rendering */
  key: string
  cart_item_id?: string
  sku_id: string
  qty: number
  title: string
  image?: string
  spec?: string
  price_cents: number
  is_available?: boolean
  unavailable_reason?: string
  available_stock?: number
}

function fromDirect(item: DirectSkuItem): DisplayItem {
  return {
    key: String(item.sku_id),
    sku_id: item.sku_id,
    qty: item.qty,
    title: item.title ?? `SKU #${item.sku_id}`,
    image: item.image,
    price_cents: item.price_cents ?? 0,
  }
}

function fromCartItem(item: CartItem): DisplayItem {
  return {
    key: String(item.id),
    cart_item_id: String(item.id),
    sku_id: item.sku_id,
    qty: item.qty,
    title: item.product?.title || item.product_title,
    image: item.product?.main_image || item.sku_image,
    spec: formatCartSkuAttrs(item.sku_attrs) || undefined,
    price_cents: item.current_price_cents,
    is_available: item.is_available,
    unavailable_reason: item.unavailable_reason,
    available_stock: item.available_stock,
  }
}

// ─── Page ────────────────────────────────────────────────────────────────────

export default function OrderConfirmPage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const ensureAuth = useAuthGuard()

  const [buyerRemark, setBuyerRemark] = useState('')
  const [selectedAddress, setSelectedAddress] = useState<Address | null>(null)
  const [useBalance, setUseBalance] = useState(false)

  // Read checkout mode once on mount (useState initialiser runs once)
  const [checkoutMode] = useState<CheckoutMode>(() => readCheckoutMode())

  // ── Address query ──────────────────────────────────────────────────────────

  const addressQuery = useQuery({
    queryKey: ['addresses'],
    queryFn: getAddresses,
    enabled: isLoggedIn,
  })

  // ── Balance query ──────────────────────────────────────────────────────────

  const balanceQuery = useQuery({
    queryKey: ['balance'],
    queryFn: () => getMyBalance({ page: 1, page_size: 1 }),
    enabled: isLoggedIn,
  })

  const balanceCents = balanceQuery.data?.balance_cents ?? 0

  // Auto-select default or first address
  useEffect(() => {
    if (!selectedAddress && addressQuery.data) {
      const defaultAddr = addressQuery.data.find((a) => a.is_default) ?? addressQuery.data[0]
      if (defaultAddr) setSelectedAddress(defaultAddr)
    }
  }, [addressQuery.data, selectedAddress])

  // Restore address selected on the address-list page
  useDidShow(() => {
    const saved = Taro.getStorageSync('selected-address') as Address | null
    if (saved && typeof saved === 'object' && saved.id) {
      setSelectedAddress(saved)
      Taro.removeStorageSync('selected-address')
    }
  })

  // ── Auth guard ─────────────────────────────────────────────────────────────

  useEffect(() => {
    if (!isLoggedIn) {
      void ensureAuth(undefined, '/pages/order/confirm/index')
    }
  }, [ensureAuth, isLoggedIn])

  // ── Cart query (only in cart mode) ────────────────────────────────────────

  const cartQuery = useQuery({
    queryKey: ['cart'],
    queryFn: getCart,
    enabled: isLoggedIn && checkoutMode.type === 'cart',
  })

  // ── Derive display items ───────────────────────────────────────────────────

  const displayItems = useMemo<DisplayItem[]>(() => {
    if (checkoutMode.type === 'direct') {
      return [fromDirect(checkoutMode.item)]
    }
    if (checkoutMode.type === 'cart') {
      const cartIds = checkoutMode.ids
      const allItems = cartQuery.data?.items ?? []
      return allItems
        .filter((item) => cartIds.includes(String(item.id)))
        .map(fromCartItem)
    }
    return []
  }, [checkoutMode, cartQuery.data])

  const unavailableItems = useMemo(
    () =>
      checkoutMode.type === 'cart'
        ? displayItems.filter((item) => item.is_available === false)
        : [],
    [checkoutMode.type, displayItems],
  )
  const missingCartItemCount = useMemo(() => {
    if (checkoutMode.type !== 'cart') return 0
    return Math.max(checkoutMode.ids.length - displayItems.length, 0)
  }, [checkoutMode, displayItems.length])
  const payableItems = useMemo(
    () =>
      checkoutMode.type === 'cart'
        ? displayItems.filter((item) => item.is_available !== false)
        : displayItems,
    [checkoutMode.type, displayItems],
  )

  // ── Price calculation ──────────────────────────────────────────────────────

  const itemsTotalCents = useMemo(
    () => payableItems.reduce((sum, item) => sum + item.price_cents * item.qty, 0),
    [payableItems],
  )

  const balanceDeductCents = useMemo(
    () => (useBalance ? Math.min(balanceCents, itemsTotalCents) : 0),
    [useBalance, balanceCents, itemsTotalCents],
  )

  const payableCents = itemsTotalCents - balanceDeductCents

  // ── Pay hook ───────────────────────────────────────────────────────────────

  const { pay, loading: payLoading } = usePay({
    onSuccess: (orderId) => {
      void Taro.redirectTo({ url: `/pages/order/detail/index?id=${orderId}` })
    },
  })

  // ── Create order mutation ──────────────────────────────────────────────────

  const orderMutation = useMutation({
    mutationFn: async () => {
      if (!selectedAddress) throw new Error('请选择收货地址')
      if (displayItems.length === 0) throw new Error('请先选择商品')
      if (missingCartItemCount > 0) throw new Error('部分商品已不存在，请返回购物车重新选择')
      if (unavailableItems.length > 0) throw new Error('部分商品已失效，请返回购物车处理')

      if (checkoutMode.type === 'cart') {
        const precheck = await precheckCartItems(checkoutMode.ids)
        const conflicts = precheck.conflicts ?? []
        const hasBlockingConflict = conflicts.some((conflict) => conflict.reason !== 'price_changed')
        if (hasBlockingConflict) {
          const conflict = conflicts.find((item) => item.reason !== 'price_changed')
          throw new Error(
            conflict ? formatCartConflictReason(conflict.reason) : '部分商品库存或状态已变化，请返回购物车确认',
          )
        }
      }

      const key = `${Date.now()}-${Math.random().toString(36).slice(2)}`
      return createOrder({
        address_id: selectedAddress.id,
        items: payableItems.map((i) => ({ sku_id: i.sku_id, qty: i.qty })),
        buyer_remark: buyerRemark || undefined,
        idempotency_key: key,
        use_balance: useBalance,
      })
    },
    onSuccess: async (result) => {
      await pay(result.order_id)
    },
    onError: (err: Error) => {
      void Taro.showToast({ title: err.message || '下单失败', icon: 'none' })
    },
  })

  // ── Loading state (cart mode waits for cart data) ──────────────────────────

  const isCartLoading = checkoutMode.type === 'cart' && cartQuery.isLoading

  // ── Early returns ──────────────────────────────────────────────────────────

  if (!isLoggedIn) {
    return <View className='page-shell order-confirm-page' />
  }

  if (checkoutMode.type === 'empty') {
    return (
      <View className='page-shell order-confirm-page'>
        <View className='order-confirm-page__empty'>
          <Text>没有可下单的商品</Text>
          <Button type='primary' onClick={() => void Taro.switchTab({ url: '/pages/home/index' })}>
            去选购
          </Button>
        </View>
      </View>
    )
  }

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <View className='page-shell order-confirm-page'>
      {/* Address */}
      <View className='order-confirm-page__section'>
        {addressQuery.isLoading ? (
          <Skeleton animated rows={2} />
        ) : selectedAddress ? (
          <AddressCard
            address={selectedAddress}
            onClick={() => void Taro.navigateTo({ url: '/pages/address/list/index?mode=select' })}
          />
        ) : (
          <Cell
            title='添加收货地址'
            extra='去添加'
            clickable
            onClick={() => void Taro.navigateTo({ url: '/pages/address/edit/index' })}
          />
        )}
      </View>

      {/* Items */}
      <CellGroup className='order-confirm-page__section' title='商品清单'>
        {isCartLoading ? (
          <Skeleton animated rows={3} />
        ) : (
          <>
            {missingCartItemCount > 0 ? (
              <View className='order-confirm-page__alert'>
                <Text className='order-confirm-page__alert-title'>部分商品已从购物车移除</Text>
                <Text className='order-confirm-page__alert-text'>请返回购物车重新选择后再提交订单。</Text>
              </View>
            ) : null}
            {unavailableItems.length > 0 ? (
              <View className='order-confirm-page__alert'>
                <Text className='order-confirm-page__alert-title'>
                  {unavailableItems.length} 件商品已失效，当前无法提交订单
                </Text>
                <Text className='order-confirm-page__alert-text'>请返回购物车调整数量或清理失效商品。</Text>
              </View>
            ) : null}
            {displayItems.map((item) => (
              <View key={item.key} className='order-confirm-page__item'>
                {item.image ? (
                  <Image
                    className='order-confirm-page__item-image'
                    src={item.image}
                    mode='aspectFill'
                  />
                ) : null}
                <View className='order-confirm-page__item-main'>
                  <Text className='order-confirm-page__item-title'>{item.title}</Text>
                  {item.spec ? (
                    <Text className='order-confirm-page__item-spec'>{item.spec}</Text>
                  ) : null}
                  {item.is_available === false ? (
                    <Text className='order-confirm-page__item-status'>
                      {formatCartUnavailableReason(item.unavailable_reason)}
                    </Text>
                  ) : item.available_stock && item.available_stock <= 5 ? (
                    <Text className='order-confirm-page__item-status'>
                      库存仅剩 {item.available_stock} 件
                    </Text>
                  ) : null}
                  <View className='order-confirm-page__item-bottom'>
                    {item.price_cents > 0 ? <PriceText cents={item.price_cents} /> : null}
                    <Text className='order-confirm-page__item-qty'>x{item.qty}</Text>
                  </View>
                </View>
              </View>
            ))}
          </>
        )}
      </CellGroup>

      {/* Buyer remark */}
      <View className='order-confirm-page__section order-confirm-page__message'>
        <Input
          value={buyerRemark}
          placeholder='如需备注请输入留言（选填）'
          onChange={(value) => setBuyerRemark(String(value))}
        />
      </View>

      {/* Balance toggle */}
      {isLoggedIn && balanceCents > 0 && (
        <View className='order-confirm-page__section'>
          <Cell
            title={`使用余额（可用 ¥${(balanceCents / 100).toFixed(2)}）`}
            extra={
              <Switch
                checked={useBalance}
                onChange={(val: boolean) => setUseBalance(val)}
              />
            }
          />
        </View>
      )}

      {/* Price summary */}
      <View className='order-confirm-page__section order-confirm-page__summary'>
        <View className='order-confirm-page__summary-row'>
          <Text className='order-confirm-page__summary-label'>商品合计</Text>
          <Text className='order-confirm-page__summary-value'>
            ¥{(itemsTotalCents / 100).toFixed(2)}
          </Text>
        </View>
        <View className='order-confirm-page__summary-row'>
          <Text className='order-confirm-page__summary-label'>运费</Text>
          <Text className='order-confirm-page__summary-value'>¥0.00</Text>
        </View>
        {useBalance && balanceDeductCents > 0 && (
          <View className='order-confirm-page__summary-row'>
            <Text className='order-confirm-page__summary-label'>余额抵扣</Text>
            <Text className='order-confirm-page__summary-value' style={{ color: '#52c41a' }}>
              -¥{(balanceDeductCents / 100).toFixed(2)}
            </Text>
          </View>
        )}
        <View className='order-confirm-page__summary-row'>
          <Text className='order-confirm-page__summary-label'>实付</Text>
          <Text className='order-confirm-page__summary-value'>
            ¥{(payableCents / 100).toFixed(2)}
          </Text>
        </View>
      </View>

      {/* Fixed footer */}
      <View className='order-confirm-page__fixed-bar'>
        <View className='order-confirm-page__fixed-main'>
          <View className='order-confirm-page__fixed-total'>
            <Text className='order-confirm-page__fixed-label'>实付</Text>
            <Text className='order-confirm-page__fixed-amount'>
              ¥{(payableCents / 100).toFixed(2)}
            </Text>
          </View>
          <Button
            className='order-confirm-page__pay-button'
            type='primary'
            disabled={
              !selectedAddress ||
              payableItems.length === 0 ||
              isCartLoading ||
              unavailableItems.length > 0 ||
              missingCartItemCount > 0
            }
            loading={orderMutation.isPending || payLoading}
            onClick={() => orderMutation.mutate()}
          >
            提交订单
          </Button>
        </View>
        <SafeArea position='bottom' />
      </View>
    </View>
  )
}
