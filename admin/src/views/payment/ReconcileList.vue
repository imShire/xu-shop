<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getReconcileList, resolveReconcile } from '@/api/payment'
import { formatAmount, formatDate } from '@/utils/format'

const searchForm = ref({ date: '', channel: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getReconcileList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '账单日期', prop: 'bill_date', width: 120, formatter: (r) => formatDate(r.bill_date) },
  { label: '交易单号', prop: 'transaction_id', minWidth: 180 },
  { label: '订单号', prop: 'order_no', width: 160 },
  { label: '我方金额', slot: 'our_amount', width: 120, align: 'right' },
  { label: '平台金额', slot: 'wx_amount', width: 120, align: 'right' },
  { label: '差异类型', prop: 'diff_type', width: 120 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
]

async function handleResolve(row: any) {
  try {
    await ElMessageBox.confirm('确认标记该差异已处理？', '提示', { type: 'warning' })
    await resolveReconcile(row.id)
    ElMessage.success('已处理')
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
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
        <el-date-picker v-model="searchForm.date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 180px" />
        <el-select v-model="searchForm.channel" placeholder="渠道" clearable style="width: 120px">
          <el-option label="微信" value="wechat" />
          <el-option label="支付宝" value="alipay" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #our_amount="{ row }">{{ formatAmount(row.our_amount_cents) }}</template>
      <template #wx_amount="{ row }">
        <span style="color: #ef4444">{{ formatAmount(row.wx_amount_cents) }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 'resolved' ? 'success' : 'warning'" size="small">
          {{ row.status === 'resolved' ? '已处理' : '待处理' }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button
          v-if="row.status !== 'resolved'"
          text
          type="primary"
          size="small"
          @click="handleResolve(row)"
        >
          处理
        </el-button>
      </template>
    </ProTable>
  </div>
</template>
