<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getUserList } from '@/api/user'
import { formatTime, formatAmount } from '@/utils/format'

const searchForm = ref({ keyword: '', level: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((p) =>
  getUserList({ ...p, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '用户 ID', prop: 'id', width: 130 },
  { label: '昵称', prop: 'nickname', minWidth: 140 },
  { label: '手机号', prop: 'phone', width: 130 },
  { label: '等级', slot: 'level', width: 100 },
  { label: '累计支付', slot: 'amount', width: 130 },
  { label: '当前积分', prop: 'point_balance', width: 100, align: 'right' },
  { label: '升级时间', prop: 'level_updated_at', width: 160, formatter: (r) => formatTime(r.level_updated_at) },
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
        <el-input v-model="searchForm.keyword" placeholder="昵称 / 手机号 / ID" clearable style="width: 200px" />
        <el-select v-model="searchForm.level" placeholder="等级" clearable style="width: 120px">
          <el-option label="V0 普通" value="V0" />
          <el-option label="V1 银卡" value="V1" />
          <el-option label="V2 金卡" value="V2" />
          <el-option label="V3 钻石" value="V3" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #level="{ row }">
        <el-tag size="small">{{ row.level || 'V0' }}</el-tag>
      </template>
      <template #amount="{ row }">{{ formatAmount(row.total_paid_cents || 0) }}</template>
    </ProTable>
  </div>
</template>
