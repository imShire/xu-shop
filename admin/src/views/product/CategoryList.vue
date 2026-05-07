<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { getCategoryList, createCategory, updateCategory, deleteCategory } from '@/api/product'
import UploadImage from '@/components/UploadImage/index.vue'

const loading = ref(false)
const categories = ref<any[]>([])

const dialogVisible = ref(false)
const dialogTitle = ref('新建分类')
const formRef = ref<FormInstance>()
const saving = ref(false)

interface CategoryForm {
  id: string
  name: string
  parent_id?: string
  sort: number
  icon: string
  status: 'enabled' | 'disabled'
}

function createDefaultForm(): CategoryForm {
  return {
    id: '',
    name: '',
    parent_id: undefined,
    sort: 0,
    icon: '',
    status: 'enabled',
  }
}

function normalizeParentId(value: unknown): string {
  if (typeof value === 'string' && value !== '' && value !== '0') return value
  return '0'
}

const form = ref<CategoryForm>(createDefaultForm())

async function loadData() {
  loading.value = true
  try {
    categories.value = await getCategoryList() ?? []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = createDefaultForm()
  dialogTitle.value = '新建分类'
  dialogVisible.value = true
}

function openEdit(row: any) {
  const parentId = normalizeParentId(row.parent_id)
  form.value = {
    id: row.id,
    name: row.name,
    parent_id: parentId !== '0' ? parentId : undefined,
    sort: row.sort || 0,
    icon: row.icon || '',
    status: row.status || 'enabled',
  }
  dialogTitle.value = '编辑分类'
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const payload = {
      ...form.value,
      parent_id: normalizeParentId(form.value.parent_id),
    }
    if (form.value.id) {
      await updateCategory(form.value.id, payload)
    } else {
      await createCategory(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确认删除分类「${row.name}」？`, '提示', { type: 'warning' })
  await deleteCategory(row.id)
  ElMessage.success('删除成功')
  await loadData()
}

onMounted(loadData)

// 只有顶级分类（parent_id 为 '0' / null / undefined）才能被选为父级
const rootCategories = computed(() =>
  categories.value.filter((c: any) => !c.parent_id || c.parent_id === '0')
)

// id → name 映射，用于列表中显示父级名称
const categoryNameMap = computed(() => {
  const map: Record<string, string> = {}
  categories.value.forEach((c: any) => {
    map[c.id] = c.name
  })
  return map
})
</script>

<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">商品分类</h3>
      <el-button type="primary" @click="openCreate">+ 新建分类</el-button>
    </div>

    <el-table
      v-loading="loading"
      :data="categories"
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      border
    >
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column label="图标" width="70" align="center">
        <template #default="{ row }">
          <el-image
            v-if="row.icon"
            :src="row.icon"
            style="width: 36px; height: 36px; border-radius: 4px"
            fit="cover"
          />
          <span v-else style="color: var(--el-text-color-placeholder); font-size: 12px">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="分类名称" min-width="150" />
      <el-table-column label="父级分类" width="140">
        <template #default="{ row }">
          <span v-if="row.parent_id && row.parent_id !== '0' && categoryNameMap[row.parent_id]">
            {{ categoryNameMap[row.parent_id] }}
          </span>
          <span v-else style="color: var(--el-text-color-placeholder)">—</span>
        </template>
      </el-table-column>
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="420px" destroy-on-close>
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="分类名" prop="name" :rules="[{ required: true }]">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="图标">
          <div>
            <UploadImage v-model="form.icon" />
            <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-top: 4px">
              建议 128×128 正方形图片，支持 JPG / PNG
            </div>
          </div>
        </el-form-item>
        <el-form-item label="父分类">
          <el-select v-model="form.parent_id" clearable placeholder="顶级分类">
            <el-option
              v-for="cat in rootCategories"
              :key="cat.id"
              :label="cat.name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="enabled">启用</el-radio>
            <el-radio value="disabled">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
