<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getShipmentList } from '@/api/shipping'
import { formatTime } from '@/utils/format'

const searchForm = ref({ order_id: '', tracking_no: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getShipmentList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '订单 ID', prop: 'order_id', width: 180 },
  { label: '快递公司', prop: 'carrier_code', width: 120 },
  { label: '运单号', prop: 'tracking_no', minWidth: 160 },
  { label: '物流状态', slot: 'status', width: 100, align: 'center' },
  { label: '发货时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
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
        <el-input v-model="searchForm.order_id" placeholder="订单 ID" clearable style="width: 180px" />
        <el-input v-model="searchForm.tracking_no" placeholder="运单号" clearable style="width: 180px" />
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #status="{ row }">
        <el-tag size="small">{{ row.status }}</el-tag>
      </template>
    </ProTable>
  </div>
</template>
