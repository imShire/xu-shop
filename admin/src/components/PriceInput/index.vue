<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: number
  placeholder?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const yuan = computed({
  get: () => (props.modelValue / 100).toFixed(2),
  set: (val: string) => {
    const num = parseFloat(val)
    emit('update:modelValue', isNaN(num) ? 0 : Math.round(num * 100))
  },
})
</script>

<template>
  <el-input
    v-model="yuan"
    :placeholder="placeholder || '请输入金额'"
    :disabled="disabled"
    type="number"
    step="0.01"
    min="0"
  >
    <template #prefix>¥</template>
    <template #suffix>元</template>
  </el-input>
</template>
