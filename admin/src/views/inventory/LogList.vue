<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getInventoryLogs } from '@/api/inventory'
import { formatTime } from '@/utils/format'

const searchForm = ref({
  sku_id: '',
  type: '',
  start_date: '',
  end_date: '',
})

const dateRange = ref<[string, string] | null>(null)

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getInventoryLogs({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: 'SKU ID', prop: 'sku_id', width: 120 },
  { label: '操作类型', slot: 'type', width: 100, align: 'center' },
  { label: '变动数量', slot: 'change', width: 100, align: 'center' },
  { label: '变动前', prop: 'balance_before', width: 90, align: 'center' },
  { label: '变动后', prop: 'balance_after', width: 90, align: 'center' },
  { label: '备注', prop: 'reason', minWidth: 120 },
  { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

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
        <el-input v-model="searchForm.sku_id" placeholder="SKU ID" clearable style="width: 150px" />
        <el-select v-model="searchForm.type" placeholder="操作类型" clearable style="width: 130px">
          <el-option label="入库" value="in" />
          <el-option label="出库" value="out" />
          <el-option label="手动调整" value="adjust" />
          <el-option label="订单出库" value="order" />
          <el-option label="退款入库" value="refund" />
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
      </template>

      <template #type="{ row }">
        <el-tag size="small">{{ row.type }}</el-tag>
      </template>

      <template #change="{ row }">
        <span :style="{ color: row.change > 0 ? '#22c55e' : '#ef4444', fontWeight: 600 }">
          {{ row.change > 0 ? '+' : '' }}{{ row.change }}
        </span>
      </template>
    </ProTable>
  </div>
</template>
