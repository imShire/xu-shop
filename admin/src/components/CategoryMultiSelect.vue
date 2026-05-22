<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getCategoryList } from '@/api/product'
import type { Category } from '@/types'

interface CascaderOption {
  value: string
  label: string
  children?: CascaderOption[]
}

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    placeholder?: string
    disabled?: boolean
  }>(),
  { placeholder: '选择分类（可多选）' }
)

const emit = defineEmits<{ 'update:modelValue': [v: string[]] }>()

const options = ref<CascaderOption[]>([])

function toCascade(items: Category[] | undefined): CascaderOption[] {
  return (items || []).map((c) => ({
    value: c.id,
    label: c.name,
    children: c.children && c.children.length > 0 ? toCascade(c.children) : undefined,
  }))
}

onMounted(async () => {
  try {
    const data = await getCategoryList()
    options.value = toCascade(data as unknown as Category[])
  } catch {
    options.value = []
  }
})

const cascaderProps = {
  multiple: true,
  checkStrictly: true,
  emitPath: false,
  value: 'value',
  label: 'label',
  children: 'children',
}

function onChange(v: any) {
  emit('update:modelValue', Array.isArray(v) ? (v as string[]) : [])
}
</script>

<template>
  <el-cascader
    :model-value="props.modelValue"
    :options="options"
    :props="cascaderProps"
    :placeholder="placeholder"
    :disabled="disabled"
    clearable
    filterable
    collapse-tags
    collapse-tags-tooltip
    style="width: 100%"
    @update:model-value="onChange"
  />
</template>
