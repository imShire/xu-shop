<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import { getArticles, deleteArticle } from '@/api/cms'
import { formatTime } from '@/utils/format'

const router = useRouter()
const searchForm = ref({ keyword: '', status: '' })

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getArticles({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: 'ID', prop: 'id', width: 130 },
  { label: '标题', prop: 'title', minWidth: 200 },
  { label: '状态', slot: 'status', width: 90, align: 'center' },
  { label: '排序', prop: 'sort', width: 70, align: 'center' },
  { label: '创建时间', prop: 'created_at', width: 160, formatter: (r) => formatTime(r.created_at) },
]

async function handleDelete(row: { id: string; title: string }) {
  await ElMessageBox.confirm(`确认删除文章「${row.title}」？`, '确认删除', { type: 'warning' })
  try {
    await deleteArticle(row.id)
    ElMessage.success('已删除')
    fetch(searchForm.value)
  } catch {
    ElMessage.error('删除失败')
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
        <el-input v-model="searchForm.keyword" placeholder="搜索标题" clearable style="width: 200px" />
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 120px">
          <el-option label="草稿" value="draft" />
          <el-option label="已发布" value="published" />
        </el-select>
        <el-button type="primary" @click="fetch(searchForm)">搜索</el-button>
      </template>

      <template #toolbar>
        <el-button type="primary" @click="router.push('/cms/articles/create')">新增文章</el-button>
      </template>

      <template #status="{ row }">
        <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">
          {{ row.status === 'published' ? '已发布' : '草稿' }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="router.push(`/cms/articles/edit/${row.id}`)">编辑</el-button>
        <el-button text type="danger" size="small" @click="handleDelete(row)">删除</el-button>
      </template>
    </ProTable>
  </div>
</template>
