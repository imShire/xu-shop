<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import AudienceGroupNode from './AudienceGroupNode.vue'
import type { Group } from './AudienceGroupNode.vue'
import { listUserTags, type UserTag } from '@/api/tag'

const props = defineProps<{ modelValue: Group | Record<string, any> | null }>()
const emit = defineEmits<{ 'update:modelValue': [v: Group] }>()

function ensureGroup(v: any): Group {
  if (v && (v.op === 'and' || v.op === 'or') && Array.isArray(v.conditions)) {
    return v as Group
  }
  return { op: 'and', conditions: [] }
}

const root = ref<Group>(ensureGroup(props.modelValue))

watch(
  () => props.modelValue,
  (v) => {
    root.value = ensureGroup(v)
  }
)

watch(
  root,
  (v) => emit('update:modelValue', v),
  { deep: true }
)

const tags = ref<UserTag[]>([])
onMounted(async () => {
  try {
    tags.value = (await listUserTags()) || []
  } catch {
    tags.value = []
  }
})

defineExpose({ value: root })
</script>

<template>
  <div class="audience-builder">
    <AudienceGroupNode :group="root" :tags="tags" :level="0" />
  </div>
</template>

<style scoped>
.audience-builder {
  border: 1px dashed var(--el-border-color);
  border-radius: 6px;
  padding: 12px;
  background: var(--el-fill-color-lighter);
}
</style>
