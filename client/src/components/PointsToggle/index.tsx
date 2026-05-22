import { useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import { Switch } from '@/ui/nutui'
import { formatYuan } from '@/utils/price'
import './index.scss'

interface PointsToggleProps {
  /** 用户当前可用积分余额 */
  balance: number
  /** 此订单允许使用的最大积分数（来自 quote.max_point_used） */
  maxUsable: number
  /** 抵扣比例：1 积分 = X 分人民币（来自规则，默认 1） */
  pointToCent?: number
  /** 当前选中使用的积分 */
  value: number
  onChange: (points: number) => void
  /** 当前抵扣金额（cents） */
  deductCents?: number
  disabled?: boolean
  disabledReason?: string
}

export default function PointsToggle({
  balance,
  maxUsable,
  value,
  onChange,
  deductCents = value,
  disabled,
  disabledReason,
}: PointsToggleProps) {
  // 默认：开启时全部用满 maxUsable
  const [enabled, setEnabled] = useState(value > 0)

  useEffect(() => {
    setEnabled(value > 0)
  }, [value])

  const canUse = !disabled && maxUsable > 0 && balance > 0

  const handleSwitch = (checked: boolean) => {
    if (!canUse) return
    setEnabled(checked)
    if (checked) {
      onChange(Math.min(maxUsable, balance))
    } else {
      onChange(0)
    }
  }

  return (
    <View className='points-toggle'>
      <View className='points-toggle__row'>
        <View className='points-toggle__main'>
          <Text className='points-toggle__title'>积分抵扣</Text>
          <Text className='points-toggle__sub'>
            {canUse
              ? `可用 ${balance} 分，本单最多用 ${maxUsable} 分`
              : disabledReason || '本订单暂不可用积分'}
          </Text>
        </View>
        <Switch
          checked={enabled && canUse}
          disabled={!canUse}
          onChange={handleSwitch}
        />
      </View>

      {enabled && canUse && deductCents > 0 ? (
        <View className='points-toggle__hint'>
          <Text>
            已使用 {value} 分，抵扣
          </Text>
          <Text className='points-toggle__amount'>-{formatYuan(deductCents)}</Text>
        </View>
      ) : null}
    </View>
  )
}
