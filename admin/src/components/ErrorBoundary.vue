<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue'
import { report } from '@/utils/clog'

const props = defineProps<{
  // 用于标识出错位置（写入 extra.boundary）
  name?: string
}>()

const error = ref<Error | null>(null)
const reported = ref(false)

onErrorCaptured((err, instance, info) => {
  error.value = err instanceof Error ? err : new Error(String(err))
  try {
    report('error', error.value.message, {
      stack: error.value.stack,
      extra: {
        boundary: props.name ?? 'unnamed',
        info,
        component: (instance?.$options as { name?: string } | undefined)?.name,
      },
    })
    reported.value = true
  } catch {
    // 静默
  }
  // eslint-disable-next-line no-console
  console.error('[ErrorBoundary]', err, info)
  return false
})

function retry() {
  error.value = null
  reported.value = false
}
</script>

<template>
  <div v-if="error" class="error-boundary">
    <el-result icon="error" title="页面出现异常" :sub-title="error.message">
      <template #extra>
        <el-button type="primary" @click="retry">重试</el-button>
        <span v-if="reported" class="reported-tip">已自动上报</span>
      </template>
    </el-result>
  </div>
  <slot v-else />
</template>

<style scoped>
.error-boundary {
  padding: 24px;
}
.reported-tip {
  margin-left: 12px;
  color: var(--text-secondary, #909399);
  font-size: 12px;
}
</style>
