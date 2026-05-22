<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import {
  listPointAdjustTickets,
  createPointAdjustTicket,
  approvePointAdjustTicket,
  rejectPointAdjustTicket,
  type PointAdjustTicket,
} from '@/api/marketing'
import { formatTime } from '@/utils/format'

const searchForm = ref({})
const dialogVisible = ref(false)
const newTicket = ref({ user_id: '', change: 0, reason: '' })

const { list, total, page, pageSize, loading, fetch } = useTable<PointAdjustTicket>((p) =>
  listPointAdjustTickets({ ...p, ...searchForm.value })
)

const statusMap: Record<string, { label: string; type: 'info' | 'success' | 'danger' | 'warning' }> = {
  pending: { label: '待审批', type: 'warning' },
  approved: { label: '已通过', type: 'success' },
  rejected: { label: '已驳回', type: 'danger' },
}

const columns: ColumnDef[] = [
  { label: '工单 ID', prop: 'id', width: 130 },
  { label: '用户 ID', prop: 'user_id', width: 130 },
  { label: '调整', slot: 'change', width: 100, align: 'right' },
  { label: '原因', prop: 'reason', minWidth: 200 },
  { label: '申请人', prop: 'applicant_id', width: 130 },
  { label: '审批人', prop: 'approver_id', width: 130 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '时间', prop: 'created_at', width: 160, formatter: (r) => formatTime(r.created_at) },
]

async function submit() {
  if (!newTicket.value.user_id || !newTicket.value.reason) {
    ElMessage.warning('请填写完整')
    return
  }
  await createPointAdjustTicket({ ...newTicket.value })
  ElMessage.success('已提交，等待审批')
  dialogVisible.value = false
  newTicket.value = { user_id: '', change: 0, reason: '' }
  fetch(searchForm.value)
}

async function approve(row: PointAdjustTicket) {
  await ElMessageBox.confirm(`确认审批通过？该工单将立即影响用户积分余额`, '审批通过', { type: 'warning' })
  await approvePointAdjustTicket(row.id)
  ElMessage.success('已通过')
  fetch(searchForm.value)
}

async function reject(row: PointAdjustTicket) {
  await ElMessageBox.confirm(`确认驳回此工单？`, '审批驳回', { type: 'warning' })
  await rejectPointAdjustTicket(row.id)
  ElMessage.success('已驳回')
  fetch(searchForm.value)
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
      <template #toolbar>
        <el-button type="primary" @click="dialogVisible = true">新建调整工单</el-button>
      </template>

      <template #change="{ row }">
        <span :style="{ color: row.change >= 0 ? 'var(--el-color-success)' : 'var(--el-color-danger)' }">
          {{ row.change > 0 ? '+' : '' }}{{ row.change }}
        </span>
      </template>
      <template #status="{ row }">
        <el-tag :type="statusMap[row.status]?.type" size="small">{{ statusMap[row.status]?.label }}</el-tag>
      </template>

      <template #actions="{ row }">
        <template v-if="row.status === 'pending'">
          <el-button text type="success" size="small" @click="approve(row)">通过</el-button>
          <el-button text type="danger" size="small" @click="reject(row)">驳回</el-button>
        </template>
      </template>
    </ProTable>

    <el-dialog v-model="dialogVisible" title="新建积分调整工单" width="500px">
      <el-form label-width="100px">
        <el-form-item label="用户 ID" required>
          <el-input v-model="newTicket.user_id" />
        </el-form-item>
        <el-form-item label="调整数量" required>
          <el-input-number v-model="newTicket.change" :step="1" />
          <span style="margin-left: 8px; color: var(--el-text-color-secondary)">正数加分，负数扣分</span>
        </el-form-item>
        <el-form-item label="原因" required>
          <el-input v-model="newTicket.reason" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submit">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>
