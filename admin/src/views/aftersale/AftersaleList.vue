<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getAftersaleList } from '@/api/aftersale'
import { formatAmount, formatTime } from '@/utils/format'
import {
  AFTERSALE_STATUS_LABEL,
  AFTERSALE_TYPE_LABEL,
  type AftersaleStatus,
  type AftersaleType,
} from '@/types/aftersale'

const router = useRouter()

const searchForm = ref<{
  keyword: string
  status: AftersaleStatus | ''
  type: AftersaleType | ''
  applied_range: [string, string] | []
}>({
  keyword: '',
  status: '',
  type: '',
  applied_range: [],
})

function buildParams() {
  const [from, to] = searchForm.value.applied_range || []
  return {
    keyword: searchForm.value.keyword || undefined,
    status: searchForm.value.status || undefined,
    type: searchForm.value.type || undefined,
    applied_from: from || undefined,
    applied_to: to || undefined,
  }
}

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getAftersaleList({ ...params, ...buildParams() }),
)

const columns: ColumnDef[] = [
  { label: '售后单号', prop: 'aftersale_no', width: 200 },
  { label: '订单号', prop: 'order_no', width: 200 },
  { label: '用户', slot: 'user', width: 140 },
  { label: '类型', slot: 'type', width: 110, align: 'center' },
  { label: '状态', slot: 'status', width: 130, align: 'center' },
  { label: '退款金额', slot: 'amount', width: 120, align: 'right' },
  {
    label: '申请时间',
    prop: 'applied_at',
    width: 170,
    formatter: (r) => formatTime(r.applied_at),
  },
  { label: '操作', slot: 'actions', width: 100, fixed: 'right' },
]

function handleSearch() {
  page.value = 1
  fetch()
}

function handleReset() {
  searchForm.value = { keyword: '', status: '', type: '', applied_range: [] }
  page.value = 1
  fetch()
}

function goDetail(id: string) {
  router.push(`/aftersale/detail/${id}`)
}

onMounted(() => fetch())
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
      @refresh="fetch()"
    >
      <template #search>
        <el-input
          v-model="searchForm.keyword"
          placeholder="售后单号 / 订单号"
          clearable
          style="width: 220px"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="searchForm.status"
          placeholder="状态"
          clearable
          style="width: 150px"
        >
          <el-option
            v-for="(v, k) in AFTERSALE_STATUS_LABEL"
            :key="k"
            :label="v.label"
            :value="k"
          />
        </el-select>
        <el-select
          v-model="searchForm.type"
          placeholder="类型"
          clearable
          style="width: 130px"
        >
          <el-option
            v-for="(label, k) in AFTERSALE_TYPE_LABEL"
            :key="k"
            :label="label"
            :value="k"
          />
        </el-select>
        <el-date-picker
          v-model="searchForm.applied_range"
          type="datetimerange"
          range-separator="至"
          start-placeholder="申请起"
          end-placeholder="申请止"
          value-format="YYYY-MM-DDTHH:mm:ss[Z]"
          style="width: 360px"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </template>

      <template #user="{ row }">
        <span>{{ row.user_nickname || row.user_id }}</span>
      </template>

      <template #type="{ row }">
        <el-tag size="small">
          {{ AFTERSALE_TYPE_LABEL[row.type as keyof typeof AFTERSALE_TYPE_LABEL] || row.type }}
        </el-tag>
      </template>

      <template #status="{ row }">
        <el-tag
          :type="AFTERSALE_STATUS_LABEL[row.status as keyof typeof AFTERSALE_STATUS_LABEL]?.type || ''"
          size="small"
        >
          {{ AFTERSALE_STATUS_LABEL[row.status as keyof typeof AFTERSALE_STATUS_LABEL]?.label || row.status }}
        </el-tag>
      </template>

      <template #amount="{ row }">
        <span style="color: #f59e0b">{{ formatAmount(row.refund_amount_cents) }}</span>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="goDetail(row.id)">详情</el-button>
      </template>
    </ProTable>
  </div>
</template>
