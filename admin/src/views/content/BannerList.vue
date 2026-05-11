<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import UploadImage from '@/components/UploadImage/index.vue'
import LinkPicker from '@/components/LinkPicker/index.vue'
import type { LinkConfig } from '@/types/link'
import {
  getBanners,
  createBanner,
  updateBanner,
  deleteBanner,
  toggleBanner,
  sortBanners,
} from '@/api/banner'
import type { Banner, BannerForm } from '@/api/banner'

const auth = useAuthStore()
const canEdit = () => auth.isSuperAdmin || auth.perms.includes('banner.edit')

const loading = ref(false)
const banners = ref<Banner[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const defaultForm = (): BannerForm & { link_config: LinkConfig | null } => ({
  title: '',
  image_url: '',
  link_url: '',
  sort: 0,
  link_config: null,
})

const form = ref<BannerForm & { id?: string; link_config: LinkConfig | null }>(defaultForm())

watch(() => form.value.link_config, (val) => {
  if (val?.url) form.value.link_url = val.url
})

const previewVisible = ref(false)
const previewUrl = ref('')

async function loadData() {
  loading.value = true
  try {
    const res = await getBanners()
    banners.value = Array.isArray(res) ? res : []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = defaultForm()
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: Banner) {
  form.value = {
    id: row.id,
    title: row.title,
    image_url: row.image_url,
    link_url: row.link_url,
    sort: row.sort,
    link_config: row.link_config ? { ...row.link_config } : null,
  }
  isEdit.value = true
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  if (!form.value.image_url) {
    ElMessage.warning('请上传图片')
    return
  }
  saving.value = true
  try {
    const payload: BannerForm = {
      title: form.value.title,
      image_url: form.value.image_url,
      link_url: form.value.link_url,
      sort: form.value.sort ?? 0,
      link_config: form.value.link_config ?? null,
    }
    if (isEdit.value && form.value.id) {
      await updateBanner(form.value.id, payload)
    } else {
      await createBanner(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleToggle(row: Banner) {
  try {
    await toggleBanner(row.id)
    await loadData()
  } catch {
    // revert on error — reload will restore correct state
    await loadData()
  }
}

async function handleSortChange(row: Banner) {
  try {
    await sortBanners([{ id: row.id, sort: row.sort }])
  } catch {
    ElMessage.error('排序更新失败')
    await loadData()
  }
}

async function handleDelete(row: Banner) {
  await ElMessageBox.confirm(
    `确认删除 Banner「${row.title || row.id}」？`,
    '提示',
    { type: 'warning' },
  )
  await deleteBanner(row.id)
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
      <el-button v-if="canEdit()" type="primary" @click="openCreate">+ 新增 Banner</el-button>
    </div>

    <el-table v-loading="loading" :data="banners" border style="width: 100%">
      <!-- 图片预览 -->
      <el-table-column label="图片预览" width="100" align="center">
        <template #default="{ row }">
          <el-image
            :src="row.image_url"
            style="width: 60px; height: 40px; object-fit: cover; cursor: pointer; border-radius: 4px"
            fit="cover"
            @click="handlePreview(row.image_url)"
          />
        </template>
      </el-table-column>

      <!-- 标题 -->
      <el-table-column prop="title" label="标题" min-width="120" show-overflow-tooltip />

      <!-- 跳转链接 -->
      <el-table-column label="跳转链接" min-width="200">
        <template #default="{ row }">
          <span
            style="display: inline-block; max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; vertical-align: middle"
            :title="row.link_config?.target_name || row.link_url"
          >{{ row.link_config?.target_name || row.link_url || '—' }}</span>
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

    <div v-if="!banners.length && !loading" style="display: flex; justify-content: center; padding: 40px 0">
      <el-empty description="暂无 Banner" />
    </div>

    <!-- 新增/编辑 Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑 Banner' : '新增 Banner'"
      width="480px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" label-width="90px">
        <el-form-item label="图片" prop="image_url">
          <UploadImage v-model="form.image_url" />
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
            建议尺寸 750×300，支持 JPG / PNG
          </div>
        </el-form-item>
        <el-form-item label="标题" prop="title">
          <el-input v-model="form.title" placeholder="可选" maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item label="跳转链接" prop="link_url">
          <div style="width: 100%">
            <LinkPicker v-model="form.link_config" style="margin-bottom: 8px" />
            <el-input v-model="form.link_url" placeholder="可选，如 /pages/product/detail?id=1" />
          </div>
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

    <!-- 图片放大预览 -->
    <el-dialog v-model="previewVisible" title="图片预览" width="600px">
      <img :src="previewUrl" style="width: 100%; border-radius: 4px" />
    </el-dialog>
  </div>
</template>
