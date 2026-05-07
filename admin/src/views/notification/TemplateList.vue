<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { getTemplateList, updateTemplate, testSendNotification } from '@/api/notification'
import { formatTime } from '@/utils/format'

interface NotificationTemplateItem {
  id: string
  code: string
  channel: string
  template_id_external: string
  fields: Record<string, unknown>
  enabled: boolean
  created_at: string
  updated_at: string
}

interface EditForm {
  template_id_external: string
  fields_text: string
  enabled: boolean
}

const loading = ref(false)
const templates = ref<NotificationTemplateItem[]>([])

const editDialog = ref({ visible: false, code: '' })
const editForm = ref<EditForm>({ template_id_external: '', fields_text: '{}', enabled: true })
const editFormRef = ref<FormInstance>()
const saving = ref(false)

const testDialog = ref({ visible: false, code: '', openid: '' })
const testLoading = ref(false)

function stringifyFields(fields: Record<string, unknown> | null | undefined) {
  return JSON.stringify(fields || {}, null, 2)
}

function parseFields(text: string) {
  const normalized = text.trim() || '{}'
  const parsed = JSON.parse(normalized)
  if (parsed === null || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error('字段映射必须是 JSON 对象')
  }
  return parsed as Record<string, unknown>
}

async function loadData() {
  loading.value = true
  try {
    templates.value = await getTemplateList() ?? []
  } finally {
    loading.value = false
  }
}

function openEdit(row: NotificationTemplateItem) {
  editDialog.value = { visible: true, code: row.code }
  editForm.value = {
    template_id_external: row.template_id_external || '',
    fields_text: stringifyFields(row.fields),
    enabled: row.enabled,
  }
}

async function handleSave() {
  await editFormRef.value?.validate()
  let fields: Record<string, unknown>
  try {
    fields = parseFields(editForm.value.fields_text)
  } catch (error: any) {
    ElMessage.error(error?.message || '字段映射格式错误')
    return
  }
  saving.value = true
  try {
    await updateTemplate(editDialog.value.code, {
      template_id_external: editForm.value.template_id_external.trim(),
      fields,
      enabled: editForm.value.enabled,
    })
    ElMessage.success('保存成功')
    editDialog.value.visible = false
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

function openTest(row: NotificationTemplateItem) {
  testDialog.value = { visible: true, code: row.code, openid: '' }
}

async function handleTest() {
  if (!testDialog.value.openid) return ElMessage.warning('请输入测试目标')
  testLoading.value = true
  try {
    await testSendNotification(testDialog.value.code, { openid: testDialog.value.openid })
    ElMessage.success('测试消息已发送')
    testDialog.value.visible = false
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    testLoading.value = false
  }
}

async function toggleEnabled(row: NotificationTemplateItem, nextEnabled: boolean) {
  const previousEnabled = row.enabled
  row.enabled = nextEnabled
  try {
    await updateTemplate(row.code, { enabled: row.enabled })
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
  } catch (e: any) {
    row.enabled = previousEnabled
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">通知模板</h3>
    </div>

    <el-table v-loading="loading" :data="templates" border stripe>
      <el-table-column prop="code" label="模板编码" min-width="160" />
      <el-table-column prop="channel" label="渠道" width="120" />
      <el-table-column prop="template_id_external" label="外部模板ID" min-width="200" />
      <el-table-column label="字段映射" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">
          <code>{{ stringifyFields(row.fields) }}</code>
        </template>
      </el-table-column>
      <el-table-column label="启用" width="80" align="center">
        <template #default="{ row }">
          <el-switch :model-value="row.enabled" @change="(value: boolean) => toggleEnabled(row, value)" />
        </template>
      </el-table-column>
      <el-table-column label="更新时间" width="150" :formatter="(row: any) => formatTime(row.updated_at)" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button text size="small" @click="openTest(row)">测试</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="editDialog.visible" title="编辑模板" width="640px">
      <el-form ref="editFormRef" :model="editForm" label-width="100px">
        <el-form-item label="外部模板ID">
          <el-input v-model="editForm.template_id_external" placeholder="对接平台的模板ID" />
        </el-form-item>
        <el-form-item
          label="字段映射"
          prop="fields_text"
          :rules="[{ required: true, message: '请输入字段映射 JSON', trigger: 'blur' }]"
        >
          <el-input v-model="editForm.fields_text" type="textarea" :rows="8" placeholder='{"thing1":"content"}' />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="editForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="testDialog.visible" title="测试发送" width="400px">
      <el-form label-width="80px">
        <el-form-item label="目标">
          <el-input v-model="testDialog.openid" placeholder="用户 openid" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="testDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="testLoading" @click="handleTest">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>
