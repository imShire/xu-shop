<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getOrderList } from '@/api/order'
import { shipOrder, batchShipOrders, getBatchShipStatus } from '@/api/shipping'
import { formatAmount, formatTime } from '@/utils/format'

const searchForm = ref({ order_no: '', phone: '' })
const selectedRows = ref<any[]>([])

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getOrderList({ ...params, status: 'paid', ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '订单号', prop: 'order_no', width: 180 },
  { label: '用户', slot: 'user', width: 120 },
  { label: '商品', slot: 'items', minWidth: 160 },
  { label: '实付', slot: 'amount', width: 100, align: 'right' },
  { label: '付款时间', prop: 'paid_at', width: 150, formatter: (r) => formatTime(r.paid_at) },
]

const singleShip = ref({ visible: false, orderId: '', carrier_code: '', tracking_no: '' })
const batchShip = ref({ visible: false, carrier_code: '', tracking_no: '' })
const shipLoading = ref(false)

// Progress dialog state
const progress = ref({
  visible: false,
  total: 0,
  done: 0,
  failed: 0,
  errors: [] as string[],
  pdf_url: '',
})
let pollTimer: ReturnType<typeof setInterval> | null = null

const progressPercent = computed(() => {
  if (!progress.value.total) return 0
  return Math.round(((progress.value.done + progress.value.failed) / progress.value.total) * 100)
})

const progressDone = computed(
  () => progress.value.done + progress.value.failed === progress.value.total && progress.value.total > 0
)

function clearPollTimer() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function startPolling(taskId: string) {
  clearPollTimer()
  pollTimer = setInterval(async () => {
    try {
      const res = await getBatchShipStatus(taskId)
      const data = res?.data ?? res
      progress.value.total = data.total ?? 0
      progress.value.done = data.done ?? 0
      progress.value.failed = data.failed ?? 0
      progress.value.errors = data.errors ?? []
      progress.value.pdf_url = data.pdf_url ?? ''

      if (progress.value.done + progress.value.failed >= progress.value.total && progress.value.total > 0) {
        clearPollTimer()
        fetch(searchForm.value)
      }
    } catch {
      // silently retry on poll error
    }
  }, 2000)
}

function closeProgressDialog() {
  clearPollTimer()
  progress.value.visible = false
}

function openSingleShip(row: any) {
  singleShip.value = { visible: true, orderId: row.id, carrier_code: '', tracking_no: '' }
}

async function handleSingleShip() {
  shipLoading.value = true
  try {
    await ElMessageBox.confirm('确认发货？发货后无法撤回', '发货确认', { type: 'warning' })
    await shipOrder(singleShip.value.orderId, {
      carrier_code: singleShip.value.carrier_code,
      tracking_no: singleShip.value.tracking_no,
    })
    ElMessage.success('发货成功')
    singleShip.value.visible = false
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    shipLoading.value = false
  }
}

function openBatchShip() {
  if (!selectedRows.value.length) return ElMessage.warning('请先选择订单')
  batchShip.value = { visible: true, carrier_code: '', tracking_no: '' }
}

async function handleBatchShip() {
  shipLoading.value = true
  try {
    await ElMessageBox.confirm(`确认批量发货 ${selectedRows.value.length} 个订单？`, '批量发货', { type: 'warning' })
    const res = await batchShipOrders(
      selectedRows.value.map((r) => ({
        order_id: r.id,
        carrier_code: batchShip.value.carrier_code,
        tracking_no: batchShip.value.tracking_no,
      }))
    )
    const taskId: string = res?.data?.task_id ?? res?.task_id
    batchShip.value.visible = false

    // Reset and open progress dialog
    progress.value = {
      visible: true,
      total: selectedRows.value.length,
      done: 0,
      failed: 0,
      errors: [],
      pdf_url: '',
    }
    startPolling(taskId)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    shipLoading.value = false
  }
}

onMounted(() => fetch(searchForm.value))
onUnmounted(() => clearPollTimer())

</script>

<template>
  <div class="page-card">
    <ProTable
      :columns="columns"
      :data="list"
      :total="total"
      :loading="loading"
      v-model:page="page"
      v-model:page-size="pageSize"
      selection
      @refresh="fetch(searchForm)"
      @selection-change="(rows) => (selectedRows = rows)"
    >
      <template #search>
        <el-input v-model="searchForm.order_no" placeholder="订单号" clearable style="width: 180px" />
        <el-input v-model="searchForm.phone" placeholder="手机号" clearable style="width: 150px" />
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button type="primary" @click="openBatchShip">批量发货</el-button>
      </template>

      <template #user="{ row }">
        <div>
          <div>{{ row.address_snapshot?.name }}</div>
          <div style="font-size: 12px; color: var(--text-secondary)">{{ row.address_snapshot?.phone }}</div>
        </div>
      </template>

      <template #items="{ row }">
        <div class="text-truncate">
          {{ row.items?.[0]?.product_snapshot?.title }}
          <span v-if="row.items?.length > 1" style="color: var(--text-secondary)">等{{ row.items.length }}件</span>
        </div>
      </template>

      <template #amount="{ row }">
        <span style="font-weight: 600; color: #f59e0b">{{ formatAmount(row.pay_cents) }}</span>
      </template>

      <template #actions="{ row }">
        <el-button type="primary" size="small" @click="openSingleShip(row)">发货</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="singleShip.visible" title="快速发货" width="420px">
      <el-form label-width="80px">
        <el-form-item label="快递公司">
          <el-input v-model="singleShip.carrier_code" placeholder="如：顺丰" />
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="singleShip.tracking_no" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="singleShip.visible = false">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="handleSingleShip">确认发货</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchShip.visible" :title="`批量发货（已选 ${selectedRows.length} 单）`" width="420px">
      <el-form label-width="80px">
        <el-form-item label="快递公司">
          <el-input v-model="batchShip.carrier_code" placeholder="统一快递公司" />
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="batchShip.tracking_no" placeholder="同一运单号或留空" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchShip.visible = false">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="handleBatchShip">确认批量发货</el-button>
      </template>
    </el-dialog>

    <!-- Batch ship progress dialog -->
    <el-dialog
      v-model="progress.visible"
      title="批量发货进度"
      width="480px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="progressDone"
      @closed="clearPollTimer"
    >
      <div style="padding: 8px 0">
        <el-progress
          :percentage="progressPercent"
          :status="progressDone ? (progress.failed > 0 ? 'warning' : 'success') : undefined"
          :striped="!progressDone"
          :striped-flow="!progressDone"
          :duration="10"
        />

        <div style="margin-top: 12px; text-align: center; color: var(--el-text-color-regular)">
          <template v-if="!progressDone">
            处理中 {{ progress.done + progress.failed }}/{{ progress.total }}...
          </template>
          <template v-else>
            <span style="color: var(--el-color-success)">成功 {{ progress.done }} 单</span>
            <span v-if="progress.failed > 0" style="margin-left: 12px; color: var(--el-color-warning)">
              失败 {{ progress.failed }} 单
            </span>
          </template>
        </div>

        <template v-if="progressDone && progress.errors.length > 0">
          <div style="margin-top: 16px; font-size: 13px; color: var(--el-text-color-secondary)">错误详情：</div>
          <el-scrollbar style="max-height: 200px; margin-top: 6px; border: 1px solid var(--el-border-color); border-radius: 4px; padding: 8px">
            <div
              v-for="(err, idx) in progress.errors"
              :key="idx"
              style="font-size: 12px; line-height: 1.8; color: var(--el-color-danger)"
            >
              {{ err }}
            </div>
          </el-scrollbar>
        </template>

        <div v-if="progressDone && progress.pdf_url" style="margin-top: 16px; text-align: center">
          <el-button type="primary" link :href="progress.pdf_url" target="_blank" tag="a">
            下载面单
          </el-button>
        </div>
      </div>

      <template #footer>
        <el-button :disabled="!progressDone" @click="closeProgressDialog">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>
