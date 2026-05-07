<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import {
  getProductList,
  putOnSale,
  putOffSale,
  copyProduct,
  deleteProduct,
  batchOnSale,
  batchOffSale,
} from '@/api/product'
import { getCategoryList } from '@/api/product'
import { formatAmount, formatTime, productStatusMap } from '@/utils/format'

const router = useRouter()

const searchForm = ref({
  title: '',
  category_id: undefined as number | undefined,
  status: '' as string,
})

const categories = ref<any[]>([])
const selectedRows = ref<any[]>([])

const { list, total, page, pageSize, loading, fetch } = useTable((params) =>
  getProductList({ ...params, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '商品', slot: 'product', minWidth: 200 },
  { label: '分类', prop: 'category_name', width: 100 },
  {
    label: '价格',
    slot: 'price',
    width: 140,
  },
  { label: '库存', prop: 'total_stock', width: 80 },
  { label: '销量', slot: 'sales', width: 80 },
  { label: '运费', slot: 'freight', width: 100 },
  {
    label: '状态',
    slot: 'status',
    width: 90,
    align: 'center',
  },
  {
    label: '创建时间',
    prop: 'created_at',
    width: 150,
    formatter: (row) => formatTime(row.created_at),
  },
]

async function handleOnSale(row: any) {
  await putOnSale(row.id)
  ElMessage.success('已上架')
  fetch(searchForm.value)
}

async function handleOffSale(row: any) {
  await putOffSale(row.id)
  ElMessage.success('已下架')
  fetch(searchForm.value)
}

async function handleCopy(row: any) {
  await copyProduct(row.id)
  ElMessage.success('复制成功')
  fetch(searchForm.value)
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm(`确认删除「${row.title}」？`, '提示', { type: 'warning' })
  await deleteProduct(row.id)
  ElMessage.success('删除成功')
  fetch(searchForm.value)
}

async function handleBatchOnSale() {
  if (!selectedRows.value.length) return ElMessage.warning('请先选择商品')
  await batchOnSale(selectedRows.value.map((r) => r.id))
  ElMessage.success('批量上架成功')
  fetch(searchForm.value)
}

async function handleBatchOffSale() {
  if (!selectedRows.value.length) return ElMessage.warning('请先选择商品')
  await batchOffSale(selectedRows.value.map((r) => r.id))
  ElMessage.success('批量下架成功')
  fetch(searchForm.value)
}

function handleSearch() {
  page.value = 1
  fetch(searchForm.value)
}

function handleReset() {
  searchForm.value = { title: '', category_id: undefined, status: '' }
  handleSearch()
}

onMounted(async () => {
  const cats = await getCategoryList().catch(() => [])
  categories.value = cats
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
      selection
      @refresh="fetch(searchForm)"
      @selection-change="(rows) => (selectedRows = rows)"
    >
      <template #search>
        <el-input
          v-model="searchForm.title"
          placeholder="商品名称"
          clearable
          style="width: 180px"
          @keyup.enter="handleSearch"
        />
        <el-select
          v-model="searchForm.category_id"
          placeholder="商品分类"
          clearable
          style="width: 150px"
        >
          <el-option
            v-for="cat in categories"
            :key="cat.id"
            :label="cat.name"
            :value="cat.id"
          />
        </el-select>
        <el-select
          v-model="searchForm.status"
          placeholder="商品状态"
          clearable
          style="width: 120px"
        >
          <el-option label="草稿" value="draft" />
          <el-option label="在售" value="onsale" />
          <el-option label="下架" value="offsale" />
        </el-select>
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-button @click="handleReset">重置</el-button>
      </template>

      <template #toolbar>
        <div style="display: flex; gap: 8px">
          <el-button type="primary" @click="router.push('/product/edit')">
            + 新建商品
          </el-button>
          <el-button @click="handleBatchOnSale">批量上架</el-button>
          <el-button @click="handleBatchOffSale">批量下架</el-button>
        </div>
      </template>

      <template #product="{ row }">
        <div style="display: flex; align-items: center; gap: 8px">
          <img
            v-if="row.main_image"
            :src="row.main_image"
            style="width: 40px; height: 40px; border-radius: 4px; object-fit: cover; flex-shrink: 0"
          />
          <div style="min-width: 0">
            <div class="text-truncate" style="font-weight: 500">{{ row.title }}</div>
            <div class="text-truncate" style="font-size: 12px; color: var(--text-secondary)">
              {{ row.subtitle }}
            </div>
          </div>
        </div>
      </template>

      <template #price="{ row }">
        <div>
          <span style="font-weight: 600; color: #f59e0b">{{ formatAmount(row.price_min_cents) }}</span>
          <span v-if="row.price_max_cents !== row.price_min_cents"> ~ {{ formatAmount(row.price_max_cents) }}</span>
        </div>
      </template>

      <template #sales="{ row }">
        {{ (row.virtual_sales || 0) + (row.sales || 0) }}
      </template>

      <template #freight="{ row }">
        <el-tag v-if="!row.freight_template_name" type="success" size="small">包邮</el-tag>
        <span v-else>{{ row.freight_template_name }}</span>
      </template>

      <template #status="{ row }">
        <el-tag :type="productStatusMap[row.status]?.type || ''" size="small">
          {{ productStatusMap[row.status]?.label || row.status }}
        </el-tag>
      </template>

      <template #actions="{ row }">
        <el-button text type="primary" size="small" @click="router.push(`/product/edit/${row.id}`)">
          编辑
        </el-button>
        <el-button
          v-if="row.status !== 'onsale'"
          text
          type="success"
          size="small"
          @click="handleOnSale(row)"
        >
          上架
        </el-button>
        <el-button v-else text type="warning" size="small" @click="handleOffSale(row)">
          下架
        </el-button>
        <el-dropdown>
          <el-button text size="small">更多</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleCopy(row)">复制</el-dropdown-item>
              <el-dropdown-item @click="handleDelete(row)" style="color: #ef4444">删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </ProTable>
  </div>
</template>
