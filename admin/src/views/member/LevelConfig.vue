<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listMemberLevels, updateMemberLevel, type MemberLevel } from '@/api/marketing'
import { yuanToCents, centsToYuan } from '@/utils/format'

interface Row extends MemberLevel {
  threshold_yuan: number
}

const list = ref<Row[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data = await listMemberLevels()
    list.value = (data || []).map((d) => ({
      ...d,
      benefits: d.benefits || {},
      threshold_yuan: Number(centsToYuan(d.threshold_amount_cents || 0)),
    }))
  } finally {
    loading.value = false
  }
}

async function save(row: Row) {
  const payload: MemberLevel = {
    code: row.code,
    name: row.name,
    threshold_amount_cents: yuanToCents(row.threshold_yuan || 0),
    points_multiplier: row.points_multiplier,
    benefits: row.benefits,
    sort: row.sort,
    is_active: row.is_active,
  }
  await updateMemberLevel(row.code, payload)
  ElMessage.success(`「${row.name}」已保存`)
}

onMounted(load)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <el-alert type="info" :closable="false" style="margin-bottom: 12px">
      升级阈值按累计支付金额计算；保级窗口、积分倍率均通过 benefits / config 字段配置（详见 docs/prd/16-membership.md）
    </el-alert>

    <el-table :data="list" border>
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column prop="code" label="代码" width="100" />
      <el-table-column label="等级名称" width="160">
        <template #default="{ row }"><el-input v-model="row.name" /></template>
      </el-table-column>
      <el-table-column label="升级阈值（元）" width="180">
        <template #default="{ row }">
          <el-input-number v-model="row.threshold_yuan" :min="0" :precision="2" />
        </template>
      </el-table-column>
      <el-table-column label="积分倍率" width="140">
        <template #default="{ row }">
          <el-input-number v-model="row.points_multiplier" :min="0" :step="0.1" :precision="2" />
        </template>
      </el-table-column>
      <el-table-column label="保级窗口（天）" width="160">
        <template #default="{ row }">
          <el-input-number
            :model-value="row.benefits?.keep_window_days ?? 365"
            :min="0"
            @update:model-value="(v: any) => (row.benefits = { ...row.benefits, keep_window_days: v })"
          />
        </template>
      </el-table-column>
      <el-table-column label="启用" width="80">
        <template #default="{ row }">
          <el-switch v-model="row.is_active" />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" size="small" text @click="save(row)">保存</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
