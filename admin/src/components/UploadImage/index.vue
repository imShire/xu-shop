<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { uploadImage } from '@/api/product'

interface Props {
  modelValue?: string | string[]
  multiple?: boolean
  limit?: number
  accept?: string
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  limit: 9,
  accept: 'image/*',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | string[]]
}>()

const uploading = ref(false)
const previewVisible = ref(false)
const previewUrl = ref('')

function getList(): string[] {
  if (!props.modelValue) return []
  if (Array.isArray(props.modelValue)) return [...props.modelValue]
  return props.modelValue ? [props.modelValue] : []
}

async function handleChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return

  const files = Array.from(input.files)
  uploading.value = true
  try {
    const urls: string[] = []
    for (const file of files) {
      const res = await uploadImage(file)
      urls.push(res.url)
    }

    if (props.multiple) {
      const list = [...getList(), ...urls].slice(0, props.limit)
      emit('update:modelValue', list)
    } else {
      emit('update:modelValue', urls[0])
    }
  } catch {
    ElMessage.error('上传失败')
  } finally {
    uploading.value = false
    input.value = ''
  }
}

function removeImage(idx: number) {
  if (props.multiple) {
    const list = getList()
    list.splice(idx, 1)
    emit('update:modelValue', list)
  } else {
    emit('update:modelValue', '')
  }
}

function preview(url: string) {
  previewUrl.value = url
  previewVisible.value = true
}
</script>

<template>
  <div class="upload-image">
    <div class="image-list">
      <div
        v-for="(url, idx) in getList()"
        :key="idx"
        class="image-item"
      >
        <img :src="url" class="preview-img" @click="preview(url)" />
        <div class="image-actions">
          <el-icon class="action-icon" @click="preview(url)"><ZoomIn /></el-icon>
          <el-icon class="action-icon" @click="removeImage(idx)"><Delete /></el-icon>
        </div>
      </div>

      <label
        v-if="multiple ? getList().length < limit : !modelValue"
        class="upload-trigger"
        :class="{ uploading }"
      >
        <el-icon v-if="!uploading" size="24"><Plus /></el-icon>
        <el-icon v-else class="rotating" size="24"><Loading /></el-icon>
        <input
          type="file"
          :accept="accept"
          :multiple="multiple"
          style="display: none"
          @change="handleChange"
        />
      </label>
    </div>

    <el-dialog v-model="previewVisible" title="图片预览" width="600px">
      <img :src="previewUrl" style="width: 100%; border-radius: 4px" />
    </el-dialog>
  </div>
</template>

<style scoped>
.upload-image {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.image-item {
  width: 80px;
  height: 80px;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
  border: 1px solid var(--border-color);

  &:hover .image-actions {
    opacity: 1;
  }
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: pointer;
}

.image-actions {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.action-icon {
  color: #fff;
  cursor: pointer;
  font-size: 16px;

  &:hover {
    color: var(--el-color-primary);
  }
}

.upload-trigger {
  width: 80px;
  height: 80px;
  border: 1px dashed var(--border-color);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.2s;

  &:hover {
    border-color: var(--el-color-primary);
    color: var(--el-color-primary);
  }

  &.uploading {
    pointer-events: none;
    opacity: 0.6;
  }
}

.rotating {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
