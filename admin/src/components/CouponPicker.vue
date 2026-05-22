<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { listCouponTemplates, type CouponTemplate } from '@/api/marketing'

const props = withDefaults(
  defineProps<{
    modelValue: string | null
    placeholder?: string
    onlyOnline?: boolean
  }>(),
  { placeholder: '选择券模板', onlyOnline: true }
)

const emit = defineEmits<{ 'update:modelValue': [v: string | null] }>()

const list = ref<CouponTemplate[]>([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const res = await listCouponTemplates({
      page: 1,
      page_size: 200,
      ...(props.onlyOnline ? { status: 'online' } : {}),
    })
    list.value = (res?.list ?? []) as CouponTemplate[]
  } finally {
    loading.value = false
  }
}

onMounted(load)
defineExpose({ reload: load })

function typeLabel(t: CouponTemplate['type']): string {
  return ({ amount: '满减', discount: '折扣', no_threshold: '无门槛', exchange: '兑换' } as const)[t]
}
</script>

<template>
  <el-select
    :model-value="props.modelValue"
    filterable
    :placeholder="placeholder"
    :loading="loading"
    clearable
    style="width: 100%"
    @update:model-value="(v: any) => emit('update:modelValue', v ?? null)"
  >
    <el-option
      v-for="t in list"
      :key="t.id"
      :label="`[${typeLabel(t.type)}] ${t.name}`"
      :value="t.id"
    />
  </el-select>
</template>
