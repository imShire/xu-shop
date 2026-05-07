import { View, Text } from '@tarojs/components'
import type { Address } from '@/types/biz'
import { Tag } from '@/ui/nutui'
import { formatAddress } from '@/utils/address'
import './index.scss'

interface AddressCardProps {
  address: Address
  onClick?: () => void
  className?: string
}

export default function AddressCard({ address, onClick, className = '' }: AddressCardProps) {
  return (
    <View className={`address-card ${className}`} onClick={onClick}>
      <View className='address-card__top'>
        <Text className='address-card__name'>{address.name}</Text>
        <Text className='address-card__phone'>{address.phone}</Text>
        {address.is_default && <Tag type='primary'>默认</Tag>}
      </View>
      <Text className='address-card__detail'>
        {formatAddress(address)}
      </Text>
    </View>
  )
}
