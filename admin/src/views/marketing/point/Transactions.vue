<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { listPointTransactions, type PointTransaction } from '@/api/marketing'
import { formatTime } from '@/utils/format'

const searchForm = ref({ user_id: '', type: '', start: '', end: '' })
const dateRange = ref<[string, string] | null>(null)

const { list, total, page, pageSize, loading, fetch } = useTable<PointTransaction>((p) =>
  listPointTransactions({ ...p, ...searchForm.value })
)

const typeMap: Record<string, { label: string; type: 'success' | 'warning' | 'info' | 'danger' | 'primary' }> = {
  earn: { label: '入账', type: 'success' },
  spend: { label: '消费', type: 'warning' },
  expire: { label: '过期', type: 'info' },
  refund: { label: '退还', type: 'primary' },
  admin_adjust: { label: '人工调整', type: 'danger' },
  freeze: { label: '冻结', type: 'info' },
  unfreeze: { label: '解冻', type: 'success' },
}

const columns: ColumnDef[] = [
  { label: '流水 ID', prop: 'id', width: 130 },
  { label: '用户 ID', prop: 'user_id', width: 130 },
  { label: '类型', slot: 'type', width: 100, align: 'center' },
  { label: '变动', slot: 'change', width: 100, align: 'right' },
  { label: '余额', prop: 'balance_after', width: 100, align: 'right' },
  { label: '原因', prop: 'reason', minWidth: 200 },
  { label: '过期时间', prop: 'expire_at', width: 160, formatter: (r) => formatTime(r.expire_at) },
  { label: '时间', prop: 'created_at', width: 160, formatter: (r) => formatTime(r.created_at) },
]

function search() {
  if (dateRange.value) {
    searchForm.value.start = dateRange.value[0]
    searchForm.value.end = dateRange.value[1]
  } else {
    searchForm.value.start = ''
    searchForm.value.end = ''
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
        <el-input v-model="searchForm.user_id" placeholder="用户 ID" clearable style="width: 160px" />
        <el-select v-model="searchForm.type" placeholder="类型" clearable style="width: 130px">
          <el-option v-for="(v, k) in typeMap" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          start-placeholder="开始"
          end-placeholder="结束"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          style="width: 320px"
        />
        <el-button type="primary" @click="search">搜索</el-button>
      </template>

      <template #type="{ row }">
        <el-tag :type="typeMap[row.type]?.type || 'info'" size="small">{{ typeMap[row.type]?.label || row.type }}</el-tag>
      </template>
      <template #change="{ row }">
        <span :style="{ color: row.change >= 0 ? 'var(--el-color-success)' : 'var(--el-color-danger)' }">
          {{ row.change > 0 ? '+' : '' }}{{ row.change }}
        </span>
      </template>
    </ProTable>
  </div>
</template>
