import dayjs from 'dayjs'

export const centsToYuan = (cents: number): string => (cents / 100).toFixed(2)

export const yuanToCents = (yuan: number | string): number =>
  Math.round(Number(yuan) * 100)

export const formatTime = (ts: string | number | undefined | null): string => {
  if (!ts) return '-'
  return dayjs(ts).format('YYYY-MM-DD HH:mm:ss')
}

export const formatDate = (ts: string | number | undefined | null): string => {
  if (!ts) return '-'
  return dayjs(ts).format('YYYY-MM-DD')
}

export const formatAmount = (cents: number): string => `¥${centsToYuan(cents)}`

export const orderStatusMap: Record<string, { label: string; type: string }> = {
  pending: { label: '待付款', type: 'warning' },
  paid: { label: '待发货', type: 'primary' },
  shipped: { label: '已发货', type: '' },
  completed: { label: '已完成', type: 'success' },
  cancelled: { label: '已取消', type: 'info' },
  refunding: { label: '退款中', type: 'danger' },
  refunded: { label: '已退款', type: 'danger' },
}

export const productStatusMap: Record<string, { label: string; type: string }> = {
  draft: { label: '草稿', type: 'info' },
  onsale: { label: '在售', type: 'success' },
  offsale: { label: '下架', type: 'warning' },
}
