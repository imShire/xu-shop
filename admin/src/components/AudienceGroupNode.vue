<script lang="ts">
export interface Condition {
  kind: 'tag' | 'behavior'
  tag_code?: string
  has?: boolean
  event?: string
  op?: string
  value?: any
  days?: number
}
export interface Group {
  op: 'and' | 'or'
  conditions: Array<Condition | Group>
}
export const eventOptions = [
  { label: '已支付订单数', value: 'order_paid' },
  { label: '累计 GMV(分)', value: 'gmv_cents' },
  { label: '近 N 天活跃', value: 'last_active_days' },
  { label: '加购次数', value: 'cart_add_count' },
]
export const opOptions = [
  { label: '≥', value: 'gte' },
  { label: '≤', value: 'lte' },
  { label: '=', value: 'eq' },
]
export function isGroup(c: Condition | Group): c is Group {
  return (c as Group).op !== undefined && Array.isArray((c as Group).conditions)
}
</script>

<script setup lang="ts">
import { Plus, Delete } from '@element-plus/icons-vue'
import type { UserTag } from '@/api/tag'
import { tagCategoryLabels } from '@/api/tag'
import AudienceGroupNode from './AudienceGroupNode.vue'

const props = defineProps<{
  group: Group
  tags: UserTag[]
  level?: number
}>()

const level = props.level ?? 0

function addTag() {
  props.group.conditions.push({ kind: 'tag', tag_code: '', has: true })
}
function addBehavior() {
  props.group.conditions.push({ kind: 'behavior', event: 'order_paid', op: 'gte', value: 1, days: 30 })
}
function addGroup() {
  props.group.conditions.push({ op: 'and', conditions: [] })
}
function removeAt(idx: number) {
  props.group.conditions.splice(idx, 1)
}
</script>

<template>
  <div class="group-node" :style="{ marginLeft: level > 0 ? '16px' : '0' }">
    <div class="group-header">
      <el-radio-group v-model="group.op" size="small">
        <el-radio-button value="and">满足全部 (AND)</el-radio-button>
        <el-radio-button value="or">满足任一 (OR)</el-radio-button>
      </el-radio-group>
      <el-button-group style="margin-left: 12px">
        <el-button size="small" :icon="Plus" @click="addTag">加标签</el-button>
        <el-button size="small" :icon="Plus" @click="addBehavior">加行为</el-button>
        <el-button size="small" :icon="Plus" @click="addGroup">加分组</el-button>
      </el-button-group>
    </div>

    <div v-for="(c, idx) in group.conditions" :key="idx" class="cond-row">
      <!-- 嵌套分组 -->
      <AudienceGroupNode
        v-if="isGroup(c)"
        :group="(c as Group)"
        :tags="tags"
        :level="level + 1"
        style="flex: 1"
      />

      <!-- 标签条件 -->
      <div v-else-if="c.kind === 'tag'" class="cond-inline">
        <el-select v-model="c.has" size="small" style="width: 90px">
          <el-option :value="true" label="包含" />
          <el-option :value="false" label="不含" />
        </el-select>
        <el-select
          v-model="c.tag_code"
          filterable
          placeholder="选择标签"
          size="small"
          style="width: 260px; margin-left: 8px"
        >
          <el-option
            v-for="t in tags"
            :key="t.code"
            :label="`[${tagCategoryLabels[t.category] || t.category}] ${t.name}`"
            :value="t.code"
          />
        </el-select>
      </div>

      <!-- 行为条件 -->
      <div v-else class="cond-inline">
        <el-select v-model="c.event" size="small" style="width: 160px">
          <el-option v-for="o in eventOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-select v-model="c.op" size="small" style="width: 80px; margin-left: 8px">
          <el-option v-for="o in opOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-input-number
          v-model="c.value"
          controls-position="right"
          size="small"
          style="width: 130px; margin-left: 8px"
        />
        <span style="margin: 0 8px">在</span>
        <el-input-number
          v-model="c.days"
          controls-position="right"
          :min="1"
          size="small"
          style="width: 100px"
        />
        <span style="margin-left: 6px">天内</span>
      </div>

      <el-button
        :icon="Delete"
        size="small"
        text
        type="danger"
        style="margin-left: 8px"
        @click="removeAt(idx)"
      />
    </div>

    <div v-if="group.conditions.length === 0" class="empty-tip">尚未添加条件</div>
  </div>
</template>

<style scoped>
.group-node {
  border-left: 2px solid var(--el-color-primary-light-5);
  padding: 10px 12px;
  margin-bottom: 8px;
  background: var(--el-bg-color);
  border-radius: 4px;
}
.group-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}
.cond-row {
  display: flex;
  align-items: center;
  margin: 6px 0;
}
.cond-inline {
  display: flex;
  align-items: center;
  flex: 1;
}
.empty-tip {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
  padding: 8px 4px;
}
</style>
