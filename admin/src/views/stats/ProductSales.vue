<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProductRanking, exportStats } from '@/api/stats'
import { formatAmount } from '@/utils/format'
import dayjs from 'dayjs'

const dateRange = ref<[string, string]>([
  dayjs().subtract(29, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
])

const loading = ref(false)
const list = ref<any[]>([])

async function loadData() {
  if (!dateRange.value) return
  loading.value = true
  try {
    const [from, to] = dateRange.value
    const res = await getProductRanking({ from, to, page_size: 50 })
    list.value = (res as any)?.list || []
  } finally {
    loading.value = false
  }
}

async function handleExport() {
  const [from, to] = dateRange.value || []
  const blob = await exportStats({ type: 'products', from, to })
  const url = URL.createObjectURL(blob as unknown as Blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `product_sales_${Date.now()}.xlsx`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(loadData)
</script>

<template>
  <div style="padding: 20px">
    <div class="page-card" style="margin-bottom: 16px; display: flex; align-items: center; gap: 12px">
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        value-format="YYYY-MM-DD"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        style="width: 280px"
      />
      <el-button type="primary" @click="loadData">查询</el-button>
      <el-button @click="handleExport">导出</el-button>
    </div>

    <div class="page-card">
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column type="index" label="排名" width="60" align="center" />
        <el-table-column prop="product_name" label="商品名称" min-width="180" />
        <el-table-column label="销售数量" prop="qty" width="100" align="right" />
        <el-table-column label="销售金额" width="120" align="right">
          <template #default="{ row }">
            <span style="color: #f59e0b">{{ formatAmount(row.amount_cents) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>
