<script setup lang="ts">
import { ref, onMounted } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
} from 'echarts/components'
import { getUserStats } from '@/api/stats'
import dayjs from 'dayjs'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent, TitleComponent])

const dateRange = ref<[string, string]>([
  dayjs().subtract(29, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
])

const loading = ref(false)
const stats = ref<any>({})
const trendOption = ref<any>({})

async function loadData() {
  if (!dateRange.value) return
  loading.value = true
  try {
    const [from, to] = dateRange.value
    const res = await getUserStats({ from, to })
    stats.value = res || {}
    trendOption.value = buildTrend((res as any)?.trend || [])
  } finally {
    loading.value = false
  }
}

function buildTrend(data: any[]) {
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: data.map((d: any) => d.date) },
    yAxis: { type: 'value', name: '新增用户数' },
    series: [
      {
        name: '新增用户',
        type: 'line',
        smooth: true,
        data: data.map((d: any) => d.count),
        itemStyle: { color: '#3b82f6' },
        areaStyle: { opacity: 0.1, color: '#3b82f6' },
      },
    ],
  }
}

onMounted(loadData)
</script>

<template>
  <div v-loading="loading" style="padding: 20px">
    <!-- 日期选择 -->
    <div class="page-card" style="margin-bottom: 16px; display: flex; align-items: center; gap: 12px">
      <span style="font-size: 14px; color: var(--text-secondary)">统计周期：</span>
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

    <!-- 核心指标 -->
    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="6">
        <div class="stats-card">
          <div class="stats-label">新增用户</div>
          <div class="stats-value">{{ stats.new_user_count || 0 }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card">
          <div class="stats-label">付费用户</div>
          <div class="stats-value" style="color: #f59e0b">{{ stats.paid_user_count || 0 }}</div>
        </div>
      </el-col>
    </el-row>

    <!-- 趋势图 -->
    <el-card>
      <template #header>新增用户趋势</template>
      <VChart :option="trendOption" style="height: 320px" autoresize />
    </el-card>
  </div>
</template>
