<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getPaymentList } from '@/api/payment'
import { formatAmount, formatTime } from '@/utils/format'

const searchForm = ref({ order_no: '', channel: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getPaymentList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '订单ID', prop: 'order_id', width: 180 },
  { label: '交易单号', prop: 'transaction_id', minWidth: 200 },
  { label: '渠道', prop: 'channel', width: 100, align: 'center' },
  { label: '金额', slot: 'amount', width: 110, align: 'right' },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

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
        <el-select v-model="searchForm.channel" placeholder="支付渠道" clearable style="width: 120px">
          <el-option label="微信" value="wechat" />
          <el-option label="支付宝" value="alipay" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #amount="{ row }">
        <span style="font-weight: 600; color: #f59e0b">{{ formatAmount(row.amount_cents) }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 'success' ? 'success' : 'info'" size="small">
          {{ row.status === 'success' ? '成功' : row.status }}
        </el-tag>
      </template>
    </ProTable>
  </div>
</template>
