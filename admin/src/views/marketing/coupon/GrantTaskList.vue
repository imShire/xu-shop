<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { listCouponGrantTasks, type CouponGrantTask } from '@/api/marketing'
import { formatTime } from '@/utils/format'

const router = useRouter()
const searchForm = ref({ status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable<CouponGrantTask>((p) =>
  listCouponGrantTasks({ ...p, ...searchForm.value })
)

const statusTag: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'danger' | 'primary' }> = {
  pending: { label: '排队中', type: 'info' },
  running: { label: '执行中', type: 'primary' },
  done: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'danger' },
}

const columns: ColumnDef[] = [
  { label: '任务 ID', prop: 'id', width: 130 },
  { label: '券模板', prop: 'template_name', minWidth: 180 },
  { label: '总数 / 成功 / 失败', slot: 'progress', width: 200 },
  { label: '状态', slot: 'status', width: 100, align: 'center' },
  { label: '失败明细', slot: 'fail', width: 120 },
  { label: '创建时间', prop: 'created_at', width: 160, formatter: (r) => formatTime(r.created_at) },
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
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="排队中" value="pending" />
          <el-option label="执行中" value="running" />
          <el-option label="已完成" value="done" />
          <el-option label="失败" value="failed" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button type="primary" @click="router.push('/marketing/coupon/grant/create')">新建发放任务</el-button>
      </template>

      <template #progress="{ row }">
        {{ row.total }} / <span style="color: var(--el-color-success)">{{ row.succeeded }}</span> /
        <span style="color: var(--el-color-danger)">{{ row.failed }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="statusTag[row.status]?.type" size="small">{{ statusTag[row.status]?.label || row.status }}</el-tag>
      </template>

      <template #fail="{ row }">
        <el-link v-if="row.fail_csv_url" :href="row.fail_csv_url" target="_blank" type="primary">下载</el-link>
        <span v-else>-</span>
      </template>
    </ProTable>
  </div>
</template>
