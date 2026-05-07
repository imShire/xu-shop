import { Image, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import CartListItem from '@/components/CartListItem'
import EmptyState from '@/components/EmptyState'
import {
  batchDeleteCartItems,
  cleanInvalidCartItems,
  deleteCartItem,
  getCart,
  precheckCartItems,
  updateCartItem,
} from '@/services/cart'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import type { CartItem } from '@/types/biz'
import { Button, Checkbox, SafeArea, Skeleton } from '@/ui/nutui'
import { formatCartConflictReason } from '@/utils/cart'
import { formatPrice } from '@/utils/price'
import './index.scss'

const EMPTY_CART_ITEMS: CartItem[] = []
const EMPTY_CART_DATA = { items: EMPTY_CART_ITEMS, total: 0 }

export default function CartPage() {
  const isLoggedIn = useAuthStore((state) => state.isLoggedIn)
  const refreshCount = useCartStore((state) => state.refreshCount)
  const queryClient = useQueryClient()

  const [editMode, setEditMode] = useState(false)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [qtyDrafts, setQtyDrafts] = useState<Record<string, number>>({})
  const [isFlushingQty, setIsFlushingQty] = useState(false)

  const selectionInitializedRef = useRef(false)
  const previousAvailableIdsRef = useRef<Set<string>>(new Set())
  const qtyTimersRef = useRef<Record<string, ReturnType<typeof setTimeout>>>({})

  const cartQuery = useQuery({
    queryKey: ['cart'],
    queryFn: getCart,
    enabled: isLoggedIn,
    retry: 0,
  })

  const cartData = cartQuery.data ?? EMPTY_CART_DATA
  const cartItems = useMemo(
    () => (isLoggedIn ? cartData.items : EMPTY_CART_ITEMS),
    [cartData.items, isLoggedIn],
  )
  const availableItems = useMemo(
    () => cartItems.filter((item) => item.is_available && item.available_stock !== 0),
    [cartItems],
  )
  const invalidItems = useMemo(
    () => cartItems.filter((item) => !item.is_available || item.available_stock === 0),
    [cartItems],
  )
  const itemById = useMemo(
    () => new Map(cartItems.map((item) => [item.id, item])),
    [cartItems],
  )
  const selectedItems = useMemo(
    () =>
      cartItems
        .filter((item) => item.is_available && selectedIds.has(item.id))
        .map((item) => ({ ...item, qty: qtyDrafts[item.id] ?? item.qty })),
    [cartItems, qtyDrafts, selectedIds],
  )

  useEffect(() => {
    if (!isLoggedIn) {
      selectionInitializedRef.current = false
      previousAvailableIdsRef.current = new Set()
      setSelectedIds(new Set())
      Object.values(qtyTimersRef.current).forEach((timer) => clearTimeout(timer))
      qtyTimersRef.current = {}
      setQtyDrafts({})
    }
  }, [isLoggedIn])

  useEffect(
    () => () => {
      Object.values(qtyTimersRef.current).forEach((timer) => clearTimeout(timer))
    },
    [],
  )

  useEffect(() => {
    const availableIdSet = new Set(availableItems.map((item) => item.id))

    setSelectedIds((prev) => {
      if (!selectionInitializedRef.current) {
        selectionInitializedRef.current = true
        previousAvailableIdsRef.current = availableIdSet
        return new Set(availableIdSet)
      }

      const next = new Set<string>()
      prev.forEach((id) => {
        if (availableIdSet.has(id)) next.add(id)
      })
      availableIdSet.forEach((id) => {
        if (!previousAvailableIdsRef.current.has(id)) {
          next.add(id)
        }
      })

      previousAvailableIdsRef.current = availableIdSet
      if (prev.size === next.size) {
        let unchanged = true
        prev.forEach((id) => {
          if (!next.has(id)) unchanged = false
        })
        if (unchanged) return prev
      }
      return next
    })
  }, [availableItems])

  useEffect(() => {
    setQtyDrafts((prev) => {
      const next: Record<string, number> = {}
      let changed = false

      Object.entries(prev).forEach(([id, qty]) => {
        const item = itemById.get(id)
        if (!item) {
          changed = true
          return
        }
        if (item.qty === qty) {
          changed = true
          return
        }
        next[id] = qty
      })

      return changed ? next : prev
    })
  }, [itemById])

  const allSelected =
    availableItems.length > 0 &&
    availableItems.every((item) => selectedIds.has(item.id))

  const totalCents = useMemo(
    () => selectedItems.reduce((sum, item) => sum + item.current_price_cents * item.qty, 0),
    [selectedItems],
  )

  const savedCents = useMemo(
    () =>
      selectedItems.reduce((sum, item) => {
        const diff = item.snapshot_price_cents - item.current_price_cents
        return diff > 0 ? sum + diff * item.qty : sum
      }, 0),
    [selectedItems],
  )

  const syncCartState = useCallback(async () => {
    await queryClient.invalidateQueries({ queryKey: ['cart'] })
    await refreshCount()
  }, [queryClient, refreshCount])

  const updateMutation = useMutation({
    mutationFn: ({ id, qty }: { id: string; qty: number }) => updateCartItem(id, qty),
  })

  const singleDeleteMutation = useMutation({
    mutationFn: (id: string) => deleteCartItem(id),
    onSuccess: (_, id) => {
      setSelectedIds((prev) => {
        const next = new Set(prev)
        next.delete(id)
        return next
      })
      void syncCartState()
      void Taro.showToast({ title: '删除成功', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '删除失败', icon: 'none' }),
  })

  const batchDeleteMutation = useMutation({
    mutationFn: (ids: string[]) => batchDeleteCartItems(ids),
    onSuccess: () => {
      setSelectedIds(new Set())
      void syncCartState()
      void Taro.showToast({ title: '删除成功', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '删除失败', icon: 'none' }),
  })

  const cleanInvalidMutation = useMutation({
    mutationFn: cleanInvalidCartItems,
    onSuccess: () => {
      void syncCartState()
      void Taro.showToast({ title: '失效商品已清理', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '清理失败', icon: 'none' }),
  })

  const precheckMutation = useMutation({
    mutationFn: (ids: string[]) => precheckCartItems(ids),
    onError: () => void Taro.showToast({ title: '结算校验失败', icon: 'none' }),
  })

  const itemActionBusy = updateMutation.isPending || singleDeleteMutation.isPending || isFlushingQty

  const commitQtyChange = useCallback(
    async (id: string, qty: number) => {
      const item = itemById.get(id)
      if (!item || item.qty === qty) return true

      try {
        await updateMutation.mutateAsync({ id, qty })
        await syncCartState()
        return true
      } catch (error) {
        setQtyDrafts((prev) => {
          const next = { ...prev }
          delete next[id]
          return next
        })
        await syncCartState()
        void Taro.showToast({
          title: error instanceof Error ? error.message : '更新失败',
          icon: 'none',
        })
        return false
      }
    },
    [itemById, syncCartState, updateMutation],
  )

  const flushPendingQtyChanges = useCallback(async () => {
    const pendingEntries = Object.entries(qtyDrafts).filter(([id, qty]) => itemById.get(id)?.qty !== qty)
    if (pendingEntries.length === 0) return true

    setIsFlushingQty(true)
    try {
      for (const [id, qty] of pendingEntries) {
        const timer = qtyTimersRef.current[id]
        if (timer) {
          clearTimeout(timer)
          delete qtyTimersRef.current[id]
        }
        const ok = await commitQtyChange(id, qty)
        if (!ok) return false
      }
      return true
    } finally {
      setIsFlushingQty(false)
    }
  }, [commitQtyChange, itemById, qtyDrafts])

  const handleQtyChange = useCallback(
    (id: string, qty: number) => {
      setQtyDrafts((prev) => ({ ...prev, [id]: qty }))

      const timer = qtyTimersRef.current[id]
      if (timer) clearTimeout(timer)
      qtyTimersRef.current[id] = setTimeout(() => {
        void commitQtyChange(id, qty).finally(() => {
          delete qtyTimersRef.current[id]
        })
      }, 300)
    },
    [commitQtyChange],
  )

  function toggleItem(id: string, checked: boolean) {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (checked) {
        next.add(id)
      } else {
        next.delete(id)
      }
      return next
    })
  }

  function toggleSelectAll(checked: boolean) {
    setSelectedIds(checked ? new Set(availableItems.map((item) => item.id)) : new Set())
  }

  function handleDelete() {
    if (selectedIds.size === 0) {
      void Taro.showToast({ title: '请先选择商品', icon: 'none' })
      return
    }

    void Taro.showModal({
      title: '删除商品',
      content: `确定删除已选中的 ${selectedIds.size} 件商品吗？`,
      confirmText: '删除',
      confirmColor: '#ff4d4f',
    }).then(({ confirm }) => {
      if (confirm) {
        batchDeleteMutation.mutate(Array.from(selectedIds))
      }
    })
  }

  function handleSingleDelete(id: string) {
    void Taro.showModal({
      title: '删除商品',
      content: '确定删除这件商品吗？',
      confirmText: '删除',
      confirmColor: '#ff4d4f',
    }).then(({ confirm }) => {
      if (confirm) {
        singleDeleteMutation.mutate(id)
      }
    })
  }

  function handleCleanInvalid() {
    if (invalidItems.length === 0) return

    void Taro.showModal({
      title: '清理失效商品',
      content: `确定清理 ${invalidItems.length} 件失效商品吗？`,
      confirmText: '一键清理',
      confirmColor: '#ff4d4f',
    }).then(({ confirm }) => {
      if (confirm) {
        cleanInvalidMutation.mutate()
      }
    })
  }

  async function handleCheckout() {
    if (!isLoggedIn) {
      void Taro.navigateTo({ url: '/pages/auth/login/index?redirect=%2Fpages%2Fcart%2Findex' })
      return
    }

    if (selectedItems.length === 0) {
      void Taro.showToast({ title: '请先选择商品', icon: 'none' })
      return
    }

    try {
      const synced = await flushPendingQtyChanges()
      if (!synced) return

      const latestCart = await queryClient.fetchQuery({
        queryKey: ['cart'],
        queryFn: getCart,
      })
      const latestItemById = new Map(latestCart.items.map((item) => [item.id, item]))
      const latestSelectedItems = latestCart.items.filter(
        (item) => item.is_available && selectedIds.has(item.id),
      )

      if (latestSelectedItems.length === 0) {
        void Taro.showToast({ title: '已选商品状态已变化，请重新选择', icon: 'none' })
        return
      }

      const selectedItemIds = latestSelectedItems.map((item) => item.id)
      const result = await precheckMutation.mutateAsync(selectedItemIds)
      const conflicts = result.conflicts ?? []
      const blockingConflicts = conflicts.filter((conflict) => conflict.reason !== 'price_changed')
      const priceChangedConflicts = conflicts.filter((conflict) => conflict.reason === 'price_changed')

      if (blockingConflicts.length > 0) {
        const lines = blockingConflicts.slice(0, 3).map((conflict) => {
          const title =
            latestItemById.get(conflict.cart_item_id)?.product?.title ||
            latestItemById.get(conflict.cart_item_id)?.product_title ||
            '商品'
          return `${title}：${formatCartConflictReason(conflict.reason)}`
        })
        const suffix = blockingConflicts.length > 3 ? '\n更多商品状态已变化，请刷新后重试。' : ''

        await syncCartState()
        await Taro.showModal({
          title: '部分商品暂不可结算',
          content: `${lines.join('\n')}${suffix}`,
          showCancel: false,
          confirmText: '我知道了',
        })
        return
      }

      if (priceChangedConflicts.length > 0) {
        await syncCartState()
        await Taro.showModal({
          title: '价格已同步更新',
          content: `${priceChangedConflicts.length} 件商品价格已更新，当前页面展示的是最新价格。确认后继续结算。`,
          showCancel: false,
          confirmText: '继续结算',
        })
      }

      Taro.setStorageSync('checkout-cart-item-ids', JSON.stringify(selectedItemIds))
      void Taro.navigateTo({ url: '/pages/order/confirm/index' })
    } catch {
      // Error toast is handled by the mutation.
    }
  }

  if (!isLoggedIn) {
    return (
      <View className='page-shell page-shell--tab cart-page'>
        <EmptyState
          title='请先登录'
          description='登录后查看购物车'
          actionText='去登录'
          onAction={() =>
            void Taro.navigateTo({
              url: '/pages/auth/login/index?redirect=%2Fpages%2Fcart%2Findex',
            })
          }
        />
      </View>
    )
  }

  return (
    <View className='page-shell page-shell--tab cart-page'>
      <View className='cart-page__count-bar'>
        <View>
          <Text className='cart-page__count-text'>共 {cartData.total} 件商品</Text>
          {invalidItems.length > 0 ? (
            <Text className='cart-page__count-subtext'>含 {invalidItems.length} 件失效商品</Text>
          ) : null}
        </View>
        <View className='cart-page__header-actions'>
          <Button
            className='cart-page__header-edit'
            size='small'
            fill='none'
            onClick={() => setEditMode((value) => !value)}
          >
            {editMode ? '完成' : '编辑'}
          </Button>
        </View>
      </View>

      <View className='cart-page__body'>
        {cartQuery.isLoading ? (
          <View className='cart-page__list'>
            {Array.from({ length: 3 }).map((_, index) => (
              <Skeleton key={index} animated rows={3} />
            ))}
          </View>
        ) : cartItems.length === 0 ? (
          <View className='cart-page__empty'>
            <Image
              className='cart-page__empty-img'
              src='https://img12.360buyimg.com/imagetools/jfs/t1/196430/38/8105/593606/61101f89Etd10e0ab/9a1df89c1ab14f1b.png'
              mode='aspectFit'
            />
            <Text className='cart-page__empty-text'>购物车为空哦~</Text>
          </View>
        ) : (
          <>
            <View className='cart-page__notice'>
              <View className='cart-page__notice-copy'>
                <Text className='cart-page__notice-title'>
                  {invalidItems.length > 0 ? `当前有 ${invalidItems.length} 件商品暂不可结算` : '已为你自动勾选可结算商品'}
                </Text>
                <Text className='cart-page__notice-text'>
                  {invalidItems.length > 0
                    ? '可一键清理失效商品，保留列表更清爽。'
                    : '数量会实时联动库存与结算金额，确认页也会沿用最新状态。'}
                </Text>
              </View>
              {invalidItems.length > 0 ? (
                <Button
                  className='cart-page__notice-action'
                  fill='none'
                  loading={cleanInvalidMutation.isPending}
                  onClick={handleCleanInvalid}
                >
                  立即清理
                </Button>
              ) : null}
            </View>

            <View className='cart-page__list'>
              {cartItems.map((item) => (
                <CartListItem
                  key={item.id}
                  item={item}
                  checked={selectedIds.has(item.id)}
                  editMode={editMode}
                  busy={itemActionBusy}
                  displayQty={qtyDrafts[item.id] ?? item.qty}
                  maxQty={item.available_stock}
                  onToggle={(checked) => toggleItem(item.id, checked)}
                  onQtyChange={(qty) => handleQtyChange(item.id, qty)}
                  onDelete={() => handleSingleDelete(item.id)}
                />
              ))}
            </View>

            <View className='cart-page__fixed-bar'>
              <View className='cart-page__fixed-panel'>
                <View className='cart-page__fixed-left'>
                  <Checkbox checked={allSelected} onChange={toggleSelectAll}>
                    全选
                  </Checkbox>
                  {!editMode ? (
                    <View className='cart-page__fixed-meta'>
                      <Text className='cart-page__fixed-total'>
                        合计 <Text className='cart-page__fixed-total-price'>¥{formatPrice(totalCents)}</Text>
                      </Text>
                      <Text className='cart-page__fixed-summary'>
                        {savedCents > 0
                          ? `已为你节省 ¥${formatPrice(savedCents)}`
                          : isFlushingQty
                            ? '正在同步最新数量...'
                            : '已选可结算商品'}
                      </Text>
                    </View>
                  ) : null}
                </View>

                <View className='cart-page__fixed-actions'>
                  {editMode ? (
                    <Button
                      type='danger'
                      disabled={selectedIds.size === 0 || itemActionBusy || cleanInvalidMutation.isPending}
                      loading={batchDeleteMutation.isPending}
                      onClick={handleDelete}
                    >
                      删除选中{selectedIds.size > 0 ? `(${selectedIds.size})` : ''}
                    </Button>
                  ) : (
                    <Button
                      type='primary'
                      disabled={selectedItems.length === 0 || itemActionBusy || cleanInvalidMutation.isPending}
                      loading={precheckMutation.isPending}
                      onClick={handleCheckout}
                    >
                      去结算{selectedItems.length > 0 ? `(${selectedItems.length})` : ''}
                    </Button>
                  )}
                </View>
              </View>
              <SafeArea position='bottom' />
            </View>
          </>
        )}
      </View>
    </View>
  )
}
