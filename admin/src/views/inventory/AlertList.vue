<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getInventoryAlerts, markAlertRead, markAllAlertsRead } from '@/api/inventory'
import { formatTime } from '@/utils/format'

const searchForm = ref({ status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getInventoryAlerts({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: 'SKU ID', prop: 'sku_id', minWidth: 120 },
  { label: '当前库存', prop: 'current_stock', width: 100, align: 'center' },
  { label: '预警阈值', prop: 'threshold_at_alert', width: 100, align: 'center' },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

async function handleMarkRead(row: any) {
  try {
    await markAlertRead(row.id)
    ElMessage.success('已标记已读')
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

async function handleMarkAllRead() {
  try {
    await markAllAlertsRead()
    ElMessage.success('全部已读')
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
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
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="未读" value="unread" />
          <el-option label="已读" value="read" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button @click="handleMarkAllRead">全部标记已读</el-button>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 'unread' ? 'danger' : 'info'" size="small">
          {{ row.status === 'unread' ? '未读' : '已读' }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button v-if="row.status === 'unread'" text type="primary" size="small" @click="handleMarkRead(row)">
          标记已读
        </el-button>
      </template>
    </ProTable>
  </div>
</template>
