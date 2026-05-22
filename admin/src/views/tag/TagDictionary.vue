<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listUserTags,
  createUserTag,
  updateUserTag,
  deleteUserTag,
  tagCategoryLabels,
  type UserTag,
  type TagCategory,
} from '@/api/tag'

const allTags = ref<UserTag[]>([])
const loading = ref(false)
const activeCategory = ref<TagCategory | 'all'>('all')

const dialogVisible = ref(false)
const editing = ref<UserTag | null>(null)
const isCreate = ref(false)

const form = ref<UserTag>({
  code: '',
  name: '',
  category: 'business',
  parent_code: null,
  color: '#409EFF',
  description: '',
  source: 'manual',
  config: {},
  enabled: true,
  sort: 0,
})

const categories: Array<{ key: TagCategory | 'all'; label: string }> = [
  { key: 'all', label: '全部' },
  ...(Object.entries(tagCategoryLabels) as Array<[TagCategory, string]>).map(([key, label]) => ({ key, label })),
]

const filtered = computed(() => {
  if (activeCategory.value === 'all') return allTags.value
  return allTags.value.filter((t) => t.category === activeCategory.value)
})

async function load() {
  loading.value = true
  try {
    allTags.value = (await listUserTags()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  isCreate.value = true
  editing.value = null
  form.value = {
    code: '',
    name: '',
    category: activeCategory.value === 'all' ? 'business' : activeCategory.value,
    parent_code: null,
    color: '#409EFF',
    description: '',
    source: 'manual',
    config: {},
    enabled: true,
    sort: 0,
  }
  dialogVisible.value = true
}

function openEdit(t: UserTag) {
  isCreate.value = false
  editing.value = t
  form.value = { ...t, config: { ...(t.config || {}) } }
  dialogVisible.value = true
}

async function submit() {
  if (!form.value.code || !form.value.name) {
    ElMessage.warning('请填写代码和名称')
    return
  }
  if (isCreate.value) {
    await createUserTag(form.value)
    ElMessage.success('已创建')
  } else {
    await updateUserTag(form.value.code, form.value)
    ElMessage.success('已更新')
  }
  dialogVisible.value = false
  load()
}

async function remove(t: UserTag) {
  if (t.source === 'auto') {
    ElMessage.warning('自动标签不可删除')
    return
  }
  await ElMessageBox.confirm(`确认停用标签「${t.name}」？仅在无用户绑定时才会真正删除`, '停用', { type: 'warning' })
  await deleteUserTag(t.code)
  ElMessage.success('已停用')
  load()
}

onMounted(load)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px">
      <el-radio-group v-model="activeCategory" size="default">
        <el-radio-button v-for="c in categories" :key="c.key" :value="c.key">{{ c.label }}</el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="openCreate">新增业务标签</el-button>
    </div>

    <el-table :data="filtered" border>
      <el-table-column prop="code" label="代码" width="180" />
      <el-table-column label="名称" min-width="140">
        <template #default="{ row }">
          <el-tag :color="row.color" effect="dark" size="small" disable-transitions>{{ row.name }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="类别" width="120">
        <template #default="{ row }">{{ tagCategoryLabels[row.category as TagCategory] || row.category }}</template>
      </el-table-column>
      <el-table-column prop="source" label="来源" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="row.source === 'auto' ? 'info' : 'success'">
            {{ row.source === 'auto' ? '自动' : '手动' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="说明" min-width="200" show-overflow-tooltip />
      <el-table-column prop="sort" label="排序" width="80" align="center" />
      <el-table-column label="启用" width="80" align="center">
        <template #default="{ row }">
          <el-tag size="small" :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? '是' : '否' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.source !== 'auto'" text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-else text type="primary" size="small" @click="openEdit(row)">查看</el-button>
          <el-button v-if="row.source !== 'auto'" text type="danger" size="small" @click="remove(row)">停用</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isCreate ? '新增标签' : '编辑标签'" width="560px">
      <el-form label-width="100px" :disabled="form.source === 'auto'">
        <el-form-item label="代码" required>
          <el-input v-model="form.code" :disabled="!isCreate" maxlength="64" placeholder="字母/数字/_，建议加分类前缀" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="类别" required>
          <el-select v-model="form.category" style="width: 100%">
            <el-option
              v-for="(label, key) in tagCategoryLabels"
              :key="key"
              :label="label"
              :value="key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="父标签">
          <el-input v-model="form.parent_code" placeholder="可选，标签代码" />
        </el-form-item>
        <el-form-item label="颜色">
          <el-color-picker v-model="form.color" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button v-if="form.source !== 'auto'" type="primary" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
