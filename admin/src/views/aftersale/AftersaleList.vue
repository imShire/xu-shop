<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getAftersaleList, approveAftersale, rejectAftersale } from '@/api/aftersale'
import { createRefund } from '@/api/payment'
import { formatAmount, formatTime } from '@/utils/format'
import PriceInput from '@/components/PriceInput/index.vue'

const searchForm = ref({ order_no: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getAftersaleList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '订单号', prop: 'order_no', width: 180 },
  { label: '用户', prop: 'user_nickname', width: 100 },
  { label: '类型', slot: 'type', width: 90, align: 'center' },
  { label: '原因', prop: 'reason', minWidth: 150 },
  { label: '退款金额', slot: 'amount', width: 110, align: 'right' },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '申请时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
  { label: '操作', slot: 'actions', width: 180 },
]

const rejectDialog = ref({ visible: false, id: '', reason: '' })
const rejectLoading = ref(false)
const refundDialog = ref({ visible: false, id: '', amount: 0 })
const refundLoading = ref(false)

const typeMap: Record<string, string> = {
  cancel: '取消订单',
  refund: '仅退款',
  return: '退货退款',
}

const statusMap: Record<string, { label: string; type: string }> = {
  pending: { label: '待审核', type: 'warning' },
  approved: { label: '已同意', type: 'success' },
  rejected: { label: '已拒绝', type: 'danger' },
  refunding: { label: '退款中', type: '' },
  refunded: { label: '已退款', type: 'success' },
}

async function handleApprove(row: any) {
  try {
    await ElMessageBox.confirm('确认同意退款申请？退款将自动发起', '提示', { type: 'warning' })
    await approveAftersale(row.order_id)
    ElMessage.success('已同意')
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  }
}

function openReject(row: any) {
  rejectDialog.value = { visible: true, id: row.order_id, reason: '' }
}

async function handleReject() {
  rejectLoading.value = true
  try {
    await ElMessageBox.confirm('确认拒绝该售后申请？', '提示', { type: 'warning' })
    await rejectAftersale(rejectDialog.value.id, rejectDialog.value.reason)
    ElMessage.success('已拒绝')
    rejectDialog.value.visible = false
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    rejectLoading.value = false
  }
}

function openRefund(row: any) {
  refundDialog.value = { visible: true, id: row.order_id, amount: row.pay_cents }
}

async function handleRefund() {
  refundLoading.value = true
  try {
    await ElMessageBox.confirm('确认直接退款？此操作不可撤销', '退款确认', { type: 'warning' })
    await createRefund(refundDialog.value.id, { amount_cents: refundDialog.value.amount, reason: '' })
    ElMessage.success('退款成功')
    refundDialog.value.visible = false
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    refundLoading.value = false
  }
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
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option v-for="(v, k) in statusMap" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #type="{ row }">
        <el-tag size="small">{{ typeMap[row.type] || row.type }}</el-tag>
      </template>

      <template #amount="{ row }">
        <span style="color: #f59e0b">{{ formatAmount(row.pay_cents) }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="statusMap[row.status]?.type || ''" size="small">
          {{ statusMap[row.status]?.label || row.status }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <template v-if="row.status === 'pending'">
          <el-button text type="success" size="small" @click="handleApprove(row)">同意</el-button>
          <el-button text type="danger" size="small" @click="openReject(row)">拒绝</el-button>
        </template>
        <el-button
          v-if="['approved', 'pending'].includes(row.status)"
          text
          type="warning"
          size="small"
          @click="openRefund(row)"
        >
          直接退款
        </el-button>
      </template>
    </ProTable>

    <el-dialog v-model="rejectDialog.visible" title="拒绝原因" width="400px">
      <el-input v-model="rejectDialog.reason" type="textarea" placeholder="请输入拒绝原因" :rows="3" />
      <template #footer>
        <el-button @click="rejectDialog.visible = false">取消</el-button>
        <el-button type="danger" :loading="rejectLoading" @click="handleReject">确认拒绝</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="refundDialog.visible" title="直接退款" width="400px">
      <el-form label-width="90px">
        <el-form-item label="退款金额">
          <PriceInput v-model="refundDialog.amount" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="refundDialog.visible = false">取消</el-button>
        <el-button type="warning" :loading="refundLoading" @click="handleRefund">确认退款</el-button>
      </template>
    </el-dialog>
  </div>
</template>
