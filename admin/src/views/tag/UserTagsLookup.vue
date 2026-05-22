<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getUserTags,
  grantUserTag,
  revokeUserTag,
  listUserTags,
  tagCategoryLabels,
  type UserTagBinding,
  type UserTag,
  type TagCategory,
} from '@/api/tag'
import { formatTime } from '@/utils/format'

const userId = ref('')
const tags = ref<UserTagBinding[]>([])
const loading = ref(false)

const grantVisible = ref(false)
const allTags = ref<UserTag[]>([])
const grantForm = ref<{ tag_code: string; expire_at: string | null }>({ tag_code: '', expire_at: null })

async function search() {
  if (!userId.value.trim()) {
    ElMessage.warning('请输入用户 ID')
    return
  }
  loading.value = true
  try {
    tags.value = (await getUserTags(userId.value.trim())) || []
  } finally {
    loading.value = false
  }
}

async function openGrant() {
  if (!userId.value.trim()) {
    ElMessage.warning('请先查询用户')
    return
  }
  if (allTags.value.length === 0) {
    allTags.value = ((await listUserTags({ source: 'manual' })) || []).filter((t) => t.enabled)
  }
  grantForm.value = { tag_code: '', expire_at: null }
  grantVisible.value = true
}

async function submitGrant() {
  if (!grantForm.value.tag_code) {
    ElMessage.warning('请选择标签')
    return
  }
  await grantUserTag(userId.value.trim(), {
    tag_code: grantForm.value.tag_code,
    expire_at: grantForm.value.expire_at,
  })
  ElMessage.success('已添加')
  grantVisible.value = false
  search()
}

async function remove(t: UserTagBinding) {
  if (t.source === 'auto') {
    ElMessage.warning('自动标签由系统计算，不能手动移除')
    return
  }
  await ElMessageBox.confirm(`确认移除标签「${t.tag_name}」？`, '移除', { type: 'warning' })
  await revokeUserTag(userId.value.trim(), t.tag_code)
  ElMessage.success('已移除')
  search()
}
</script>

<template>
  <div class="page-card">
    <div style="display: flex; gap: 8px; margin-bottom: 16px">
      <el-input v-model="userId" placeholder="输入用户 ID" style="width: 280px" @keyup.enter="search" />
      <el-button type="primary" @click="search">查询</el-button>
      <el-button :disabled="!userId" @click="openGrant">添加手动标签</el-button>
    </div>

    <el-table :data="tags" border v-loading="loading">
      <el-table-column prop="tag_code" label="代码" width="200" />
      <el-table-column prop="tag_name" label="名称" width="180" />
      <el-table-column label="类别" width="120">
        <template #default="{ row }">{{ tagCategoryLabels[row.category as TagCategory] || row.category }}</template>
      </el-table-column>
      <el-table-column label="来源" width="100">
        <template #default="{ row }">
          <el-tag size="small" :type="row.source === 'auto' ? 'info' : 'success'">
            {{ row.source === 'auto' ? '自动' : '手动' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="score" label="分数" width="100" align="right" />
      <el-table-column label="授予时间" width="160">
        <template #default="{ row }">{{ formatTime(row.granted_at) }}</template>
      </el-table-column>
      <el-table-column label="过期时间" width="160">
        <template #default="{ row }">{{ formatTime(row.expire_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.source === 'manual'" text type="danger" size="small" @click="remove(row)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="grantVisible" title="添加手动标签" width="480px">
      <el-form label-width="100px">
        <el-form-item label="标签" required>
          <el-select v-model="grantForm.tag_code" filterable style="width: 100%">
            <el-option
              v-for="t in allTags"
              :key="t.code"
              :label="`[${tagCategoryLabels[t.category] || t.category}] ${t.name}`"
              :value="t.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="过期时间">
          <el-date-picker
            v-model="grantForm.expire_at"
            type="datetime"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="留空表示长期有效"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="primary" @click="submitGrant">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>
