<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import {
  getFreightTemplates,
  createFreightTemplate,
  updateFreightTemplate,
  deleteFreightTemplate,
} from '@/api/order'
import PriceInput from '@/components/PriceInput/index.vue'

const loading = ref(false)
const templates = ref<any[]>([])

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const form = ref<{
  id: string
  name: string
  is_default: boolean
  free_threshold: number
  base_fee: number
  rules: { regions: string[]; extra_fee: number }[]
}>({
  id: '',
  name: '',
  is_default: false,
  free_threshold: 0,
  base_fee: 0,
  rules: [],
})

async function loadData() {
  loading.value = true
  try {
    templates.value = await getFreightTemplates() ?? []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = { id: '', name: '', is_default: false, free_threshold: 0, base_fee: 0, rules: [] }
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: any) {
  Object.assign(form.value, { ...row, rules: JSON.parse(JSON.stringify(row.rules || [])) })
  isEdit.value = true
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    if (isEdit.value) {
      await updateFreightTemplate(form.value.id, form.value)
    } else {
      await createFreightTemplate(form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm('确认删除该运费模板？', '提示', { type: 'warning' })
    await deleteFreightTemplate(row.id)
    ElMessage.success('删除成功')
    await loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
  }
}

function addRule() {
  form.value.rules.push({ regions: [], extra_fee: 0 })
}

function removeRule(idx: number) {
  form.value.rules.splice(idx, 1)
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">运费模板</h3>
      <el-button type="primary" @click="openCreate">+ 新建模板</el-button>
    </div>

    <el-table v-loading="loading" :data="templates" border stripe>
      <el-table-column prop="name" label="模板名称" min-width="150" />
      <el-table-column label="基础运费" width="120" align="right">
        <template #default="{ row }">¥{{ (row.base_fee / 100).toFixed(2) }}</template>
      </el-table-column>
      <el-table-column label="免运费门槛" width="130" align="right">
        <template #default="{ row }">
          {{ row.free_threshold ? `¥${(row.free_threshold / 100).toFixed(2)}` : '-' }}
        </template>
      </el-table-column>
      <el-table-column label="默认" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑模板' : '新建模板'" width="640px">
      <el-form ref="formRef" :model="form" label-width="100px">
        <el-form-item label="模板名称" prop="name" :rules="[{ required: true }]">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="基础运费">
          <PriceInput v-model="form.base_fee" />
        </el-form-item>
        <el-form-item label="免运费门槛">
          <PriceInput v-model="form.free_threshold" placeholder="0 表示不免运费" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="form.is_default" />
        </el-form-item>

        <el-form-item label="偏远地区">
          <div style="width: 100%">
            <div v-for="(rule, idx) in form.rules" :key="idx" style="display: flex; gap: 8px; margin-bottom: 8px; align-items: flex-start">
              <el-select
                v-model="rule.regions"
                multiple
                filterable
                allow-create
                placeholder="输入省份后回车"
                style="flex: 1"
              />
              <PriceInput v-model="rule.extra_fee" placeholder="附加运费" style="width: 150px" />
              <el-button text type="danger" @click="removeRule(idx)">删除</el-button>
            </div>
            <el-button size="small" @click="addRule">+ 添加偏远地区规则</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
