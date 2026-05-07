<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getChannelCodeList, createChannelCode, updateChannelCode, deleteChannelCode } from '@/api/private-domain'
import { formatTime } from '@/utils/format'

const searchForm = ref({ name: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getChannelCodeList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '渠道名', prop: 'name', minWidth: 150 },
  { label: '二维码', slot: 'qrcode', width: 80, align: 'center' },
  { label: '扫码数', prop: 'scan_count', width: 90, align: 'right' },
  { label: '加粉数', prop: 'add_count', width: 90, align: 'right' },
  { label: '下单数', prop: 'order_count', width: 90, align: 'right' },
  { label: '创建时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

const createDialogVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const form = ref({ name: '' })

const editDialogVisible = ref(false)
const editSaving = ref(false)
const editFormRef = ref<FormInstance>()
const editForm = ref<{ id: string; name: string; remark: string }>({ id: '', name: '', remark: '' })

function openCreate() {
  form.value = { name: '' }
  createDialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    await createChannelCode(form.value)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

function openEdit(row: any) {
  editForm.value = { id: row.id, name: row.name, remark: row.remark || '' }
  editDialogVisible.value = true
}

async function handleEditSave() {
  await editFormRef.value?.validate()
  editSaving.value = true
  try {
    await updateChannelCode(editForm.value.id, { name: editForm.value.name, remark: editForm.value.remark })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    editSaving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确认删除渠道码「${row.name}」？`, '提示', { type: 'warning' })
    await deleteChannelCode(row.id)
    ElMessage.success('删除成功')
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(() => fetch(searchForm.value))
</script>

<template>
  <div class="page-card">
    <ProTable
      :columns="columns"
      :data="list"
      :total="total"
      :loading="loading"
      v-model:page="page"
      v-model:page-size="pageSize"
      @refresh="fetch(searchForm)"
    >
      <template #search>
        <el-input v-model="searchForm.name" placeholder="渠道名" clearable style="width: 150px" />
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button type="primary" @click="openCreate">+ 生成渠道码</el-button>
      </template>

      <template #qrcode="{ row }">
        <el-image
          v-if="row.qr_image_url"
          :src="row.qr_image_url"
          style="width: 40px; height: 40px"
          :preview-src-list="[row.qr_image_url]"
          fit="cover"
        />
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
        <el-button text type="danger" size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="createDialogVisible" title="生成渠道码" width="400px">
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="渠道名" prop="name" :rules="[{ required: true }]">
          <el-input v-model="form.name" placeholder="如：抖音广告" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">生成</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="editDialogVisible" title="编辑渠道码" width="400px">
      <el-form ref="editFormRef" :model="editForm" label-width="80px">
        <el-form-item label="渠道名" prop="name" :rules="[{ required: true }]">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSaving" @click="handleEditSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
