<script setup lang="ts">
import { ref, watch } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart, BarChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
} from 'echarts/components'
import { getStatsOverview, getStatsTrend, getCategoryStats } from '@/api/stats'
import { formatAmount } from '@/utils/format'
import dayjs from 'dayjs'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent, TitleComponent])

const dateRange = ref<[string, string]>([
  dayjs().subtract(6, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
])

const loading = ref(false)
const overview = ref<any>({})
const trendOption = ref<any>({})
const categoryOption = ref<any>({})

async function loadData() {
  if (!dateRange.value) return
  loading.value = true
  try {
    const [from, to] = dateRange.value
    const [ov, trend, cats] = await Promise.all([
      getStatsOverview({ from, to }),
      getStatsTrend({ from, to }),
      getCategoryStats({ from, to }),
    ])
    overview.value = ov
    trendOption.value = buildTrend(trend || [])
    categoryOption.value = buildCategory(cats || [])
  } finally {
    loading.value = false
  }
}

function buildTrend(data: any[]) {
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['订单数', '销售额(元)'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: data.map((d) => d.date) },
    yAxis: [{ type: 'value', name: '订单数' }, { type: 'value', name: '销售额' }],
    series: [
      {
        name: '订单数',
        type: 'line',
        smooth: true,
        data: data.map((d) => d.order_count),
        itemStyle: { color: '#f59e0b' },
        areaStyle: { opacity: 0.1, color: '#f59e0b' },
      },
      {
        name: '销售额(元)',
        type: 'line',
        yAxisIndex: 1,
        smooth: true,
        data: data.map((d) => (d.amount_cents / 100).toFixed(2)),
        itemStyle: { color: '#3b82f6' },
      },
    ],
  }
}

function buildCategory(data: any[]) {
  return {
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: data.map((d) => d.category_name) },
    yAxis: { type: 'value' },
    series: [
      {
        name: '销售额',
        type: 'bar',
        data: data.map((d) => (d.amount_cents / 100).toFixed(2)),
        itemStyle: { color: '#f59e0b', borderRadius: [4, 4, 0, 0] },
      },
    ],
  }
}

watch(dateRange, loadData, { immediate: true })
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
          <div class="stats-label">订单总数</div>
          <div class="stats-value">{{ overview.paid_order_count || 0 }}</div>
          <div class="stats-sub">
            <span :class="(overview.order_count_yoy || 0) >= 0 ? 'up' : 'down'">
              {{ (overview.order_count_yoy || 0) >= 0 ? '↑' : '↓' }}
              {{ Math.abs((overview.order_count_yoy || 0) * 100).toFixed(1) }}%
            </span>
            同期对比
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card">
          <div class="stats-label">销售额</div>
          <div class="stats-value">{{ formatAmount(overview.paid_amount_cents || 0) }}</div>
          <div class="stats-sub">
            <span :class="(overview.amount_yoy || 0) >= 0 ? 'up' : 'down'">
              {{ (overview.amount_yoy || 0) >= 0 ? '↑' : '↓' }}
              {{ Math.abs((overview.amount_yoy || 0) * 100).toFixed(1) }}%
            </span>
            同期对比
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card">
          <div class="stats-label">退款金额</div>
          <div class="stats-value" style="color: #ef4444">
            {{ formatAmount(overview.refund_amount_cents || 0) }}
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card">
          <div class="stats-label">净收入</div>
          <div class="stats-value" style="color: #22c55e">
            {{ formatAmount(overview.net_amount_cents || 0) }}
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="16">
        <el-card>
          <template #header>销售趋势</template>
          <VChart :option="trendOption" style="height: 300px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>分类销售占比</template>
          <VChart :option="categoryOption" style="height: 300px" autoresize />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.stats-value {
  font-size: 22px;
}
</style>
