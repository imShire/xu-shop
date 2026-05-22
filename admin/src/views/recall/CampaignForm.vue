<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import AudienceFilterBuilder from '@/components/AudienceFilterBuilder.vue'
import CouponPicker from '@/components/CouponPicker.vue'
import {
  createRecallCampaign,
  updateRecallCampaign,
  getRecallCampaign,
  type RecallCampaignForm,
} from '@/api/recall'

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string | undefined)
const isEdit = computed(() => !!id.value)
const readOnly = ref(false)
const loading = ref(false)
const dateRange = ref<[string, string] | null>(null)

const form = ref<RecallCampaignForm & { id?: string; status?: string }>({
  name: '',
  goal: 'silent_wakeup',
  audience_filter: { op: 'and', conditions: [] },
  actions: [],
  trigger_type: 'cron',
  trigger_config: { cron: '0 10 * * *' },
  effective_from: null,
  effective_to: null,
  throttle_per_user_days: 14,
  daily_quota: 0,
  total_quota: 0,
  attribution_window_days: 7,
})

const goalOptions = [
  { value: 'silent_wakeup', label: '沉默唤醒' },
  { value: 'cart_abandon', label: '弃购召回' },
  { value: 'birthday', label: '生日营销' },
  { value: 'category_cross', label: '品类交叉' },
  { value: 'member_upgrade', label: '会员升级' },
]

const actionTypes = [
  { value: 'send_coupon', label: '发优惠券' },
  { value: 'wxapp_subscribe', label: '小程序订阅消息' },
  { value: 'in_app_msg', label: '站内消息' },
]

function addAction() {
  form.value.actions.push({ type: 'send_coupon', config: {} })
}
function removeAction(i: number) {
  form.value.actions.splice(i, 1)
}

async function load() {
  if (!id.value) return
  loading.value = true
  try {
    const data = await getRecallCampaign(id.value)
    form.value = { ...data }
    readOnly.value = data.status !== 'draft'
    if (data.effective_from && data.effective_to) {
      dateRange.value = [data.effective_from, data.effective_to]
    }
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!form.value.name) {
    ElMessage.warning('请填写名称')
    return
  }
  if (form.value.actions.length === 0) {
    ElMessage.warning('至少配置一个动作')
    return
  }
  if (dateRange.value) {
    form.value.effective_from = dateRange.value[0]
    form.value.effective_to = dateRange.value[1]
  }
  const payload: RecallCampaignForm = {
    name: form.value.name,
    goal: form.value.goal,
    audience_filter: form.value.audience_filter,
    actions: form.value.actions,
    trigger_type: form.value.trigger_type,
    trigger_config: form.value.trigger_config,
    effective_from: form.value.effective_from,
    effective_to: form.value.effective_to,
    throttle_per_user_days: form.value.throttle_per_user_days,
    daily_quota: form.value.daily_quota,
    total_quota: form.value.total_quota,
    attribution_window_days: form.value.attribution_window_days,
  }
  if (isEdit.value) {
    await updateRecallCampaign(id.value!, payload)
    ElMessage.success('已保存')
  } else {
    await createRecallCampaign(payload)
    ElMessage.success('已创建')
  }
  router.push('/recall/campaigns')
}

onMounted(load)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <el-page-header :content="isEdit ? '编辑召回活动' : '新建召回活动'" @back="router.back()" />
    <el-divider />

    <el-form label-width="140px" :disabled="readOnly">
      <el-divider content-position="left">基本信息</el-divider>
      <el-form-item label="活动名称" required>
        <el-input v-model="form.name" maxlength="64" style="max-width: 360px" />
      </el-form-item>
      <el-form-item label="目标" required>
        <el-select v-model="form.goal" style="max-width: 240px">
          <el-option v-for="o in goalOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
      </el-form-item>

      <el-divider content-position="left">受众</el-divider>
      <AudienceFilterBuilder v-model="form.audience_filter" />

      <el-divider content-position="left">动作</el-divider>
      <div v-for="(a, i) in form.actions" :key="i" style="display: flex; gap: 8px; margin-bottom: 8px; align-items: center">
        <el-select v-model="a.type" style="width: 200px">
          <el-option v-for="t in actionTypes" :key="t.value" :label="t.label" :value="t.value" />
        </el-select>
        <CouponPicker
          v-if="a.type === 'send_coupon'"
          v-model="a.config.template_id"
          :only-online="false"
          style="flex: 1"
        />
        <el-input
          v-else-if="a.type === 'wxapp_subscribe'"
          v-model="a.config.template_id"
          placeholder="订阅消息模板 ID"
          style="flex: 1"
        />
        <el-input
          v-else-if="a.type === 'in_app_msg'"
          v-model="a.config.content"
          placeholder="站内消息内容"
          style="flex: 1"
        />
        <el-button type="danger" link @click="removeAction(i)">删除</el-button>
      </div>
      <el-button @click="addAction">+ 添加动作</el-button>

      <el-divider content-position="left">触发器</el-divider>
      <el-form-item label="触发类型" required>
        <el-radio-group v-model="form.trigger_type">
          <el-radio-button value="cron">定时</el-radio-button>
          <el-radio-button value="event">事件</el-radio-button>
          <el-radio-button value="immediate">立即</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.trigger_type === 'cron'" label="Cron 表达式">
        <el-input v-model="form.trigger_config.cron" placeholder="0 10 * * *" style="max-width: 280px" />
      </el-form-item>
      <el-form-item v-else-if="form.trigger_type === 'event'" label="事件名">
        <el-input v-model="form.trigger_config.event" placeholder="cart_abandon_30min" style="max-width: 280px" />
      </el-form-item>

      <el-divider content-position="left">生效与配额</el-divider>
      <el-form-item label="生效时间">
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          start-placeholder="开始"
          end-placeholder="结束"
        />
      </el-form-item>
      <el-form-item label="单用户节流（天）">
        <el-input-number v-model="form.throttle_per_user_days" :min="0" />
      </el-form-item>
      <el-form-item label="每日配额">
        <el-input-number v-model="form.daily_quota" :min="0" />
        <span style="margin-left: 8px; color: var(--el-text-color-secondary)">0 表示不限</span>
      </el-form-item>
      <el-form-item label="总配额">
        <el-input-number v-model="form.total_quota" :min="0" />
        <span style="margin-left: 8px; color: var(--el-text-color-secondary)">0 表示不限</span>
      </el-form-item>
      <el-form-item label="归因窗口（天）">
        <el-input-number v-model="form.attribution_window_days" :min="0" :max="90" />
      </el-form-item>

      <el-form-item v-if="!readOnly">
        <el-button type="primary" @click="submit">保存</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
