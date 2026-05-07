import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import EmptyState from '@/components/EmptyState'
import { getAddresses, deleteAddress, setDefaultAddress } from '@/services/address'
import type { Address } from '@/types/biz'
import { ArrowRight, Del, Location } from '@/ui/icons'
import { SafeArea, Skeleton } from '@/ui/nutui'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { useAuthStore } from '@/stores/auth'
import { ADDRESS_EDIT_DRAFT_KEY, formatAddress } from '@/utils/address'
import './index.scss'

function formatPhone(phone: string) {
  return phone.replace(/(\d{3})(\d{4})(\d{4})/, '$1 $2 $3')
}

function stopPropagation(event: { stopPropagation?: () => void }) {
  event.stopPropagation?.()
}

function AddressItem({
  address,
  isSelectMode,
  isDeleting,
  isSettingDefault,
  onSelect,
  onEdit,
  onDelete,
  onSetDefault,
}: {
  address: Address
  isSelectMode: boolean
  isDeleting: boolean
  isSettingDefault: boolean
  onSelect?: (address: Address) => void
  onEdit: (address: Address) => void
  onDelete: (address: Address) => void
  onSetDefault: (address: Address) => void
}) {
  const handlePrimaryClick = () => {
    if (isSelectMode) {
      onSelect?.(address)
      return
    }

    onEdit(address)
  }

  const defaultLabel = isSettingDefault
    ? '设置中...'
    : '设为默认'

  return (
    <View className={`address-list-page__item${address.is_default ? ' address-list-page__item--default' : ''}`}>
      <View className='address-list-page__card-head' onClick={handlePrimaryClick}>
        <View className='address-list-page__marker'>
          <Location size='32rpx' color='#c4883d' />
        </View>

        <View className='address-list-page__card-main'>
          <View className='address-list-page__identity'>
            <Text className='address-list-page__name'>{address.name}</Text>
            <Text className='address-list-page__phone'>{formatPhone(address.phone)}</Text>
            {address.is_default && (
              <Text className='address-list-page__badge'>默认地址</Text>
            )}
          </View>

          <View className='address-list-page__detail-row'>
            <View className='address-list-page__detail-pin' />
            <Text className='address-list-page__detail'>{formatAddress(address)}</Text>
          </View>
        </View>

        <View className='address-list-page__arrow'>
          <ArrowRight size='28rpx' color='#d0c4b6' />
        </View>
      </View>

      {!isSelectMode && (
        <View className='address-list-page__toolbar'>
          <View
            className={`address-list-page__action${address.is_default || isSettingDefault ? ' address-list-page__action--active' : ''}`}
            onClick={(event) => {
              stopPropagation(event)
              if (address.is_default || isSettingDefault) return
              onSetDefault(address)
            }}
          >
            <View className={`address-list-page__radio${address.is_default ? ' address-list-page__radio--checked' : ''}`}>
              {address.is_default && <View className='address-list-page__radio-dot' />}
            </View>
            <Text className='address-list-page__action-label'>{defaultLabel}</Text>
          </View>

          <View className='address-list-page__divider' />

          <View
            className='address-list-page__action'
            onClick={(event) => {
              stopPropagation(event)
              onEdit(address)
            }}
          >
            <View className='address-list-page__action-icon address-list-page__action-icon--edit' />
            <Text className='address-list-page__action-label'>编辑</Text>
          </View>

          <View className='address-list-page__divider' />

          <View
            className={`address-list-page__action${isDeleting ? ' address-list-page__action--danger' : ''}`}
            onClick={(event) => {
              stopPropagation(event)
              if (isDeleting) return
              onDelete(address)
            }}
          >
            <Del size='28rpx' color={isDeleting ? '#d07d67' : '#b9b1a6'} />
            <Text className='address-list-page__action-label'>
              {isDeleting ? '删除中...' : '删除'}
            </Text>
          </View>
        </View>
      )}
    </View>
  )
}

export default function AddressListPage() {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn)
  const ensureAuth = useAuthGuard()
  const queryClient = useQueryClient()
  const params = Taro.getCurrentInstance().router?.params
  const isSelectMode = params?.mode === 'select'

  useEffect(() => {
    if (!isLoggedIn) void ensureAuth(undefined, '/pages/address/list/index')
  }, [ensureAuth, isLoggedIn])

  const { data: addresses, isLoading, refetch } = useQuery({
    queryKey: ['addresses'],
    queryFn: getAddresses,
    enabled: isLoggedIn,
  })

  useDidShow(() => {
    void refetch()
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteAddress(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['addresses'] })
      void Taro.showToast({ title: '删除成功', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '删除失败，请重试', icon: 'none' }),
  })

  const defaultMutation = useMutation({
    mutationFn: (id: string) => setDefaultAddress(id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['addresses'] })
      void Taro.showToast({ title: '默认地址已更新', icon: 'success' })
    },
    onError: () => void Taro.showToast({ title: '设置默认地址失败', icon: 'none' }),
  })

  function handleSelect(address: Address) {
    Taro.setStorageSync('selected-address', address)
    Taro.navigateBack()
  }

  function handleEdit(address: Address) {
    Taro.setStorageSync(ADDRESS_EDIT_DRAFT_KEY, address)
    void Taro.navigateTo({ url: `/pages/address/edit/index?id=${address.id}` })
  }

  function handleCreate() {
    Taro.removeStorageSync(ADDRESS_EDIT_DRAFT_KEY)
    void Taro.navigateTo({ url: '/pages/address/edit/index' })
  }

  function handleDelete(address: Address) {
    void Taro.showModal({
      title: '删除地址',
      content: `确认删除 ${address.name} 的收货地址吗？`,
      confirmText: '删除',
      confirmColor: '#d07d67',
    }).then(({ confirm }) => {
      if (confirm) {
        deleteMutation.mutate(address.id)
      }
    })
  }

  function handleSetDefault(address: Address) {
    defaultMutation.mutate(address.id)
  }

  if (!isLoggedIn) return <View className='page-shell' />

  return (
    <View className='page-shell address-list-page'>
      <View className='address-list-page__glow address-list-page__glow--top' />
      <View className='address-list-page__glow address-list-page__glow--bottom' />

      {isLoading ? (
        <View className='address-list-page__skeletons'>
          {Array.from({ length: 4 }).map((_, index) => (
            <View key={index} className='address-list-page__skeleton-card'>
              <Skeleton animated rows={3} />
            </View>
          ))}
        </View>
      ) : addresses && addresses.length > 0 ? (
        <View className='address-list-page__list'>
          {addresses.map((address) => (
            <AddressItem
              key={address.id}
              address={address}
              isSelectMode={isSelectMode}
              isDeleting={deleteMutation.isPending && deleteMutation.variables === address.id}
              isSettingDefault={defaultMutation.isPending && defaultMutation.variables === address.id}
              onSelect={handleSelect}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onSetDefault={handleSetDefault}
            />
          ))}
        </View>
      ) : (
        <View className='address-list-page__empty'>
          <EmptyState title='暂无收货地址' description='添加一个常用地址，下单时可以更快完成选择。' />
        </View>
      )}

      <View className='address-list-page__fixed-bar'>
        <View className='address-list-page__add-button' onClick={handleCreate}>
          <Text className='address-list-page__add-icon'>+</Text>
          <Text className='address-list-page__add-label'>新增收货地址</Text>
        </View>
        <SafeArea position='bottom' />
      </View>
    </View>
  )
}
