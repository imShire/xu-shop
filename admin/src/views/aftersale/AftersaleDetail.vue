<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import {
  getAftersaleDetail,
  agreeAftersale,
  rejectAftersale,
  confirmReceivedAftersale,
  postAftersaleMessage,
  closeAftersale,
} from '@/api/aftersale'
import {
  AFTERSALE_STATUS_LABEL,
  AFTERSALE_TYPE_LABEL,
  NEGOTIATION_ROLE_LABEL,
  isTerminal,
  type AftersaleOrderDetail,
  type AftersaleStatus,
} from '@/types/aftersale'
import { formatAmount, formatTime } from '@/utils/format'
import { useAuthStore } from '@/stores/auth'
import UploadImage from '@/components/UploadImage/index.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const id = String(route.params.id || '')
const detail = ref<AftersaleOrderDetail | null>(null)
const loading = ref(false)

const hasProcessPerm = computed(
  () => auth.isSuperAdmin || auth.perms.includes('aftersale.process'),
)

async function load() {
  if (!id) return
  loading.value = true
  try {
    detail.value = await getAftersaleDetail(id)
  } catch (e: any) {
    ElMessage.error(e?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

// ============ 状态机派生 ============
const status = computed<AftersaleStatus | undefined>(() => detail.value?.status)

const canAgree = computed(() => status.value === 'applying')
const canReject = computed(() => status.value === 'applying')
const canConfirmReceived = computed(() => status.value === 'buyer_returned')
// 任何非终态都可手动关闭（兜底）
const canClose = computed(() => status.value && !isTerminal(status.value))
// 终态 completed 不允许追加协商；其他都允许（含 closed/cancelled/rejected 之前的非终态）
const canSendMessage = computed(() => status.value && status.value !== 'completed')

// ============ 时间线 ============
interface TimelineNode {
  ts?: string | null
  label: string
  type: 'primary' | 'success' | 'warning' | 'danger' | 'info'
}
const timeline = computed<TimelineNode[]>(() => {
  const d = detail.value
  if (!d) return []
  const nodes: TimelineNode[] = [
    { ts: d.applied_at, label: '买家发起售后', type: 'primary' },
  ]
  if (d.agreed_at) nodes.push({ ts: d.agreed_at, label: '商家同意', type: 'success' })
  if (d.returned_at) nodes.push({ ts: d.returned_at, label: '买家寄回', type: 'primary' })
  if (d.received_at) nodes.push({ ts: d.received_at, label: '商家确认收货', type: 'success' })
  if (d.completed_at) nodes.push({ ts: d.completed_at, label: '售后完成', type: 'success' })
  if (d.closed_at) nodes.push({ ts: d.closed_at, label: '售后关闭', type: 'info' })
  if (d.status === 'seller_rejected')
    nodes.push({ ts: d.agreed_at, label: '商家已拒绝', type: 'danger' })
  return nodes.filter((n) => n.ts)
})

// ============ 同意 ============
const agreeDialog = ref({ visible: false, seller_remark: '' })
const agreeLoading = ref(false)

function openAgree() {
  if (!detail.value) return
  agreeDialog.value = { visible: true, seller_remark: '' }
}

async function submitAgree() {
  if (!detail.value) return
  const tip =
    detail.value.type === 'refund_only'
      ? '同意后将异步发起原路退款，操作不可撤销。'
      : '同意后流程进入“等待买家寄回”。'
  try {
    await ElMessageBox.confirm(tip, '同意售后', { type: 'warning' })
  } catch {
    return
  }
  agreeLoading.value = true
  try {
    await agreeAftersale(detail.value.id, { seller_remark: agreeDialog.value.seller_remark }, crypto.randomUUID())
    ElMessage.success('已同意')
    agreeDialog.value.visible = false
    await load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    agreeLoading.value = false
  }
}

// ============ 拒绝 ============
const rejectDialog = ref({ visible: false, reason: '' })
const rejectLoading = ref(false)

function openReject() {
  rejectDialog.value = { visible: true, reason: '' }
}

async function submitReject() {
  if (!detail.value) return
  if (rejectDialog.value.reason.trim().length < 2) {
    ElMessage.warning('请输入至少 2 个字的拒绝原因')
    return
  }
  rejectLoading.value = true
  try {
    await rejectAftersale(detail.value.id, rejectDialog.value.reason.trim())
    ElMessage.success('已拒绝')
    rejectDialog.value.visible = false
    await load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    rejectLoading.value = false
  }
}

// ============ 确认收货 ============
const confirmDialog = ref({ visible: false, seller_remark: '' })
const confirmLoading = ref(false)

function openConfirmReceived() {
  confirmDialog.value = { visible: true, seller_remark: '' }
}

async function submitConfirmReceived() {
  if (!detail.value) return
  try {
    await ElMessageBox.confirm(
      detail.value.type === 'exchange'
        ? '确认收货后售后单将直接置为已完成。'
        : '确认收货后将异步发起原路退款，操作不可撤销。',
      '确认收货',
      { type: 'warning' },
    )
  } catch {
    return
  }
  confirmLoading.value = true
  try {
    await confirmReceivedAftersale(
      detail.value.id,
      { seller_remark: confirmDialog.value.seller_remark },
      crypto.randomUUID(),
    )
    ElMessage.success('已确认收货')
    confirmDialog.value.visible = false
    await load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    confirmLoading.value = false
  }
}

// ============ 关闭 ============
const closeDialog = ref({ visible: false, reason: '' })
const closeLoading = ref(false)

function openClose() {
  closeDialog.value = { visible: true, reason: '' }
}

async function submitClose() {
  if (!detail.value) return
  if (closeDialog.value.reason.trim().length < 2) {
    ElMessage.warning('请输入至少 2 个字的关闭原因')
    return
  }
  try {
    await ElMessageBox.confirm('手动关闭后售后单不可再操作，确认吗？', '关闭售后单', { type: 'warning' })
  } catch {
    return
  }
  closeLoading.value = true
  try {
    await closeAftersale(detail.value.id, closeDialog.value.reason.trim())
    ElMessage.success('已关闭')
    closeDialog.value.visible = false
    await load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    closeLoading.value = false
  }
}

// ============ 协商留言 ============
const messageDialog = ref<{ visible: boolean; content: string; evidence: string[] }>({
  visible: false,
  content: '',
  evidence: [],
})
const messageLoading = ref(false)

function openMessage() {
  messageDialog.value = { visible: true, content: '', evidence: [] }
}

async function submitMessage() {
  if (!detail.value) return
  const content = messageDialog.value.content.trim()
  const evidence = messageDialog.value.evidence
  if (!content && evidence.length === 0) {
    ElMessage.warning('内容与凭证至少填写一项')
    return
  }
  if (content.length > 1000) {
    ElMessage.warning('内容不超过 1000 字')
    return
  }
  messageLoading.value = true
  try {
    await postAftersaleMessage(detail.value.id, {
      content: content || undefined,
      evidence: evidence.length ? evidence : undefined,
    })
    ElMessage.success('已发送')
    messageDialog.value.visible = false
    await load()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    messageLoading.value = false
  }
}

const previewVisible = ref(false)
const previewList = ref<string[]>([])
const previewIndex = ref(0)
function preview(images: string[], idx: number) {
  previewList.value = images
  previewIndex.value = idx
  previewVisible.value = true
}

onMounted(() => load())
</script>

<template>
  <div class="page-card" v-loading="loading">
    <div class="header">
      <el-button :icon="ArrowLeft" text @click="router.back()">返回</el-button>
      <span class="title">售后详情</span>
      <span v-if="detail" class="sub">
        {{ detail.aftersale_no }}
      </span>
    </div>

    <template v-if="detail">
      <!-- 基础信息 -->
      <el-descriptions :column="3" border title="售后单信息" class="block">
        <el-descriptions-item label="售后单号">{{ detail.aftersale_no }}</el-descriptions-item>
        <el-descriptions-item label="关联订单">
          <el-link type="primary" @click="router.push(`/order/detail/${detail.order_id}`)">
            {{ detail.order_no }}
          </el-link>
        </el-descriptions-item>
        <el-descriptions-item label="用户">
          <el-link @click="router.push(`/user/${detail.user_id}`)">{{ detail.user_id }}</el-link>
        </el-descriptions-item>
        <el-descriptions-item label="类型">
          <el-tag size="small">{{ AFTERSALE_TYPE_LABEL[detail.type] }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="AFTERSALE_STATUS_LABEL[detail.status]?.type || ''" size="small">
            {{ AFTERSALE_STATUS_LABEL[detail.status]?.label }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="退款金额">
          <span style="color: #f59e0b">{{ formatAmount(detail.refund_amount_cents) }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ formatTime(detail.applied_at) }}</el-descriptions-item>
        <el-descriptions-item label="超时关闭">{{ formatTime(detail.auto_close_at) }}</el-descriptions-item>
        <el-descriptions-item label="售后单范围">
          {{ detail.order_item_id ? '单 SKU 售后' : '整单售后' }}
        </el-descriptions-item>
        <el-descriptions-item label="申请原因" :span="3">
          {{ detail.reason }}
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.seller_remark" label="商家备注" :span="3">
          {{ detail.seller_remark }}
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.buyer_express" label="买家寄回" :span="3">
          {{ detail.buyer_express.carrier_code }} · {{ detail.buyer_express.waybill_no }}
          <span v-if="detail.buyer_express.shipped_at" style="color: #999; margin-left: 8px">
            （{{ formatTime(detail.buyer_express.shipped_at) }}）
          </span>
        </el-descriptions-item>
        <el-descriptions-item v-if="detail.buyer_evidence?.length" label="买家凭证" :span="3">
          <div class="evidence-row">
            <img
              v-for="(img, idx) in detail.buyer_evidence"
              :key="idx"
              :src="img"
              class="evidence-img"
              @click="preview(detail.buyer_evidence, idx)"
            />
          </div>
        </el-descriptions-item>
      </el-descriptions>

      <!-- 商品快照 -->
      <el-card v-if="detail.item_snapshot" shadow="never" class="block">
        <template #header>商品信息</template>
        <div class="item-snapshot">
          <img v-if="detail.item_snapshot.image" :src="detail.item_snapshot.image as string" />
          <div class="meta">
            <div class="name">{{ detail.item_snapshot.product_name }}</div>
            <div class="sku">{{ detail.item_snapshot.sku_attrs }}</div>
            <div class="price">
              <span style="color: #f59e0b">{{
                formatAmount(Number(detail.item_snapshot.price_cents || 0))
              }}</span>
              <span style="margin-left: 12px; color: #999">× {{ detail.item_snapshot.qty }}</span>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 状态时间线 -->
      <el-card shadow="never" class="block">
        <template #header>处理进度</template>
        <el-timeline>
          <el-timeline-item
            v-for="(node, idx) in timeline"
            :key="idx"
            :type="node.type"
            :timestamp="formatTime(node.ts)"
          >
            {{ node.label }}
          </el-timeline-item>
        </el-timeline>
      </el-card>

      <!-- 协商记录 -->
      <el-card shadow="never" class="block">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center">
            <span>协商记录（{{ detail.negotiations?.length || 0 }}）</span>
            <el-button
              v-if="canSendMessage && hasProcessPerm"
              type="primary"
              size="small"
              @click="openMessage"
            >
              发送协商消息
            </el-button>
          </div>
        </template>
        <div v-if="!detail.negotiations?.length" class="empty">暂无协商记录</div>
        <div v-else class="negotiation-list">
          <div
            v-for="n in detail.negotiations"
            :key="n.id"
            class="negotiation-item"
            :class="`role-${n.role}`"
          >
            <div class="head">
              <el-tag
                size="small"
                :type="n.role === 'seller' ? 'success' : n.role === 'buyer' ? 'primary' : 'info'"
              >
                {{ NEGOTIATION_ROLE_LABEL[n.role] }}
              </el-tag>
              <span class="time">{{ formatTime(n.created_at) }}</span>
            </div>
            <div v-if="n.content" class="content">{{ n.content }}</div>
            <div v-if="n.evidence?.length" class="evidence-row">
              <img
                v-for="(img, idx) in n.evidence"
                :key="idx"
                :src="img"
                class="evidence-img"
                @click="preview(n.evidence, idx)"
              />
            </div>
          </div>
        </div>
      </el-card>

      <!-- 操作栏 -->
      <el-card v-if="hasProcessPerm" shadow="never" class="block">
        <template #header>操作</template>
        <div class="actions">
          <el-button v-if="canAgree" type="success" @click="openAgree">同意</el-button>
          <el-button v-if="canReject" type="danger" @click="openReject">拒绝</el-button>
          <el-button v-if="canConfirmReceived" type="primary" @click="openConfirmReceived">
            确认收货
          </el-button>
          <el-button v-if="canClose" type="warning" plain @click="openClose">手动关闭</el-button>
          <el-tag v-if="status === 'seller_agreed' && detail.type !== 'refund_only'" type="info">
            等待买家寄回
          </el-tag>
          <el-tag v-if="isTerminal(detail.status)" type="info">已结束，仅可查看</el-tag>
        </div>
      </el-card>
      <el-alert
        v-else
        type="info"
        :closable="false"
        title="只读模式：当前账号缺少 aftersale.process 权限，无法执行操作"
      />
    </template>

    <!-- 同意 -->
    <el-dialog v-model="agreeDialog.visible" title="同意售后" width="500px">
      <el-form label-width="80px">
        <el-form-item label="备注">
          <el-input
            v-model="agreeDialog.seller_remark"
            type="textarea"
            :rows="3"
            placeholder="选填，将写入协商记录"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="agreeDialog.visible = false">取消</el-button>
        <el-button type="success" :loading="agreeLoading" @click="submitAgree">确认同意</el-button>
      </template>
    </el-dialog>

    <!-- 拒绝 -->
    <el-dialog v-model="rejectDialog.visible" title="拒绝售后" width="500px">
      <el-form label-width="80px">
        <el-form-item label="拒绝原因" required>
          <el-input
            v-model="rejectDialog.reason"
            type="textarea"
            :rows="3"
            placeholder="2-200 字"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialog.visible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="submitReject">确认拒绝</el-button>
      </template>
    </el-dialog>

    <!-- 确认收货 -->
    <el-dialog v-model="confirmDialog.visible" title="确认收货" width="500px">
      <el-form label-width="80px">
        <el-form-item label="备注">
          <el-input
            v-model="confirmDialog.seller_remark"
            type="textarea"
            :rows="3"
            placeholder="选填"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="confirmDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="confirmLoading" @click="submitConfirmReceived">
          确认收货
        </el-button>
      </template>
    </el-dialog>

    <!-- 手动关闭 -->
    <el-dialog v-model="closeDialog.visible" title="手动关闭售后单" width="500px">
      <el-form label-width="80px">
        <el-form-item label="关闭原因" required>
          <el-input
            v-model="closeDialog.reason"
            type="textarea"
            :rows="3"
            placeholder="2-200 字，将写入审计日志"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDialog.visible = false">取消</el-button>
        <el-button type="warning" :loading="closeLoading" @click="submitClose">确认关闭</el-button>
      </template>
    </el-dialog>

    <!-- 协商留言 -->
    <el-dialog v-model="messageDialog.visible" title="发送协商消息" width="600px">
      <el-form label-width="80px">
        <el-form-item label="内容">
          <el-input
            v-model="messageDialog.content"
            type="textarea"
            :rows="4"
            placeholder="内容与凭证至少填写一项，最多 1000 字"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="凭证图片">
          <UploadImage v-model="messageDialog.evidence" multiple :limit="6" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="messageDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="messageLoading" @click="submitMessage">
          发送
        </el-button>
      </template>
    </el-dialog>

    <!-- 图片预览 -->
    <el-image-viewer
      v-if="previewVisible"
      :url-list="previewList"
      :initial-index="previewIndex"
      @close="previewVisible = false"
    />
  </div>
</template>

<style scoped lang="scss">
.header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  .title {
    font-size: 18px;
    font-weight: 600;
  }
  .sub {
    color: #999;
    font-size: 13px;
  }
}
.block {
  margin-bottom: 16px;
}
.evidence-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.evidence-img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid #eee;
}
.item-snapshot {
  display: flex;
  gap: 16px;
  img {
    width: 80px;
    height: 80px;
    object-fit: cover;
    border-radius: 4px;
  }
  .meta {
    flex: 1;
    .name {
      font-weight: 500;
      margin-bottom: 4px;
    }
    .sku {
      color: #999;
      font-size: 13px;
      margin-bottom: 4px;
    }
  }
}
.empty {
  color: #999;
  text-align: center;
  padding: 24px 0;
}
.negotiation-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.negotiation-item {
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  padding: 12px;
  background: #fafafa;
  &.role-seller {
    background: #f0f9eb;
    border-color: #e1f3d8;
  }
  &.role-system {
    background: #f4f4f5;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
    .time {
      color: #999;
      font-size: 12px;
    }
  }
  .content {
    line-height: 1.6;
    margin-bottom: 8px;
    white-space: pre-wrap;
    word-break: break-word;
  }
}
.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
</style>
