<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import {
  listRecallCampaigns,
  onlineRecallCampaign,
  pauseRecallCampaign,
  closeRecallCampaign,
  type RecallCampaign,
} from '@/api/recall'
import { formatTime } from '@/utils/format'

const router = useRouter()
const searchForm = ref({ keyword: '', status: '', goal: '' })

const { list, total, page, pageSize, loading, fetch } = useTable<RecallCampaign>((p) =>
  listRecallCampaigns({ ...p, ...searchForm.value })
)

const statusMap: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'danger' }> = {
  draft: { label: '草稿', type: 'info' },
  online: { label: '运行中', type: 'success' },
  paused: { label: '已暂停', type: 'warning' },
  closed: { label: '已关闭', type: 'danger' },
}

const goalMap: Record<string, string> = {
  silent_wakeup: '沉默唤醒',
  cart_abandon: '弃购召回',
  birthday: '生日营销',
  category_cross: '品类交叉',
  member_upgrade: '会员升级',
}

const columns: ColumnDef[] = [
  { label: '活动 ID', prop: 'id', width: 120 },
  { label: '名称', prop: 'name', minWidth: 180 },
  { label: '目标', slot: 'goal', width: 120 },
  { label: '状态', slot: 'status', width: 100, align: 'center' },
  {
    label: '生效时间',
    width: 280,
    formatter: (r) => `${formatTime(r.effective_from)} ~ ${formatTime(r.effective_to)}`,
  },
  { label: '更新时间', prop: 'updated_at', width: 160, formatter: (r) => formatTime(r.updated_at) },
]

async function online(row: RecallCampaign) {
  await ElMessageBox.confirm(`确认上线活动「${row.name}」？将立即开始触发计算`, '上线', { type: 'warning' })
  await onlineRecallCampaign(row.id)
  ElMessage.success('已上线')
  fetch(searchForm.value)
}

async function pause(row: RecallCampaign) {
  await ElMessageBox.confirm(`确认暂停活动「${row.name}」？`, '暂停', { type: 'warning' })
  await pauseRecallCampaign(row.id)
  ElMessage.success('已暂停')
  fetch(searchForm.value)
}

async function close(row: RecallCampaign) {
  await ElMessageBox.confirm(`确认关闭活动「${row.name}」？关闭后无法恢复`, '关闭', { type: 'error' })
  await closeRecallCampaign(row.id)
  ElMessage.success('已关闭')
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
          <el-option v-for="(v, k) in statusMap" :key="k" :label="v.label" :value="k" />
        </el-select>
        <el-select v-model="searchForm.goal" placeholder="目标" clearable style="width: 140px">
          <el-option v-for="(v, k) in goalMap" :key="k" :label="v" :value="k" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>
      <template #toolbar>
        <el-button type="primary" @click="router.push('/recall/campaign/create')">新建活动</el-button>
      </template>

      <template #goal="{ row }">{{ goalMap[row.goal] || row.goal }}</template>
      <template #status="{ row }">
        <el-tag :type="statusMap[row.status]?.type" size="small">{{ statusMap[row.status]?.label }}</el-tag>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="router.push(`/recall/campaign/edit/${row.id}`)">
          {{ row.status === 'draft' ? '编辑' : '查看' }}
        </el-button>
        <el-button text type="primary" size="small" @click="router.push(`/recall/campaign/funnel/${row.id}`)">漏斗</el-button>
        <el-button v-if="row.status === 'draft' || row.status === 'paused'" text type="success" size="small" @click="online(row)">
          {{ row.status === 'paused' ? '恢复' : '上线' }}
        </el-button>
        <el-button v-if="row.status === 'online'" text type="warning" size="small" @click="pause(row)">暂停</el-button>
        <el-button v-if="row.status !== 'closed'" text type="danger" size="small" @click="close(row)">关闭</el-button>
      </template>
    </ProTable>
  </div>
</template>
