<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getChannelStats } from '@/api/stats'
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
    list.value = await getChannelStats({ from, to }) || []
  } finally {
    loading.value = false
  }
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
    </div>

    <div class="page-card">
      <el-table v-loading="loading" :data="list" border stripe>
        <el-table-column prop="channel_code_name" label="渠道名称" min-width="150" />
        <el-table-column prop="scan_count" label="扫码数" width="100" align="right" />
        <el-table-column prop="add_count" label="加粉数" width="100" align="right" />
        <el-table-column prop="order_count" label="下单数" width="100" align="right" />
        <el-table-column label="销售额" width="120" align="right">
          <template #default="{ row }">
            <span style="color: #f59e0b">{{ formatAmount(row.amount_cents || 0) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="转化率" width="100" align="right">
          <template #default="{ row }">
            {{ row.scan_count ? ((row.order_count / row.scan_count) * 100).toFixed(1) : '0' }}%
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>
