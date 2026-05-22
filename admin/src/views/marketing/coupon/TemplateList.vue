<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import {
  listCouponTemplates,
  onlineCouponTemplate,
  offlineCouponTemplate,
  type CouponTemplate,
} from '@/api/marketing'
import { formatTime, formatAmount } from '@/utils/format'

const router = useRouter()
const searchForm = ref({ keyword: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable<CouponTemplate>((p) =>
  listCouponTemplates({ ...p, ...searchForm.value })
)

const typeMap: Record<string, string> = {
  amount: '满减',
  discount: '折扣',
  no_threshold: '无门槛',
  exchange: '兑换',
}
const statusTag: Record<string, { label: string; type: 'info' | 'success' | 'warning' }> = {
  draft: { label: '草稿', type: 'info' },
  online: { label: '上架', type: 'success' },
  offline: { label: '下架', type: 'warning' },
}

const columns: ColumnDef[] = [
  { label: '模板名称', prop: 'name', minWidth: 180 },
  { label: '类型', slot: 'type', width: 90 },
  { label: '面额/折扣', slot: 'value', width: 130 },
  { label: '门槛', slot: 'threshold', width: 110 },
  { label: '总量/已用', slot: 'quota', width: 130 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '领取期', slot: 'claim', width: 220 },
]

async function handleOnline(row: CouponTemplate) {
  await ElMessageBox.confirm(`确认上架券模板「${row.name}」？上架后部分字段将不可改`, '上架', { type: 'warning' })
  await onlineCouponTemplate(row.id)
  ElMessage.success('已上架')
  fetch(searchForm.value)
}

async function handleOffline(row: CouponTemplate) {
  await ElMessageBox.confirm(`确认下架「${row.name}」？已领取的券不受影响`, '下架', { type: 'warning' })
  await offlineCouponTemplate(row.id)
  ElMessage.success('已下架')
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
        <el-input v-model="searchForm.keyword" placeholder="名称" clearable style="width: 180px" />
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="草稿" value="draft" />
          <el-option label="上架" value="online" />
          <el-option label="下架" value="offline" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button type="primary" @click="router.push('/marketing/coupon/template/create')">新建模板</el-button>
      </template>

      <template #type="{ row }">
        <el-tag size="small">{{ typeMap[row.type] || row.type }}</el-tag>
      </template>

      <template #value="{ row }">
        <span v-if="row.type === 'discount'">{{ row.discount_rate }} 折</span>
        <span v-else>{{ formatAmount(row.value_cents) }}</span>
      </template>

      <template #threshold="{ row }">
        <span v-if="row.min_amount_cents">满 {{ formatAmount(row.min_amount_cents) }}</span>
        <span v-else>无</span>
      </template>

      <template #quota="{ row }">
        {{ row.used_count }} / {{ row.total_quota || '∞' }}
      </template>

      <template #status="{ row }">
        <el-tag :type="statusTag[row.status]?.type" size="small">
          {{ statusTag[row.status]?.label || row.status }}
        </el-tag>
      </template>

      <template #claim="{ row }">
        <span style="font-size: 12px">
          {{ formatTime(row.claim_start_at) }} ~ {{ formatTime(row.claim_end_at) }}
        </span>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="router.push(`/marketing/coupon/template/edit/${row.id}`)">
          {{ row.status === 'draft' ? '编辑' : '查看' }}
        </el-button>
        <el-button v-if="row.status === 'draft' || row.status === 'offline'" text type="success" size="small" @click="handleOnline(row)">
          上架
        </el-button>
        <el-button v-if="row.status === 'online'" text type="warning" size="small" @click="handleOffline(row)">
          下架
        </el-button>
      </template>
    </ProTable>
  </div>
</template>
