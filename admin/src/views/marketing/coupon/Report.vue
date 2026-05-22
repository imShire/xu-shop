<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { listUserCoupons } from '@/api/marketing'
import { formatTime, formatAmount } from '@/utils/format'

const searchForm = ref({ user_id: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((p) =>
  listUserCoupons({ ...p, ...searchForm.value })
)

const statusMap: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'danger' }> = {
  unused: { label: '未使用', type: 'info' },
  locked: { label: '锁定', type: 'warning' },
  used: { label: '已使用', type: 'success' },
  expired: { label: '已过期', type: 'danger' },
}

const columns: ColumnDef[] = [
  { label: '券名称', prop: 'name', minWidth: 180 },
  { label: '面额', slot: 'value', width: 100 },
  { label: '门槛', slot: 'min', width: 110 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '领取时间', prop: 'claimed_at', width: 160, formatter: (r) => formatTime(r.claimed_at) },
  { label: '过期时间', prop: 'expire_at', width: 160, formatter: (r) => formatTime(r.expire_at) },
  { label: '使用时间', prop: 'used_at', width: 160, formatter: (r) => formatTime(r.used_at) },
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
        <el-input v-model="searchForm.user_id" placeholder="用户 ID" clearable style="width: 180px" />
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="未使用" value="unused" />
          <el-option label="已使用" value="used" />
          <el-option label="已过期" value="expired" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #value="{ row }">{{ formatAmount(row.value_cents || 0) }}</template>
      <template #min="{ row }">
        <span v-if="row.min_amount_cents">满 {{ formatAmount(row.min_amount_cents) }}</span>
        <span v-else>无</span>
      </template>
      <template #status="{ row }">
        <el-tag :type="statusMap[row.status]?.type" size="small">
          {{ statusMap[row.status]?.label || row.status }}
        </el-tag>
      </template>
    </ProTable>
  </div>
</template>
