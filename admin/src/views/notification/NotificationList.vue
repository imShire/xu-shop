<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getNotificationList } from '@/api/notification'
import { formatTime } from '@/utils/format'

const searchForm = ref({ status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getNotificationList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '模板', prop: 'template_code', width: 150 },
  { label: '目标', prop: 'target', minWidth: 150 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '发送时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

const statusMap: Record<string, { label: string; type: string }> = {
  pending: { label: '待发送', type: 'warning' },
  success: { label: '发送成功', type: 'success' },
  failed: { label: '发送失败', type: 'danger' },
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
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option v-for="(v, k) in statusMap" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #status="{ row }">
        <el-tag :type="statusMap[row.status]?.type || ''" size="small">
          {{ statusMap[row.status]?.label || row.status }}
        </el-tag>
      </template>
    </ProTable>
  </div>
</template>
