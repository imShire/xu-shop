<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listPointRules, updatePointRule, type PointRule } from '@/api/marketing'

const rules = ref<PointRule[]>([])
const loading = ref(false)

// 各规则的字段配置 — 与 docs/prd/16-membership.md 对齐
const ruleSchemas: Record<string, Array<{ key: string; label: string; type: 'number' | 'days' | 'percent' }>> = {
  order_pay: [
    { key: 'rate', label: '返点比例 (元 / 1 分)', type: 'number' },
    { key: 'arrive_after_days', label: 'T+N 入账天数', type: 'days' },
  ],
  signin: [{ key: 'points', label: '每日签到积分', type: 'number' }],
  invite: [{ key: 'points', label: '邀请成功积分', type: 'number' }],
  review: [{ key: 'points', label: '评价奖励积分', type: 'number' }],
  birthday: [{ key: 'points', label: '生日积分', type: 'number' }],
  register: [{ key: 'points', label: '注册积分', type: 'number' }],
  deduct: [
    { key: 'rate', label: '抵扣比例（100 分 = ? 元）', type: 'number' },
    { key: 'max_per_order_pct', label: '单订单封顶百分比', type: 'percent' },
  ],
  expire: [
    { key: 'mode', label: '过期策略 (year_end / fixed_days)', type: 'number' },
    { key: 'days', label: '固定有效天数', type: 'days' },
  ],
}

const ruleNames: Record<string, string> = {
  order_pay: '订单支付返点',
  signin: '每日签到',
  invite: '邀请好友',
  review: '订单评价',
  birthday: '生日礼',
  register: '注册',
  deduct: '抵扣比例',
  expire: '过期策略',
}

async function load() {
  loading.value = true
  try {
    rules.value = await listPointRules()
  } finally {
    loading.value = false
  }
}

async function save(rule: PointRule) {
  await updatePointRule(rule.code, { enabled: rule.enabled, config: rule.config })
  ElMessage.success(`「${ruleNames[rule.code] || rule.code}」已保存`)
}

onMounted(load)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <el-alert type="info" :closable="false" style="margin-bottom: 16px">
      改动会影响后续新发生的积分事件，已发放积分不变
    </el-alert>

    <el-card v-for="rule in rules" :key="rule.code" shadow="never" style="margin-bottom: 12px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-weight: 600">{{ ruleNames[rule.code] || rule.code }}</span>
          <el-switch v-model="rule.enabled" />
        </div>
      </template>

      <el-form label-width="180px" :inline="false">
        <el-form-item
          v-for="field in ruleSchemas[rule.code] || []"
          :key="field.key"
          :label="field.label"
        >
          <el-input
            v-model="rule.config[field.key]"
            style="max-width: 220px"
            :placeholder="String(field.type)"
          />
        </el-form-item>
        <el-form-item v-if="!ruleSchemas[rule.code]" label="原始配置 (JSON)">
          <el-input
            :model-value="JSON.stringify(rule.config, null, 2)"
            type="textarea"
            :rows="4"
            @update:model-value="
              (v: string) => {
                try {
                  rule.config = JSON.parse(v)
                } catch {}
              }
            "
            style="max-width: 600px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save(rule)">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>
