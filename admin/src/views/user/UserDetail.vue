<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserDetail, disableUser, enableUser } from '@/api/user'
import { getOrderList } from '@/api/order'
import { getUserBalanceLogs, rechargeBalance } from '@/api/balance'
import { formatAmount, formatTime, orderStatusMap } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const userId = String(route.params.id)

const loading = ref(false)
const user = ref<any>(null)
const actionLoading = ref(false)

const ordersLoading = ref(false)
const orders = ref<any[]>([])

const balanceLoading = ref(false)
const balanceCents = ref(0)
const balanceLogs = ref<any[]>([])
const balanceTotal = ref(0)
const balancePage = ref(1)

const rechargeDialogVisible = ref(false)
const rechargeForm = ref({ amount_yuan: '', remark: '' })
const rechargeLoading = ref(false)

const balanceTypeMap: Record<string, string> = {
  recharge: '充值',
  consume: '消费',
  refund: '退款',
  adjust: '调整',
}

async function loadUser() {
  loading.value = true
  try {
    user.value = await getUserDetail(userId)
    balanceCents.value = user.value?.balance_cents ?? 0
  } finally {
    loading.value = false
  }
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const res = await getOrderList({ user_id: userId, page: 1, page_size: 10 } as any)
    orders.value = res?.list ?? res ?? []
  } catch {
    orders.value = []
  } finally {
    ordersLoading.value = false
  }
}

async function loadBalanceLogs(page = 1) {
  balanceLoading.value = true
  try {
    const res = await getUserBalanceLogs(userId, { page, page_size: 20 })
    balanceCents.value = res.balance_cents
    balanceLogs.value = res.list ?? []
    balanceTotal.value = res.total
    balancePage.value = page
  } catch {
    balanceLogs.value = []
  } finally {
    balanceLoading.value = false
  }
}

async function handleToggleStatus() {
  if (!user.value) return
  const isActive = user.value.status === 'active'
  const action = isActive ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确认${action}该用户？`, '提示', { type: 'warning' })
    actionLoading.value = true
    if (isActive) {
      await disableUser(userId)
    } else {
      await enableUser(userId)
    }
    ElMessage.success(`${action}成功`)
    await loadUser()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  } finally {
    actionLoading.value = false
  }
}

async function confirmRecharge() {
  const amountYuan = parseFloat(rechargeForm.value.amount_yuan)
  if (isNaN(amountYuan) || amountYuan <= 0) {
    ElMessage.error('请输入有效金额')
    return
  }
  rechargeLoading.value = true
  try {
    await rechargeBalance(userId, {
      amount_cents: Math.round(amountYuan * 100),
      remark: rechargeForm.value.remark || undefined,
    })
    ElMessage.success('充值成功')
    rechargeDialogVisible.value = false
    await loadUser()
    await loadBalanceLogs(1)
  } catch (e: any) {
    ElMessage.error(e?.message || '充值失败')
  } finally {
    rechargeLoading.value = false
  }
}

onMounted(async () => {
  await loadUser()
  loadOrders()
  loadBalanceLogs()
})
</script>

<template>
  <div v-loading="loading">
    <div style="display: flex; align-items: center; gap: 16px; margin-bottom: 16px">
      <el-button text :icon="'ArrowLeft'" @click="router.push('/user/list')">返回用户列表</el-button>
      <h3 style="font-size: 16px; font-weight: 600">用户详情</h3>
    </div>

    <template v-if="user">
      <!-- 用户信息卡片 -->
      <el-card style="margin-bottom: 16px">
        <div style="display: flex; align-items: flex-start; gap: 24px">
          <el-avatar :size="72" :src="user.avatar">
            {{ user.nickname?.charAt(0) || '?' }}
          </el-avatar>
          <div style="flex: 1">
            <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 12px">
              <span style="font-size: 18px; font-weight: 600">{{ user.nickname || '未设置昵称' }}</span>
              <el-tag :type="user.status === 'active' ? 'success' : 'danger'" size="small">
                {{ user.status === 'active' ? '正常' : '禁用' }}
              </el-tag>
            </div>
            <el-descriptions :column="3" size="small">
              <el-descriptions-item label="手机号">{{ user.phone || '-' }}</el-descriptions-item>
              <el-descriptions-item label="用户 ID">{{ user.id }}</el-descriptions-item>
              <el-descriptions-item label="注册时间">{{ formatTime(user.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="账户余额">
                <span style="color: #e6a23c; font-weight: 600">{{ formatAmount(balanceCents) }}</span>
              </el-descriptions-item>
            </el-descriptions>
          </div>
          <div style="display: flex; flex-direction: column; gap: 8px">
            <el-button
              type="warning"
              @click="rechargeDialogVisible = true"
            >
              余额充值
            </el-button>
            <el-button
              :type="user.status === 'active' ? 'danger' : 'success'"
              :loading="actionLoading"
              @click="handleToggleStatus"
            >
              {{ user.status === 'active' ? '禁用用户' : '启用用户' }}
            </el-button>
          </div>
        </div>
      </el-card>

      <!-- 标签页 -->
      <el-card>
        <el-tabs>
          <!-- 订单记录 -->
          <el-tab-pane label="订单记录" name="orders">
            <el-table v-loading="ordersLoading" :data="orders" style="width: 100%">
              <el-table-column label="订单号" prop="order_no" min-width="160" />
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="orderStatusMap[row.status]?.type || ''" size="small">
                    {{ orderStatusMap[row.status]?.label || row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="实付金额" width="110">
                <template #default="{ row }">
                  {{ formatAmount(row.pay_amount ?? row.amount_cents ?? 0) }}
                </template>
              </el-table-column>
              <el-table-column label="下单时间" width="160">
                <template #default="{ row }">
                  {{ formatTime(row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="80" fixed="right">
                <template #default="{ row }">
                  <el-button text type="primary" size="small" @click="router.push(`/order/detail/${row.id}`)">
                    查看
                  </el-button>
                </template>
              </el-table-column>
              <template #empty>
                <el-empty description="暂无订单记录" :image-size="60" />
              </template>
            </el-table>
          </el-tab-pane>

          <!-- 充值记录 -->
          <el-tab-pane label="充值记录" name="balance">
            <el-table v-loading="balanceLoading" :data="balanceLogs" style="width: 100%">
              <el-table-column label="类型" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.change_cents > 0 ? 'success' : 'danger'" size="small">
                    {{ balanceTypeMap[row.type] || row.type }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="变动金额" width="110">
                <template #default="{ row }">
                  <span :style="{ color: row.change_cents > 0 ? '#67c23a' : '#f56c6c', fontWeight: '600' }">
                    {{ row.change_cents > 0 ? '+' : '' }}{{ formatAmount(row.change_cents) }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column label="变动前" width="110">
                <template #default="{ row }">{{ formatAmount(row.balance_before_cents) }}</template>
              </el-table-column>
              <el-table-column label="变动后" width="110">
                <template #default="{ row }">{{ formatAmount(row.balance_after_cents) }}</template>
              </el-table-column>
              <el-table-column label="备注" prop="remark" min-width="120" />
              <el-table-column label="时间" width="160">
                <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
              </el-table-column>
              <template #empty>
                <el-empty description="暂无余额记录" :image-size="60" />
              </template>
            </el-table>
            <div v-if="balanceTotal > 20" style="margin-top: 12px; display: flex; justify-content: flex-end">
              <el-pagination
                :current-page="balancePage"
                :page-size="20"
                :total="balanceTotal"
                layout="prev, pager, next"
                @current-change="loadBalanceLogs"
              />
            </div>
          </el-tab-pane>

          <!-- 收货地址 -->
          <el-tab-pane label="收货地址" name="addresses">
            <el-empty description="暂无数据" :image-size="60" />
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </template>

    <template v-else-if="!loading">
      <el-empty description="用户不存在或已被删除" />
    </template>

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
  </div>
</template>

