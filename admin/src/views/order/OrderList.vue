<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getOrderList, exportOrders, cancelOrder } from '@/api/order'
import { shipOrder } from '@/api/shipping'
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format'

const router = useRouter()

const searchForm = ref({
  order_no: '',
  phone: '',
  status: '',
  start_date: '',
  end_date: '',
})

const dateRange = ref<[string, string] | null>(null)

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getOrderList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '订单号', slot: 'order_no', width: 180 },
  { label: '用户', slot: 'user', width: 120 },
  { label: '商品', slot: 'items', minWidth: 160 },
  { label: '实付金额', slot: 'amount', width: 110, align: 'right' },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '下单时间', prop: 'created_at', width: 150, formatter: (row) => formatTime(row.created_at) },
]

const shipDialogVisible = ref(false)
const currentOrderId = ref('')
const shipForm = ref({ carrier_code: '', tracking_no: '' })
const shipLoading = ref(false)

function openShipDialog(orderId: string) {
  currentOrderId.value = orderId
  shipForm.value = { carrier_code: '', tracking_no: '' }
  shipDialogVisible.value = true
}

async function handleShip() {
  if (!shipForm.value.carrier_code || !shipForm.value.tracking_no) {
    return ElMessage.warning('请填写快递信息')
  }
  shipLoading.value = true
  try {
    await ElMessageBox.confirm('确认发货？发货后无法撤回', '发货确认', { type: 'warning' })
    await shipOrder(currentOrderId.value, shipForm.value)
    ElMessage.success('发货成功')
    shipDialogVisible.value = false
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '发货失败')
  } finally {
    shipLoading.value = false
  }
}

async function handleCancel(row: any) {
  try {
    await ElMessageBox.confirm('确认取消该订单？', '提示', { type: 'warning' })
    await cancelOrder(row.id, '')
    ElMessage.success('取消成功')
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '取消失败')
  }
}

async function handleExport() {
  const blob = await exportOrders(searchForm.value)
  const url = URL.createObjectURL(blob as unknown as Blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `orders_${Date.now()}.xlsx`
  a.click()
  URL.revokeObjectURL(url)
}

function handleSearch() {
  if (dateRange.value) {
    searchForm.value.start_date = dateRange.value[0]
    searchForm.value.end_date = dateRange.value[1]
  } else {
    searchForm.value.start_date = ''
    searchForm.value.end_date = ''
  }
  page.value = 1
  fetch(searchForm.value)
}

function handleReset() {
  searchForm.value = { order_no: '', phone: '', status: '', start_date: '', end_date: '' }
  dateRange.value = null
  handleSearch()
}

onMounted(() => fetch(searchForm.value))
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
      @refresh="fetch(searchForm)"
    >
      <template #search>
        <el-input v-model="searchForm.order_no" placeholder="订单号" clearable style="width: 180px" />
        <el-input v-model="searchForm.phone" placeholder="用户手机号" clearable style="width: 150px" />
        <el-select v-model="searchForm.status" placeholder="订单状态" clearable style="width: 120px">
          <el-option v-for="(v, k) in orderStatusMap" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          value-format="YYYY-MM-DD"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          style="width: 240px"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </template>

      <template #toolbar>
        <el-button @click="handleExport">导出订单</el-button>
      </template>

      <template #order_no="{ row }">
        <el-link type="primary" @click="router.push(`/order/detail/${row.id}`)">
          {{ row.order_no }}
        </el-link>
      </template>

      <template #user="{ row }">
        <div>
          <div>{{ row.address_snapshot?.name }}</div>
          <div style="font-size: 12px; color: var(--text-secondary)">{{ row.address_snapshot?.phone }}</div>
        </div>
      </template>

      <template #items="{ row }">
        <div v-if="row.items?.[0]" class="text-truncate">
          {{ row.items[0].product_snapshot?.title }}
          <span v-if="row.items.length > 1" style="color: var(--text-secondary)">
            等{{ row.items.length }}件
          </span>
        </div>
      </template>

      <template #amount="{ row }">
        <span style="font-weight: 600; color: #f59e0b">{{ formatAmount(row.pay_cents) }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="orderStatusMap[row.status]?.type || ''" size="small">
          {{ orderStatusMap[row.status]?.label || row.status }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="router.push(`/order/detail/${row.id}`)">
          详情
        </el-button>
        <el-button
          v-if="row.status === 'paid'"
          text
          type="success"
          size="small"
          @click="openShipDialog(row.id)"
        >
          发货
        </el-button>
        <el-button
          v-if="row.status === 'pending_payment' || row.status === 'paid'"
          text
          type="danger"
          size="small"
          @click="handleCancel(row)"
        >
          取消订单
        </el-button>
      </template>
    </ProTable>

    <!-- 发货对话框 -->
    <el-dialog v-model="shipDialogVisible" title="手动发货" width="400px">
      <el-form :model="shipForm" label-width="90px">
        <el-form-item label="快递公司">
          <el-input v-model="shipForm.carrier_code" placeholder="如：顺丰" />
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="shipForm.tracking_no" placeholder="请输入快递单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="shipLoading" @click="handleShip">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>
