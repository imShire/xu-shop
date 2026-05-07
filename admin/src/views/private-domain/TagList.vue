<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { getTagList, createTag, updateTag, deleteTag } from '@/api/private-domain'

interface TagItem {
  id: string
  name: string
  source: string
  created_at: string
}

interface TagForm {
  id: string
  name: string
}

const loading = ref(false)
const tags = ref<TagItem[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const form = ref<TagForm>({ id: '', name: '' })

const colors = ['#f59e0b', '#3b82f6', '#22c55e', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4']

async function loadData() {
  loading.value = true
  try {
    const res = await getTagList()
    tags.value = Array.isArray(res) ? res : []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = { id: '', name: '' }
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: TagItem) {
  form.value = { id: row.id, name: row.name }
  isEdit.value = true
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    if (isEdit.value) {
      await updateTag(form.value.id, form.value)
    } else {
      await createTag(form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

function getTagColor(id: string) {
  const seed = Number(id) || 0
  return colors[Math.abs(seed) % colors.length]
}

async function handleDelete(row: TagItem) {
  try {
    await ElMessageBox.confirm(`确认删除标签「${row.name}」？`, '提示', { type: 'warning' })
    await deleteTag(row.id)
    ElMessage.success('删除成功')
    await loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">客户标签</h3>
      <el-button type="primary" @click="openCreate">+ 新建标签</el-button>
    </div>

    <div v-loading="loading" style="display: flex; flex-wrap: wrap; gap: 12px">
      <div
        v-for="tag in tags"
        :key="tag.id"
        style="display: flex; align-items: center; gap: 8px; padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 8px; background: var(--card-bg)"
      >
        <div
          :style="{ width: '10px', height: '10px', borderRadius: '50%', background: getTagColor(tag.id), flexShrink: 0 }"
        />
        <span style="font-weight: 500">{{ tag.name }}</span>
        <span style="font-size: 12px; color: var(--text-secondary)">{{ tag.source }}</span>
        <el-button text size="small" type="primary" @click="openEdit(tag)">编辑</el-button>
        <el-button text size="small" type="danger" @click="handleDelete(tag)">删除</el-button>
      </div>
    </div>
    <div v-if="!tags.length && !loading" style="display: flex; justify-content: center; padding: 40px 0">
      <el-empty description="暂无标签" />
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑标签' : '新建标签'" width="380px">
      <el-form ref="formRef" :model="form" label-width="70px">
        <el-form-item label="标签名" prop="name" :rules="[{ required: true }]">
          <el-input v-model="form.name" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
