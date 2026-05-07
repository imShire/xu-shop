<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getAuditLogs } from '@/api/system'
import { formatTime } from '@/utils/format'

const searchForm = ref({
  module: '',
  operator: '',
  start_date: '',
  end_date: '',
})
const dateRange = ref<[string, string] | null>(null)

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getAuditLogs({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '模块', prop: 'module', width: 100 },
  { label: '操作', prop: 'action', width: 150 },
  { label: '操作人', prop: 'operator', width: 100 },
  { label: 'IP', prop: 'ip', width: 130 },
  { label: '详情', prop: 'detail', minWidth: 200 },
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
        <el-input v-model="searchForm.module" placeholder="模块" clearable style="width: 120px" />
        <el-input v-model="searchForm.operator" placeholder="操作人" clearable style="width: 120px" />
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          value-format="YYYY-MM-DD"
          start-placeholder="开始"
          end-placeholder="结束"
          style="width: 240px"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
      </template>
    </ProTable>
  </div>
</template>
