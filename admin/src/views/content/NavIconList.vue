<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import UploadImage from '@/components/UploadImage/index.vue'
import {
  getNavIcons,
  createNavIcon,
  updateNavIcon,
  deleteNavIcon,
  toggleNavIcon,
  sortNavIcons,
} from '@/api/nav-icon'
import type { NavIcon, NavIconForm } from '@/api/nav-icon'

const auth = useAuthStore()
const canEdit = () => auth.isSuperAdmin || auth.perms.includes('nav_icon.edit')

const loading = ref(false)
const navIcons = ref<NavIcon[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const defaultForm = (): NavIconForm => ({
  title: '',
  icon_url: '',
  link_url: '',
  sort: 0,
})

const form = ref<NavIconForm & { id?: string }>(defaultForm())

const previewVisible = ref(false)
const previewUrl = ref('')

async function loadData() {
  loading.value = true
  try {
    const res = await getNavIcons()
    navIcons.value = Array.isArray(res) ? res : []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = defaultForm()
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: NavIcon) {
  form.value = {
    id: row.id,
    title: row.title,
    icon_url: row.icon_url,
    link_url: row.link_url,
    sort: row.sort,
  }
  isEdit.value = true
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  if (!form.value.icon_url) {
    ElMessage.warning('请上传图标')
    return
  }
  saving.value = true
  try {
    const payload: NavIconForm = {
      title: form.value.title,
      icon_url: form.value.icon_url,
      link_url: form.value.link_url,
      sort: form.value.sort ?? 0,
    }
    if (isEdit.value && form.value.id) {
      await updateNavIcon(form.value.id, payload)
    } else {
      await createNavIcon(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleToggle(row: NavIcon) {
  try {
    await toggleNavIcon(row.id)
    await loadData()
  } catch {
    // revert on error — reload will restore correct state
    await loadData()
  }
}

async function handleSortChange(row: NavIcon) {
  try {
    await sortNavIcons([{ id: row.id, sort: row.sort }])
  } catch {
    ElMessage.error('排序更新失败')
    await loadData()
  }
}

async function handleDelete(row: NavIcon) {
  await ElMessageBox.confirm(
    `确认删除金刚区图标「${row.title || row.id}」？`,
    '提示',
    { type: 'warning' },
  )
  await deleteNavIcon(row.id)
  ElMessage.success('删除成功')
  await loadData()
}

function handlePreview(url: string) {
  previewUrl.value = url
  previewVisible.value = true
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="display: flex; justify-content: flex-end; margin-bottom: 16px">
      <el-button v-if="canEdit()" type="primary" @click="openCreate">+ 新增图标</el-button>
    </div>

    <el-table v-loading="loading" :data="navIcons" border style="width: 100%">
      <!-- 图标预览 -->
      <el-table-column label="图标预览" width="100" align="center">
        <template #default="{ row }">
          <el-image
            :src="row.icon_url"
            style="border-radius: 50%; width: 48px; height: 48px; object-fit: cover; cursor: pointer"
            fit="cover"
            @click="handlePreview(row.icon_url)"
          />
        </template>
      </el-table-column>

      <!-- 标题 -->
      <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />

      <!-- 跳转链接 -->
      <el-table-column prop="link_url" label="跳转链接" min-width="200">
        <template #default="{ row }">
          <span
            style="display: inline-block; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; vertical-align: middle"
            :title="row.link_url"
          >{{ row.link_url || '—' }}</span>
        </template>
      </el-table-column>

      <!-- 排序 -->
      <el-table-column label="排序" width="100" align="center">
        <template #default="{ row }">
          <el-input-number
            v-model="row.sort"
            :min="0"
            :max="9999"
            :controls="false"
            size="small"
            style="width: 70px"
            @change="handleSortChange(row)"
          />
        </template>
      </el-table-column>

      <!-- 状态 -->
      <el-table-column label="状态" width="90" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.is_active"
            :disabled="!canEdit()"
            @change="handleToggle(row)"
          />
        </template>
      </el-table-column>

      <!-- 操作 -->
      <el-table-column label="操作" width="140" align="center" v-if="canEdit()">
        <template #default="{ row }">
          <el-button text size="small" type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button text size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="!navIcons.length && !loading" style="display: flex; justify-content: center; padding: 40px 0">
      <el-empty description="暂无金刚区图标" />
    </div>

    <!-- 新增/编辑 Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑图标' : '新增图标'"
      width="480px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" label-width="90px">
        <el-form-item label="图标" prop="icon_url">
          <UploadImage v-model="form.icon_url" />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
            建议尺寸 200×200，圆形图标，JPG/PNG
          </div>
        </el-form-item>
        <el-form-item
          label="标题"
          prop="title"
          :rules="[{ required: true, message: '请输入标题', trigger: 'blur' }]"
        >
          <el-input v-model="form.title" placeholder="请输入标题" maxlength="16" show-word-limit />
        </el-form-item>
        <el-form-item label="跳转链接" prop="link_url">
          <el-input v-model="form.link_url" placeholder="可选，如 /pages/product/detail?id=1" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" :max="9999" style="width: 120px" />
          <span style="margin-left: 8px; font-size: 12px; color: var(--el-text-color-secondary)">数值越小越靠前</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 图标放大预览 -->
    <el-dialog v-model="previewVisible" title="图标预览" width="400px">
      <div style="display: flex; justify-content: center">
        <img :src="previewUrl" style="width: 200px; height: 200px; border-radius: 50%; object-fit: cover" />
      </div>
    </el-dialog>
  </div>
</template>
