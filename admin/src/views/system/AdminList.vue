<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import {
  getAdminList,
  createAdmin,
  updateAdmin,
  disableAdmin,
  enableAdmin,
  resetAdminPwd,
  getRoleList,
} from '@/api/account'
import { formatTime } from '@/utils/format'

const searchForm = ref({ username: '', status: '' })
const roles = ref<any[]>([])

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getAdminList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '用户名', prop: 'username', width: 130 },
  { label: '真实姓名', prop: 'real_name', width: 120 },
  { label: '角色', slot: 'roles', minWidth: 150 },
  { label: '状态', slot: 'status', width: 80, align: 'center' },
  { label: '创建时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
  { label: '操作', slot: 'actions', width: 180 },
]

const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const pwdFormRef = ref<FormInstance>()

const form = ref({
  id: '',
  username: '',
  real_name: '',
  password: '',
  role_ids: [] as string[],
  status: 1,
})

const pwdDialog = ref({ visible: false, id: '', new_password: '' })
const pwdLoading = ref(false)

function validateStrongPassword(_: unknown, value: string, callback: (error?: Error) => void) {
  if (!value) {
    callback(new Error('请输入密码'))
    return
  }
  if (value.length < 12) {
    callback(new Error('密码至少12位'))
    return
  }
  const hasLetter = /[A-Za-z]/.test(value)
  const hasDigit = /\d/.test(value)
  const hasSpecial = /[^A-Za-z0-9]/.test(value)
  if (!hasLetter || !hasDigit || !hasSpecial) {
    callback(new Error('密码须包含字母、数字和特殊符号'))
    return
  }
  callback()
}

const passwordRules = [
  { required: true, validator: validateStrongPassword, trigger: 'blur' },
]

function isAdminActive(status: unknown) {
  return status === 'active'
}

function resolveRoleIds(roleCodes: string[] | undefined) {
  if (!Array.isArray(roleCodes) || roleCodes.length === 0) {
    return []
  }
  return roles.value
    .filter((role) => roleCodes.includes(role.code))
    .map((role) => role.id)
}

function openCreate() {
  form.value = { id: '', username: '', real_name: '', password: '', role_ids: [], status: 1 }
  isEdit.value = false
  dialogVisible.value = true
}

function openEdit(row: any) {
  form.value = {
    id: row.id,
    username: row.username,
    real_name: row.real_name || '',
    password: '',
    role_ids: resolveRoleIds(row.roles),
    status: row.status,
  }
  isEdit.value = true
  dialogVisible.value = true
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    const data = { ...form.value }
    if (isEdit.value) {
      await updateAdmin(form.value.id, data)
    } else {
      await createAdmin(data)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    fetch(searchForm.value)
  } finally {
    saving.value = false
  }
}

async function toggleStatus(row: any) {
  try {
    if (isAdminActive(row.status)) {
      await disableAdmin(row.id)
      ElMessage.success('已禁用')
    } else {
      await enableAdmin(row.id)
      ElMessage.success('已启用')
    }
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

function openResetPwd(row: any) {
  pwdDialog.value = { visible: true, id: row.id, new_password: '' }
}

async function handleResetPwd() {
  await pwdFormRef.value?.validate()
  pwdLoading.value = true
  try {
    await resetAdminPwd(pwdDialog.value.id, { new_password: pwdDialog.value.new_password })
    ElMessage.success('密码已重置')
    pwdDialog.value.visible = false
  } finally {
    pwdLoading.value = false
  }
}

onMounted(async () => {
  roles.value = await getRoleList().catch(() => [])
  fetch(searchForm.value)
})
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
        <el-input v-model="searchForm.username" placeholder="用户名" clearable style="width: 150px" />
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button type="primary" v-permission="'system.admin.edit'" @click="openCreate">+ 新建员工</el-button>
      </template>

      <template #roles="{ row }">
        <el-tag v-for="r in row.roles || []" :key="r" size="small" style="margin-right: 4px">{{ r }}</el-tag>
      </template>

      <template #status="{ row }">
        <el-tag :type="isAdminActive(row.status) ? 'success' : 'danger'" size="small">
          {{ isAdminActive(row.status) ? '正常' : '禁用' }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
        <el-button
          text
          :type="isAdminActive(row.status) ? 'danger' : 'success'"
          size="small"
          @click="toggleStatus(row)"
        >
          {{ isAdminActive(row.status) ? '禁用' : '启用' }}
        </el-button>
        <el-button text size="small" @click="openResetPwd(row)">重置密码</el-button>
      </template>
    </ProTable>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑员工' : '新建员工'" width="480px">
      <el-form ref="formRef" :model="form" label-width="90px">
        <el-form-item label="用户名" prop="username" :rules="[{ required: true }]">
          <el-input v-model="form.username" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input v-model="form.real_name" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password" :rules="passwordRules">
          <el-input v-model="form.password" type="password" show-password placeholder="至少12位，含字母、数字、特殊符号" />
        </el-form-item>
        <el-form-item label="角色">
          <el-checkbox-group v-model="form.role_ids">
            <el-checkbox v-for="role in roles" :key="role.id" :value="role.id">{{ role.name }}</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="pwdDialog.visible" title="重置密码" width="360px">
      <el-form ref="pwdFormRef" :model="pwdDialog" label-width="90px">
        <el-form-item prop="new_password" :rules="passwordRules">
          <el-input
            v-model="pwdDialog.new_password"
            type="password"
            show-password
            placeholder="至少12位，含字母、数字、特殊符号"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="pwdLoading" @click="handleResetPwd">确认重置</el-button>
      </template>
    </el-dialog>
  </div>
</template>
