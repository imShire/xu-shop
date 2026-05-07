<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getUserList, createUser, disableUser, enableUser } from '@/api/user'
import { rechargeBalance } from '@/api/balance'
import { formatTime, formatAmount } from '@/utils/format'

const router = useRouter()

const searchForm = ref({ phone: '', nickname: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getUserList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '用户', slot: 'user', minWidth: 160 },
  { label: '手机号', prop: 'phone', width: 130 },
  { label: '余额(元)', prop: 'balance_cents', width: 110, formatter: (r) => formatAmount(r.balance_cents ?? 0) },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '注册时间', prop: 'created_at', width: 150, formatter: (r) => formatTime(r.created_at) },
]

async function toggleStatus(row: any) {
  const action = row.status === 'active' ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确认${action}该用户？`, '提示', { type: 'warning' })
    if (row.status === 'active') {
      await disableUser(row.id)
    } else {
      await enableUser(row.id)
    }
    ElMessage.success(`${action}成功`)
    fetch(searchForm.value)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  }
}

// 余额充值
const rechargeDialogVisible = ref(false)
const rechargeForm = ref({ amount_yuan: '', remark: '' })
const currentUserId = ref('')
const rechargeLoading = ref(false)

function openRecharge(row: any) {
  currentUserId.value = String(row.id)
  rechargeForm.value = { amount_yuan: '', remark: '' }
  rechargeDialogVisible.value = true
}

async function confirmRecharge() {
  const amountYuan = parseFloat(rechargeForm.value.amount_yuan)
  if (isNaN(amountYuan) || amountYuan <= 0) {
    ElMessage.error('请输入有效金额')
    return
  }
  rechargeLoading.value = true
  try {
    await rechargeBalance(currentUserId.value, {
      amount_cents: Math.round(amountYuan * 100),
      remark: rechargeForm.value.remark || undefined,
    })
    ElMessage.success('充值成功')
    rechargeDialogVisible.value = false
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '充值失败')
  } finally {
    rechargeLoading.value = false
  }
}

// 新建用户弹窗
const dialogVisible = ref(false)
const formRef = ref()
const createForm = reactive({
  phone: '',
  password: '',
  nickname: '',
})
const createLoading = ref(false)

const phoneRule = [
  { required: true, message: '请输入手机号', trigger: 'blur' },
  { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' },
]
const passwordRule = [
  { required: true, message: '请输入密码', trigger: 'blur' },
  { min: 6, message: '密码至少6位', trigger: 'blur' },
]

function openCreateDialog() {
  createForm.phone = ''
  createForm.password = ''
  createForm.nickname = ''
  dialogVisible.value = true
}

async function handleCreate() {
  await formRef.value.validate()
  createLoading.value = true
  try {
    await createUser({
      phone: createForm.phone,
      password: createForm.password,
      nickname: createForm.nickname || undefined,
    })
    ElMessage.success('创建成功')
    dialogVisible.value = false
    fetch(searchForm.value)
  } catch (e: any) {
    ElMessage.error(e?.message || '创建失败')
  } finally {
    createLoading.value = false
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
      <template #toolbar>
        <el-button type="primary" @click="openCreateDialog">新建用户</el-button>
      </template>

      <template #search>
        <el-input v-model="searchForm.phone" placeholder="手机号" clearable style="width: 150px" />
        <el-input v-model="searchForm.nickname" placeholder="昵称" clearable style="width: 150px" />
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 100px">
          <el-option label="正常" value="active" />
          <el-option label="禁用" value="disabled" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #user="{ row }">
        <div style="display: flex; align-items: center; gap: 8px">
          <el-avatar :size="32" :src="row.avatar">{{ row.nickname?.charAt(0) }}</el-avatar>
          <span>{{ row.nickname || '未知' }}</span>
        </div>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
          {{ row.status === 'active' ? '正常' : '禁用' }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button
          text
          type="primary"
          size="small"
          @click="router.push(`/user/${row.id}`)"
        >
          查看
        </el-button>
        <el-button
          text
          type="warning"
          size="small"
          @click="openRecharge(row)"
        >
          充值
        </el-button>
        <el-button
          text
          :type="row.status === 'active' ? 'danger' : 'success'"
          size="small"
          @click="toggleStatus(row)"
        >
          {{ row.status === 'active' ? '禁用' : '启用' }}
        </el-button>
      </template>
    </ProTable>

    <!-- 余额充值弹窗 -->
    <el-dialog v-model="rechargeDialogVisible" title="余额充值" width="400px" :close-on-click-modal="false">
      <el-form label-width="80px">
        <el-form-item label="充值金额">
          <el-input v-model="rechargeForm.amount_yuan" placeholder="请输入金额（元）" type="number" min="0.01">
            <template #append>元</template>
          </el-input>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="rechargeForm.remark" placeholder="可选备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="rechargeLoading" @click="confirmRecharge">确认充值</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogVisible" title="新建用户" width="420px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="createForm" label-width="80px">
        <el-form-item label="手机号" prop="phone" :rules="phoneRule">
          <el-input v-model="createForm.phone" placeholder="请输入手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="密码" prop="password" :rules="passwordRule">
          <el-input v-model="createForm.password" type="password" placeholder="请输入密码（至少6位）" show-password />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="createForm.nickname" placeholder="选填，默认使用手机后4位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">确认创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>
