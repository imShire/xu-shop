import { Text, View } from '@tarojs/components'
import type { CSSProperties } from 'react'
import type { CouponTemplate, UserCoupon } from '@/types/biz'
import { Tag } from '@/ui/nutui'
import { formatPrice, formatYuan } from '@/utils/price'
import './index.scss'

interface BaseProps {
  /** 显示状态徽章（已领取/已使用/已过期）。默认 unused */
  status?: 'unused' | 'locked' | 'used' | 'expired' | 'claimed' | 'claimable'
  /** 主操作按钮内容 + 回调；不传则不渲染按钮 */
  action?: { text: string; onClick: () => void; disabled?: boolean }
  /** 整卡点击 */
  onClick?: () => void
  /** 是否禁用（灰色） */
  disabled?: boolean
  /** 选中样式（CouponPicker 用） */
  selected?: boolean
  className?: string
  style?: CSSProperties
}

interface FromTemplateProps extends BaseProps {
  template: CouponTemplate
  userCoupon?: undefined
}

interface FromUserProps extends BaseProps {
  userCoupon: UserCoupon
  template?: undefined
}

type Props = FromTemplateProps | FromUserProps

const STATUS_LABEL: Record<NonNullable<BaseProps['status']>, string> = {
  unused: '未使用',
  locked: '锁定中',
  used: '已使用',
  expired: '已过期',
  claimed: '已领取',
  claimable: '可领取',
}

function formatRange(from?: string | null, to?: string | null) {
  const fmt = (s?: string | null) => (s ? s.slice(0, 10).replace(/-/g, '/') : '')
  if (!from && !to) return '长期有效'
  return `${fmt(from)} - ${fmt(to)}`
}

export default function CouponItem(props: Props) {
  const { status, action, onClick, disabled, selected, className = '', style } = props

  const data = props.userCoupon
    ? {
        name: props.userCoupon.name,
        description: undefined as string | undefined,
        type: props.userCoupon.type,
        valueCents: props.userCoupon.value_cents,
        discountRate: props.userCoupon.discount_rate ?? null,
        minAmountCents: props.userCoupon.min_amount_cents,
        from: undefined as string | undefined,
        to: props.userCoupon.expire_at,
        scope: undefined as string | undefined,
      }
    : {
        name: props.template.name,
        description: props.template.description,
        type: props.template.type,
        valueCents: props.template.value_cents,
        discountRate: props.template.discount_rate ?? null,
        minAmountCents: props.template.min_amount_cents,
        from: props.template.valid_from ?? undefined,
        to: props.template.valid_to ?? undefined,
        scope: props.template.scope_type === 'all' ? '全场可用' : '部分商品可用',
      }

  const cls = [
    'coupon-item',
    disabled ? 'coupon-item--disabled' : '',
    selected ? 'coupon-item--selected' : '',
    className,
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <View className={cls} style={style} onClick={() => !disabled && onClick?.()}>
      <View className='coupon-item__left'>
        {data.type === 'discount' ? (
          <Text className='coupon-item__amount'>
            {((data.discountRate ?? 1) * 10).toFixed(1)}
            <Text className='coupon-item__amount-unit'>折</Text>
          </Text>
        ) : (
          <Text className='coupon-item__amount'>
            <Text className='coupon-item__amount-unit'>¥</Text>
            {formatPrice(data.valueCents)}
          </Text>
        )}
        <Text className='coupon-item__threshold'>
          {data.minAmountCents > 0 ? `满 ${formatYuan(data.minAmountCents)} 可用` : '无门槛'}
        </Text>
      </View>

      <View className='coupon-item__right'>
        <View className='coupon-item__head'>
          <Text className='coupon-item__name'>{data.name}</Text>
          {status ? (
            <Tag type='primary' plain>
              {STATUS_LABEL[status]}
            </Tag>
          ) : null}
        </View>
        {data.description ? (
          <Text className='coupon-item__desc'>{data.description}</Text>
        ) : null}
        {data.scope ? <Text className='coupon-item__scope'>{data.scope}</Text> : null}
        <Text className='coupon-item__date'>{formatRange(data.from, data.to)}</Text>

        {action ? (
          <View
            className={`coupon-item__action ${action.disabled ? 'coupon-item__action--disabled' : ''}`}
            onClick={(e) => {
              e.stopPropagation?.()
              if (!action.disabled) action.onClick()
            }}
          >
            <Text>{action.text}</Text>
          </View>
        ) : null}
      </View>
    </View>
  )
}
