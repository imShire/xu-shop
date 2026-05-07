<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { getRoleList, createRole, updateRole, deleteRole, getPermissions } from '@/api/account'

const loading = ref(false)
const roles = ref<any[]>([])
const permissions = ref<any[]>([])
const permGroups = ref<Record<string, any[]>>({})

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const form = ref({
  id: '',
  name: '',
  code: '',
  perm_codes: [] as string[],
})

async function loadData() {
  loading.value = true
  try {
    const [roleRes, permRes] = await Promise.all([getRoleList(), getPermissions()])
    roles.value = roleRes ?? []
    permissions.value = permRes ?? []

    const groups: Record<string, any[]> = {}
    permRes.forEach((p: any) => {
      if (!groups[p.group]) groups[p.group] = []
      groups[p.group].push(p)
    })
    permGroups.value = groups
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = { id: '', name: '', code: '', perm_codes: [] }
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: any) {
  const permCodes = (row.permissions || []).map((p: any) => p.code)
  form.value = { id: row.id, name: row.name, code: row.code, perm_codes: permCodes }
  isEdit.value = true
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    if (isEdit.value) {
      await updateRole(form.value.id, form.value)
    } else {
      await createRole(form.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确认删除角色「${row.name}」？`, '提示', { type: 'warning' })
  await deleteRole(row.id)
  ElMessage.success('删除成功')
  await loadData()
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">角色权限</h3>
      <el-button type="primary" @click="openCreate">+ 新建角色</el-button>
    </div>

    <el-table v-loading="loading" :data="roles" border stripe>
      <el-table-column prop="name" label="角色名称" width="150" />
      <el-table-column prop="code" label="角色代码" width="150" />
      <el-table-column label="权限" min-width="200">
        <template #default="{ row }">
          <el-tag
            v-for="perm in (row.permissions || []).slice(0, 5)"
            :key="perm.code"
            size="small"
            style="margin-right: 4px; margin-bottom: 2px"
          >
            {{ perm.name }}
          </el-tag>
          <span v-if="(row.permissions || []).length > 5" style="font-size: 12px; color: var(--text-secondary)">
            +{{ row.permissions.length - 5 }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="600px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="角色名" prop="name" :rules="[{ required: true }]">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="代码" prop="code" :rules="[{ required: true }]">
          <el-input v-model="form.code" placeholder="如：admin" />
        </el-form-item>
        <el-form-item label="权限">
          <div style="max-height: 360px; overflow-y: auto; width: 100%">
            <div v-for="(perms, group) in permGroups" :key="group" style="margin-bottom: 12px">
              <div style="font-weight: 600; margin-bottom: 6px; color: var(--text-primary)">{{ group }}</div>
              <el-checkbox-group v-model="form.perm_codes" style="display: flex; flex-wrap: wrap; gap: 4px">
                <el-checkbox v-for="p in perms" :key="p.code" :value="p.code" style="margin: 0">
                  {{ p.name }}
                </el-checkbox>
              </el-checkbox-group>
            </div>
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
