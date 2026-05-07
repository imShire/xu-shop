<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getRefundList } from '@/api/payment'
import { formatAmount, formatTime } from '@/utils/format'

const searchForm = ref({ order_no: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getRefundList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '订单ID', prop: 'order_id', width: 180 },
  { label: '退款单号', prop: 'refund_no', width: 200 },
  { label: '退款金额', slot: 'amount', width: 110, align: 'right' },
  { label: '原因', prop: 'reason', minWidth: 150 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

const statusMap: Record<string, { label: string; type: string }> = {
  pending: { label: '处理中', type: 'warning' },
  success: { label: '退款成功', type: 'success' },
  failed: { label: '退款失败', type: 'danger' },
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

      <template #amount="{ row }">
        <span style="color: #ef4444">{{ formatAmount(row.amount_cents) }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="statusMap[row.status]?.type || ''" size="small">
          {{ statusMap[row.status]?.label || row.status }}
        </el-tag>
      </template>
    </ProTable>
  </div>
</template>
