import { View, Text, Image } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import EvidenceUploader from '@/components/EvidenceUploader'
import {
  AFTERSALE_STATUS_LABEL,
  AFTERSALE_TYPE_LABEL,
  cancelAftersale,
  getAftersaleDetail,
  isAftersaleTerminal,
  postAftersaleMessage,
} from '@/services/aftersale'
import type { AftersaleStatus, AftersaleType } from '@/types/biz'
import { Button, Dialog, Skeleton, TextArea } from '@/ui/nutui'
import { showErrorToast } from '@/utils/error'
import './index.scss'

function formatTime(iso?: string | null) {
  if (!iso) return ''
  return iso.slice(0, 16).replace('T', ' ')
}

interface TimelineNode {
  key: string
  title: string
  time?: string | null
  reached: boolean
}

function buildTimeline(
  type: AftersaleType,
  status: AftersaleStatus,
  timestamps: {
    applied_at?: string
    agreed_at?: string | null
    returned_at?: string | null
    received_at?: string | null
    completed_at?: string | null
  },
): TimelineNode[] {
  const needReturn = type === 'refund_return' || type === 'exchange'

  const all: TimelineNode[] = [
    { key: 'applied', title: '提交申请', time: timestamps.applied_at, reached: true },
  ]

  // 终态：拒绝 / 撤销 / 关闭
  if (status === 'seller_rejected') {
    all.push({
      key: 'rejected',
      title: '商家已拒绝',
      time: timestamps.completed_at,
      reached: true,
    })
    return all
  }
  if (status === 'cancelled' || status === 'closed') {
    all.push({
      key: status,
      title: status === 'cancelled' ? '已撤销' : '已关闭',
      time: timestamps.completed_at,
      reached: true,
    })
    return all
  }

  const reachedAgreed =
    status === 'seller_agreed' ||
    status === 'buyer_returned' ||
    status === 'seller_received' ||
    status === 'completed'

  all.push({
    key: 'agreed',
    title: '商家同意',
    time: timestamps.agreed_at,
    reached: reachedAgreed,
  })

  if (needReturn) {
    const reachedReturned =
      status === 'buyer_returned' ||
      status === 'seller_received' ||
      status === 'completed'
    const reachedReceived =
      status === 'seller_received' || status === 'completed'

    all.push({
      key: 'returned',
      title: '买家寄回',
      time: timestamps.returned_at,
      reached: reachedReturned,
    })
    all.push({
      key: 'received',
      title: '商家收货',
      time: timestamps.received_at,
      reached: reachedReceived,
    })
  }

  all.push({
    key: 'completed',
    title: type === 'exchange' ? '换货完成' : '退款完成',
    time: timestamps.completed_at,
    reached: status === 'completed',
  })

  return all
}

export default function AftersaleDetailPage() {
  const id = Taro.getCurrentInstance().router?.params?.id ?? ''
  const queryClient = useQueryClient()

  const detailQuery = useQuery({
    queryKey: ['aftersale-detail', id],
    queryFn: () => getAftersaleDetail(id),
    enabled: id.length > 0,
  })

  const [content, setContent] = useState('')
  const [evidence, setEvidence] = useState<string[]>([])
  const [cancelOpen, setCancelOpen] = useState(false)

  const messageMutation = useMutation({
    mutationFn: () =>
      postAftersaleMessage(id, {
        content: content.trim() || undefined,
        evidence: evidence.length > 0 ? evidence : undefined,
      }),
    onSuccess: () => {
      setContent('')
      setEvidence([])
      void queryClient.invalidateQueries({ queryKey: ['aftersale-detail', id] })
      void Taro.showToast({ title: '已发送', icon: 'success' })
    },
    onError: (err) => showErrorToast(err, '发送失败'),
  })

  const cancelMutation = useMutation({
    mutationFn: () => cancelAftersale(id),
    onSuccess: () => {
      setCancelOpen(false)
      void queryClient.invalidateQueries({ queryKey: ['aftersale-detail', id] })
      void queryClient.invalidateQueries({ queryKey: ['aftersale-list'] })
      void Taro.showToast({ title: '已撤销申请', icon: 'success' })
    },
    onError: (err) => showErrorToast(err, '撤销失败'),
  })

  const order = detailQuery.data
  const timeline = useMemo(() => {
    if (!order) return []
    return buildTimeline(order.type, order.status, {
      applied_at: order.applied_at,
      agreed_at: order.agreed_at,
      returned_at: order.returned_at,
      received_at: order.received_at,
      completed_at: order.completed_at,
    })
  }, [order])

  if (detailQuery.isLoading || !order) {
    return (
      <View className='page-shell aftersale-detail-page'>
        <Skeleton animated rows={8} />
      </View>
    )
  }

  const terminal = isAftersaleTerminal(order.status)
  const canCancel = order.status === 'applying'
  const canSubmitExpress =
    order.status === 'seller_agreed' &&
    (order.type === 'refund_return' || order.type === 'exchange')
  const canAppendMessage = !terminal

  function handleSendMessage() {
    if (!content.trim() && evidence.length === 0) {
      void Taro.showToast({ title: '请填写内容或上传凭证', icon: 'none' })
      return
    }
    messageMutation.mutate()
  }

  return (
    <View className='page-shell aftersale-detail-page'>
      {/* 状态卡 */}
      <View className='aftersale-detail-page__status-card'>
        <Text className='aftersale-detail-page__status-text'>
          {AFTERSALE_STATUS_LABEL[order.status]}
        </Text>
        <View>
          <Text className='aftersale-detail-page__status-sub'>
            {AFTERSALE_TYPE_LABEL[order.type]} · 售后单号 {order.aftersale_no}
          </Text>
        </View>
      </View>

      {/* 时间线 */}
      <View className='aftersale-detail-page__section'>
        <Text className='aftersale-detail-page__section-title'>处理进度</Text>
        <View className='aftersale-detail-page__timeline'>
          {timeline.map((node) => (
            <View key={node.key} className='aftersale-detail-page__timeline-step'>
              <View
                className={`aftersale-detail-page__timeline-dot${
                  node.reached ? ' aftersale-detail-page__timeline-dot--active' : ''
                }`}
              />
              <View className='aftersale-detail-page__timeline-content'>
                <Text className='aftersale-detail-page__timeline-title'>{node.title}</Text>
                {node.time && (
                  <View>
                    <Text className='aftersale-detail-page__timeline-time'>
                      {formatTime(node.time)}
                    </Text>
                  </View>
                )}
              </View>
            </View>
          ))}
        </View>
      </View>

      {/* 基础信息 */}
      <View className='aftersale-detail-page__section'>
        <Text className='aftersale-detail-page__section-title'>申请信息</Text>
        <View className='aftersale-detail-page__meta-row'>
          <Text>关联订单</Text>
          <Text>{order.order_no}</Text>
        </View>
        <View className='aftersale-detail-page__meta-row'>
          <Text>退款金额</Text>
          <Text>¥{(order.refund_amount_cents / 100).toFixed(2)}</Text>
        </View>
        <View className='aftersale-detail-page__meta-row'>
          <Text>申请原因</Text>
          <Text>{order.reason}</Text>
        </View>
        {order.buyer_evidence && order.buyer_evidence.length > 0 && (
          <View className='aftersale-detail-page__evidence-thumbs'>
            {order.buyer_evidence.map((url) => (
              <Image
                key={url}
                src={url}
                mode='aspectFill'
                className='aftersale-detail-page__evidence-thumb'
              />
            ))}
          </View>
        )}
      </View>

      {/* 寄回信息 */}
      {order.buyer_express && (
        <View className='aftersale-detail-page__section'>
          <Text className='aftersale-detail-page__section-title'>寄回物流</Text>
          <View className='aftersale-detail-page__meta-row'>
            <Text>快递公司</Text>
            <Text>{order.buyer_express.carrier_code}</Text>
          </View>
          <View className='aftersale-detail-page__meta-row'>
            <Text>运单号</Text>
            <Text>{order.buyer_express.waybill_no}</Text>
          </View>
        </View>
      )}

      {/* 协商记录 */}
      <View className='aftersale-detail-page__section'>
        <Text className='aftersale-detail-page__section-title'>协商记录</Text>
        {(order.negotiations ?? []).length === 0 && (
          <Text className='aftersale-detail-page__timeline-time'>暂无</Text>
        )}
        {(order.negotiations ?? []).map((n) => (
          <View key={n.id} className='aftersale-detail-page__nego-item'>
            <View className='aftersale-detail-page__nego-header'>
              <Text
                className={`aftersale-detail-page__nego-role aftersale-detail-page__nego-role--${n.role}`}
              >
                {n.role === 'buyer' ? '我' : n.role === 'seller' ? '商家' : '系统'}
              </Text>
              <Text>{formatTime(n.created_at)}</Text>
            </View>
            <Text className='aftersale-detail-page__nego-content'>{n.content}</Text>
            {n.evidence && n.evidence.length > 0 && (
              <View className='aftersale-detail-page__nego-images'>
                {n.evidence.map((url) => (
                  <Image
                    key={url}
                    src={url}
                    mode='aspectFill'
                    className='aftersale-detail-page__nego-img'
                  />
                ))}
              </View>
            )}
          </View>
        ))}
      </View>

      {/* 追加留言 */}
      {canAppendMessage && (
        <View className='aftersale-detail-page__compose'>
          <Text className='aftersale-detail-page__section-title'>追加留言 / 凭证</Text>
          <TextArea
            placeholder='补充说明（选填，最多 1000 字）'
            value={content}
            onChange={(v) => setContent(String(v))}
            maxLength={1000}
            rows={3}
          />
          <View style={{ marginTop: '10px' }}>
            <EvidenceUploader value={evidence} onChange={setEvidence} max={6} />
          </View>
          <View className='aftersale-detail-page__compose-actions'>
            <Button
              type='primary'
              size='small'
              loading={messageMutation.isPending}
              onClick={handleSendMessage}
            >
              发送
            </Button>
          </View>
        </View>
      )}

      {/* 底部操作栏 */}
      {(canCancel || canSubmitExpress) && (
        <View className='aftersale-detail-page__action-bar'>
          <View className='aftersale-detail-page__action-row'>
            {canCancel && (
              <Button
                type='default'
                plain
                block
                onClick={() => setCancelOpen(true)}
              >
                撤销申请
              </Button>
            )}
            {canSubmitExpress && (
              <Button
                type='primary'
                block
                onClick={() =>
                  void Taro.navigateTo({
                    url: `/pages/aftersale/express/index?id=${order.id}`,
                  })
                }
              >
                去寄回
              </Button>
            )}
          </View>
        </View>
      )}

      <Dialog
        title='撤销售后申请'
        visible={cancelOpen}
        confirmText={cancelMutation.isPending ? '提交中...' : '确认撤销'}
        cancelText='再想想'
        onConfirm={() => cancelMutation.mutate()}
        onCancel={() => setCancelOpen(false)}
      >
        <View style={{ padding: '12px 0' }}>
          撤销后将关闭本次售后单，是否继续？
        </View>
      </Dialog>
    </View>
  )
}
