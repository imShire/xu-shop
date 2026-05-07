<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from 'echarts/components'
import { getWorkbenchStats, getStatsTrend } from '@/api/stats'
import { getInventoryAlerts } from '@/api/inventory'
import { getOrderList } from '@/api/order'
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format'
import dayjs from 'dayjs'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

const router = useRouter()

const statsLoading = ref(false)
const stats = ref({
  today_order_count: 0,
  today_sales: 0,
  pending_ship: 0,
  aftersale_pending: 0,
})

const alertCount = ref(0)
const recentOrders = ref<any[]>([])
const trendOption = ref<any>({})

async function loadData() {
  statsLoading.value = true
  try {
    const [overview, trend, alerts, orders] = await Promise.all([
      getWorkbenchStats().catch(() => null),
      getStatsTrend({
        from: dayjs().subtract(6, 'day').format('YYYY-MM-DD'),
        to: dayjs().format('YYYY-MM-DD'),
      }).catch(() => []),
      getInventoryAlerts({ page: 1, page_size: 1, status: 'unread' }).catch(() => ({ total: 0 })),
      getOrderList({ page: 1, page_size: 10 }).catch(() => ({ list: [] })),
    ])

    if (overview) {
      stats.value = overview
    }

    alertCount.value = alerts?.total || 0
    recentOrders.value = orders?.list || []

    const trendData = trend || []
    trendOption.value = buildTrendChart(trendData)
  } finally {
    statsLoading.value = false
  }
}

function buildTrendChart(data: any[]) {
  return {
    tooltip: { trigger: 'axis' },
    legend: { data: ['订单数', '销售额(元)'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'category',
      data: data.map((d) => d.date),
    },
    yAxis: [
      { type: 'value', name: '订单数' },
      { type: 'value', name: '销售额' },
    ],
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

onMounted(loadData)
</script>

<template>
  <div>
    <!-- 统计卡片 -->
    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="6">
        <div class="stats-card" v-loading="statsLoading">
          <div class="stats-label">今日订单</div>
          <div class="stats-value">{{ stats.today_order_count }}</div>
          <div class="stats-sub">单</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card" v-loading="statsLoading">
          <div class="stats-label">今日销售额</div>
          <div class="stats-value">{{ formatAmount(stats.today_sales) }}</div>
          <div class="stats-sub">元</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card" v-loading="statsLoading">
          <div class="stats-label">待发货</div>
          <div class="stats-value" style="color: #f59e0b">{{ stats.pending_ship }}</div>
          <div class="stats-sub">
            <el-link type="primary" @click="router.push('/shipping/pending')">去处理</el-link>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stats-card" v-loading="statsLoading">
          <div class="stats-label">售后处理中</div>
          <div class="stats-value" style="color: #ef4444">{{ stats.aftersale_pending }}</div>
          <div class="stats-sub">
            <el-link type="danger" @click="router.push('/aftersale/list')">去处理</el-link>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 库存预警提醒 -->
    <el-alert
      v-if="alertCount > 0"
      :title="`有 ${alertCount} 条库存预警待处理`"
      type="warning"
      show-icon
      style="margin-bottom: 16px"
    >
      <template #default>
        <el-link @click="router.push('/inventory/alerts')">查看详情</el-link>
      </template>
    </el-alert>

    <el-row :gutter="16">
      <!-- 趋势图 -->
      <el-col :span="16">
        <el-card>
          <template #header>近 7 天销售趋势</template>
          <VChart
            :option="trendOption"
            style="height: 280px"
            autoresize
          />
        </el-card>
      </el-col>

      <!-- 近期订单 -->
      <el-col :span="8">
        <el-card>
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>近期订单</span>
              <el-link @click="router.push('/order/list')">查看全部</el-link>
            </div>
          </template>
          <div style="max-height: 280px; overflow-y: auto">
            <div
              v-for="order in recentOrders"
              :key="order.id"
              class="order-item"
              @click="router.push(`/order/detail/${order.id}`)"
            >
              <div class="order-no">{{ order.order_no }}</div>
              <div class="order-info">
                <span>{{ formatAmount(order.pay_cents) }}</span>
                <el-tag
                  :type="orderStatusMap[order.status]?.type || ''"
                  size="small"
                >
                  {{ orderStatusMap[order.status]?.label || order.status }}
                </el-tag>
              </div>
              <div class="order-time">{{ formatTime(order.created_at) }}</div>
            </div>
            <el-empty v-if="!recentOrders.length" description="暂无订单" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.order-item {
  padding: 10px 0;
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  transition: background 0.2s;

  &:last-child {
    border-bottom: none;
  }

  &:hover {
    background: var(--el-color-primary-light-9);
    margin: 0 -16px;
    padding: 10px 16px;
    border-radius: 4px;
  }
}

.order-no {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.order-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 4px 0;
  font-size: 13px;
}

.order-time {
  font-size: 11px;
  color: var(--text-secondary);
}
</style>
