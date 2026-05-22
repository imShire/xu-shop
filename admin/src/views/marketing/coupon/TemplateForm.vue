<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  getCouponTemplate,
  createCouponTemplate,
  updateCouponTemplate,
  type CouponTemplateForm,
} from '@/api/marketing'
import { yuanToCents, centsToYuan } from '@/utils/format'
import CategoryMultiSelect from '@/components/CategoryMultiSelect.vue'

const route = useRoute()
const router = useRouter()
const id = computed(() => (route.params.id as string) || '')
const isEdit = computed(() => !!id.value)
const readOnly = ref(false)

interface FormState {
  name: string
  description: string
  type: 'amount' | 'discount' | 'no_threshold' | 'exchange'
  value_yuan: number
  discount_rate: number
  max_discount_yuan: number
  min_amount_yuan: number
  scope_type: 'all' | 'category' | 'spu' | 'sku'
  scope_targets: string[]
  validity_mode: 'absolute' | 'relative'
  valid_range: [string, string] | null
  valid_days: number
  total_quota: number
  per_user_limit: number
  per_order_limit: number
  stack_with_points: boolean
  claim_range: [string, string] | null
}

const form = ref<FormState>({
  name: '',
  description: '',
  type: 'amount',
  value_yuan: 10,
  discount_rate: 9,
  max_discount_yuan: 0,
  min_amount_yuan: 0,
  scope_type: 'all',
  scope_targets: [],
  validity_mode: 'absolute',
  valid_range: null,
  valid_days: 7,
  total_quota: 0,
  per_user_limit: 1,
  per_order_limit: 1,
  stack_with_points: true,
  claim_range: null,
})

const rules = {
  name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
}

const formRef = ref<any>()

async function load() {
  if (!isEdit.value) return
  const t = await getCouponTemplate(id.value)
  readOnly.value = t.status === 'online'
  form.value = {
    name: t.name,
    description: t.description || '',
    type: t.type,
    value_yuan: Number(centsToYuan(t.value_cents || 0)),
    discount_rate: t.discount_rate || 9,
    max_discount_yuan: Number(centsToYuan(t.max_discount_cents || 0)),
    min_amount_yuan: Number(centsToYuan(t.min_amount_cents || 0)),
    scope_type: t.scope_type,
    scope_targets: t.scope_targets || [],
    validity_mode: t.validity_mode,
    valid_range:
      t.valid_from && t.valid_to ? [t.valid_from, t.valid_to] : null,
    valid_days: t.valid_days || 7,
    total_quota: t.total_quota || 0,
    per_user_limit: t.per_user_limit || 1,
    per_order_limit: t.per_order_limit || 1,
    stack_with_points: t.stack_with_points,
    claim_range:
      t.claim_start_at && t.claim_end_at ? [t.claim_start_at, t.claim_end_at] : null,
  }
}

function buildPayload(): CouponTemplateForm {
  return {
    name: form.value.name,
    description: form.value.description,
    type: form.value.type,
    value_cents: yuanToCents(form.value.value_yuan || 0),
    discount_rate: form.value.type === 'discount' ? form.value.discount_rate : null,
    max_discount_cents: yuanToCents(form.value.max_discount_yuan || 0),
    min_amount_cents: yuanToCents(form.value.min_amount_yuan || 0),
    scope_type: form.value.scope_type,
    scope_targets: form.value.scope_targets,
    validity_mode: form.value.validity_mode,
    valid_from: form.value.validity_mode === 'absolute' ? form.value.valid_range?.[0] || null : null,
    valid_to: form.value.validity_mode === 'absolute' ? form.value.valid_range?.[1] || null : null,
    valid_days: form.value.validity_mode === 'relative' ? form.value.valid_days : null,
    total_quota: form.value.total_quota,
    per_user_limit: form.value.per_user_limit,
    per_order_limit: form.value.per_order_limit,
    stack_with_points: form.value.stack_with_points,
    claim_start_at: form.value.claim_range?.[0] || null,
    claim_end_at: form.value.claim_range?.[1] || null,
  }
}

async function submit() {
  await formRef.value?.validate()
  const payload = buildPayload()
  if (isEdit.value) {
    await updateCouponTemplate(id.value, payload)
    ElMessage.success('已更新')
  } else {
    await createCouponTemplate(payload)
    ElMessage.success('已创建')
  }
  router.push('/marketing/coupon/templates')
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <el-page-header :content="isEdit ? '编辑券模板' : '新建券模板'" @back="router.back()" style="margin-bottom: 16px" />
    <el-alert v-if="readOnly" type="warning" :closable="false" style="margin-bottom: 12px">
      模板已上架，仅可查看；如需修改请先下架
    </el-alert>

    <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" :disabled="readOnly">
      <el-divider content-position="left">基本信息</el-divider>
      <el-form-item label="模板名称" prop="name">
        <el-input v-model="form.name" maxlength="128" show-word-limit style="max-width: 480px" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="2" style="max-width: 600px" />
      </el-form-item>

      <el-divider content-position="left">优惠规则</el-divider>
      <el-form-item label="类型" prop="type">
        <el-radio-group v-model="form.type">
          <el-radio value="amount">满减</el-radio>
          <el-radio value="discount">折扣</el-radio>
          <el-radio value="no_threshold">无门槛</el-radio>
          <el-radio value="exchange">兑换券</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="form.type === 'discount'" label="折扣">
        <el-input-number v-model="form.discount_rate" :min="0.1" :max="9.9" :step="0.1" :precision="1" />
        <span style="margin-left: 8px; color: var(--el-text-color-secondary)">折（如 8.5 表示 85%）</span>
      </el-form-item>
      <el-form-item v-else label="面额（元）">
        <el-input-number v-model="form.value_yuan" :min="0" :precision="2" />
      </el-form-item>

      <el-form-item v-if="form.type === 'discount'" label="最高优惠（元）">
        <el-input-number v-model="form.max_discount_yuan" :min="0" :precision="2" />
        <span style="margin-left: 8px; color: var(--el-text-color-secondary)">0 表示不限</span>
      </el-form-item>

      <el-form-item v-if="form.type !== 'no_threshold'" label="使用门槛（元）">
        <el-input-number v-model="form.min_amount_yuan" :min="0" :precision="2" />
        <span style="margin-left: 8px; color: var(--el-text-color-secondary)">订单达到此金额可用，0 表示无门槛</span>
      </el-form-item>

      <el-form-item label="可叠积分">
        <el-switch v-model="form.stack_with_points" />
      </el-form-item>

      <el-divider content-position="left">适用范围</el-divider>
      <el-form-item label="范围类型">
        <el-radio-group v-model="form.scope_type">
          <el-radio value="all">全场</el-radio>
          <el-radio value="category">指定品类</el-radio>
          <el-radio value="spu">指定 SPU</el-radio>
          <el-radio value="sku">指定 SKU</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.scope_type === 'category'" label="选择品类">
        <CategoryMultiSelect v-model="form.scope_targets" />
      </el-form-item>
      <el-form-item v-if="form.scope_type === 'spu' || form.scope_type === 'sku'" label="目标 ID">
        <el-input
          :model-value="form.scope_targets.join(',')"
          placeholder="多个用英文逗号分隔"
          @update:model-value="(v: string) => (form.scope_targets = v.split(',').map((s) => s.trim()).filter(Boolean))"
          style="max-width: 600px"
        />
      </el-form-item>

      <el-divider content-position="left">有效期</el-divider>
      <el-form-item label="有效期模式">
        <el-radio-group v-model="form.validity_mode">
          <el-radio value="absolute">固定时间段</el-radio>
          <el-radio value="relative">领取后 N 天</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item v-if="form.validity_mode === 'absolute'" label="时间段">
        <el-date-picker
          v-model="form.valid_range"
          type="datetimerange"
          start-placeholder="开始"
          end-placeholder="结束"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
        />
      </el-form-item>
      <el-form-item v-else label="有效天数">
        <el-input-number v-model="form.valid_days" :min="1" :max="3650" />
      </el-form-item>

      <el-divider content-position="left">发放规则</el-divider>
      <el-form-item label="领取时间">
        <el-date-picker
          v-model="form.claim_range"
          type="datetimerange"
          start-placeholder="开始"
          end-placeholder="结束"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
        />
      </el-form-item>
      <el-form-item label="总量上限">
        <el-input-number v-model="form.total_quota" :min="0" />
        <span style="margin-left: 8px; color: var(--el-text-color-secondary)">0 表示不限</span>
      </el-form-item>
      <el-form-item label="每人限领">
        <el-input-number v-model="form.per_user_limit" :min="1" />
      </el-form-item>
      <el-form-item label="单单限用">
        <el-input-number v-model="form.per_order_limit" :min="1" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :disabled="readOnly" @click="submit">保存</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
