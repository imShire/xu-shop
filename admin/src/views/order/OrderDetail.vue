<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrderDetail, addOrderRemark } from '@/api/order'
import { createRefund } from '@/api/payment'
import { shipOrder } from '@/api/shipping'
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format'
import PriceInput from '@/components/PriceInput/index.vue'

const route = useRoute()
const router = useRouter()
const orderId = String(route.params.id)

const loading = ref(false)
const order = ref<any>(null)

const remarkDialogVisible = ref(false)
const remarkContent = ref('')
const remarkLoading = ref(false)

const refundDialogVisible = ref(false)
const refundAmount = ref(0)
const refundReason = ref('')
const refundLoading = ref(false)

const shipDialogVisible = ref(false)
const shipForm = ref({ carrier_code: '', tracking_no: '' })
const shipLoading = ref(false)

async function loadOrder() {
  loading.value = true
  try {
    order.value = await getOrderDetail(orderId)
  } finally {
    loading.value = false
  }
}

async function handleAddRemark() {
  if (!remarkContent.value.trim()) return ElMessage.warning('请输入备注内容')
  remarkLoading.value = true
  try {
    await addOrderRemark(orderId, remarkContent.value)
    ElMessage.success('备注添加成功')
    remarkDialogVisible.value = false
    await loadOrder()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    remarkLoading.value = false
  }
}

async function handleRefund() {
  refundLoading.value = true
  try {
    await ElMessageBox.confirm(`确认退款 ${formatAmount(refundAmount.value)} 元？此操作不可撤销`, '退款确认', { type: 'warning' })
    await createRefund(orderId, { amount_cents: refundAmount.value, reason: refundReason.value })
    ElMessage.success('退款申请已提交')
    refundDialogVisible.value = false
    await loadOrder()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    refundLoading.value = false
  }
}

async function handleShip() {
  if (!shipForm.value.carrier_code || !shipForm.value.tracking_no) {
    return ElMessage.warning('请填写快递信息')
  }
  shipLoading.value = true
  try {
    await ElMessageBox.confirm('确认发货？发货后无法撤回', '发货确认', { type: 'warning' })
    await shipOrder(orderId, shipForm.value)
    ElMessage.success('发货成功')
    shipDialogVisible.value = false
    await loadOrder()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    shipLoading.value = false
  }
}

const logSteps = [
  { status: 'pending', label: '待付款' },
  { status: 'paid', label: '已付款' },
  { status: 'shipped', label: '已发货' },
  { status: 'completed', label: '已完成' },
]

function getActiveStep(status: string) {
  const idx = logSteps.findIndex((s) => s.status === status)
  return idx >= 0 ? idx : 0
}

onMounted(loadOrder)
</script>

<template>
  <div class="page-container" v-loading="loading">
    <div style="display: flex; align-items: center; gap: 16px; margin-bottom: 16px">
      <el-button text :icon="'ArrowLeft'" @click="router.back()">返回</el-button>
      <h3 style="font-size: 16px; font-weight: 600">订单详情</h3>
    </div>

    <template v-if="order">
      <!-- 状态时间线 -->
      <el-card style="margin-bottom: 16px">
        <el-steps :active="getActiveStep(order.status)" align-center>
          <el-step
            v-for="step in logSteps"
            :key="step.status"
            :title="step.label"
          />
        </el-steps>

        <div style="display: flex; justify-content: flex-end; gap: 8px; margin-top: 16px">
          <el-tag :type="orderStatusMap[order.status]?.type || ''" size="large">
            {{ orderStatusMap[order.status]?.label || order.status }}
          </el-tag>
          <el-button
            v-if="order.status === 'paid'"
            type="primary"
            size="small"
            @click="shipDialogVisible = true"
          >
            发货
          </el-button>
          <el-button
            v-if="['paid', 'shipped'].includes(order.status)"
            type="warning"
            size="small"
            @click="refundDialogVisible = true; refundAmount = order.pay_cents"
          >
            退款
          </el-button>
          <el-button size="small" @click="remarkDialogVisible = true">添加备注</el-button>
        </div>
      </el-card>

      <el-row :gutter="16">
        <!-- 商品明细 -->
        <el-col :span="16">
          <el-card style="margin-bottom: 16px">
            <template #header>商品明细</template>
            <div v-for="item in order.items" :key="item.id" style="display: flex; gap: 12px; padding: 10px 0; border-bottom: 1px solid var(--border-color)">
              <img v-if="item.product_snapshot?.image" :src="item.product_snapshot.image" style="width: 60px; height: 60px; border-radius: 4px; object-fit: cover" />
              <div style="flex: 1">
                <div style="font-weight: 500">{{ item.product_snapshot?.title }}</div>
                <div style="font-size: 12px; color: var(--text-secondary)">
                  {{ item.product_snapshot?.attrs ? Object.values(item.product_snapshot.attrs).join(' / ') : '' }}
                </div>
                <div style="display: flex; justify-content: space-between; margin-top: 4px">
                  <span style="color: #f59e0b">{{ formatAmount(item.price_cents) }}</span>
                  <span>× {{ item.qty }}</span>
                  <span style="font-weight: 600">{{ formatAmount(item.price_cents * item.qty) }}</span>
                </div>
              </div>
            </div>

            <div style="margin-top: 12px; text-align: right">
              <div style="color: var(--text-secondary)">运费：{{ formatAmount(order.freight_cents) }}</div>
              <div style="color: var(--text-secondary)">优惠：-{{ formatAmount(order.coupon_discount_cents) }}</div>
              <div style="font-size: 16px; font-weight: 700; color: #f59e0b; margin-top: 8px">
                实付：{{ formatAmount(order.pay_cents) }}
              </div>
            </div>
          </el-card>

          <!-- 操作日志 -->
          <el-card v-if="order.logs?.length">
            <template #header>操作记录</template>
            <el-timeline>
              <el-timeline-item
                v-for="log in order.logs"
                :key="log.id"
                :timestamp="formatTime(log.created_at)"
                placement="top"
              >
                <div>{{ [log.from_status, log.to_status].filter(Boolean).join(' → ') }}{{ log.reason ? `（${log.reason}）` : '' }}</div>
                <div style="font-size: 12px; color: var(--text-secondary)">{{ log.operator_type }}</div>
              </el-timeline-item>
            </el-timeline>
          </el-card>

          <!-- 备注 -->
          <el-card v-if="order.remarks?.length" style="margin-top: 16px">
            <template #header>备注列表</template>
            <div v-for="r in order.remarks" :key="r.id" style="padding: 8px 0; border-bottom: 1px solid var(--border-color)">
              <div>{{ r.content }}</div>
              <div style="font-size: 12px; color: var(--text-secondary)">{{ formatTime(r.created_at) }}</div>
            </div>
          </el-card>
        </el-col>

        <!-- 右侧信息 -->
        <el-col :span="8">
          <!-- 收货地址 -->
          <el-card style="margin-bottom: 16px">
            <template #header>收货地址</template>
            <el-descriptions :column="1" size="small">
              <el-descriptions-item label="收货人">{{ order.address_snapshot?.name }}</el-descriptions-item>
              <el-descriptions-item label="手机号">{{ order.address_snapshot?.phone }}</el-descriptions-item>
              <el-descriptions-item label="地址">
                {{ order.address_snapshot?.province }}{{ order.address_snapshot?.city }}{{ order.address_snapshot?.district }}{{ order.address_snapshot?.street }}{{ order.address_snapshot?.detail }}
              </el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- 支付信息 -->
          <el-card style="margin-bottom: 16px">
            <template #header>支付信息</template>
            <el-descriptions :column="1" size="small">
              <el-descriptions-item label="下单时间">{{ formatTime(order.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="付款时间">{{ formatTime(order.paid_at) }}</el-descriptions-item>
              <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
            </el-descriptions>
          </el-card>

          <!-- 物流信息 -->
          <el-card v-if="order.shipment">
            <template #header>物流信息</template>
            <el-descriptions :column="1" size="small">
              <el-descriptions-item label="快递公司">{{ order.shipment.carrier }}</el-descriptions-item>
              <el-descriptions-item label="运单号">{{ order.shipment.tracking_no }}</el-descriptions-item>
              <el-descriptions-item label="发货时间">{{ formatTime(order.shipment.shipped_at) }}</el-descriptions-item>
            </el-descriptions>
            <div v-if="order.shipment.tracks?.length" style="margin-top: 12px">
              <el-timeline>
                <el-timeline-item
                  v-for="track in order.shipment.tracks"
                  :key="track.time"
                  :timestamp="track.time"
                  size="small"
                >
                  {{ track.content }}
                </el-timeline-item>
              </el-timeline>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>

    <!-- 备注对话框 -->
    <el-dialog v-model="remarkDialogVisible" title="添加备注" width="400px">
      <el-input v-model="remarkContent" type="textarea" :rows="4" placeholder="请输入备注内容" />
      <template #footer>
        <el-button @click="remarkDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="remarkLoading" @click="handleAddRemark">确认</el-button>
      </template>
    </el-dialog>

    <!-- 退款对话框 -->
    <el-dialog v-model="refundDialogVisible" title="申请退款" width="400px">
      <el-form label-width="90px">
        <el-form-item label="退款金额">
          <PriceInput v-model="refundAmount" />
        </el-form-item>
        <el-form-item label="退款原因">
          <el-input v-model="refundReason" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialogVisible = false">取消</el-button>
        <el-button type="warning" :loading="refundLoading" @click="handleRefund">确认退款</el-button>
      </template>
    </el-dialog>

    <!-- 发货对话框 -->
    <el-dialog v-model="shipDialogVisible" title="手动发货" width="400px">
      <el-form :model="shipForm" label-width="90px">
        <el-form-item label="快递公司">
          <el-input v-model="shipForm.carrier_code" placeholder="如：顺丰" />
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="shipForm.tracking_no" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="handleShip">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
  max-width: 1200px;
}
</style>
