<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { getProductList, getCategoryList } from '@/api/product'
import { getArticles } from '@/api/cms'
import type { LinkConfig } from '@/types/link'

const props = defineProps<{
  modelValue: LinkConfig | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: LinkConfig]
}>()

const PAGE_SIZE = 20

const FIXED_PAGES: { name: string; url: string }[] = [
  { name: '商品列表', url: '/pages/product/list' },
  { name: '购物车', url: '/pages/cart/index' },
  { name: '个人中心', url: '/pages/user/profile' },
  { name: '订单列表', url: '/pages/order/list' },
]

const dialogVisible = ref(false)
const activeTab = ref('product')

const displayText = computed(() => {
  if (!props.modelValue) return ''
  return props.modelValue.target_name || props.modelValue.url || ''
})

// --- Product tab ---
const productKeyword = ref('')
const productLoading = ref(false)
const productList = ref<any[]>([])
const productPage = ref(1)
const productTotal = ref(0)

async function fetchProducts() {
  productLoading.value = true
  try {
    const res = await getProductList({
      keyword: productKeyword.value || undefined,
      page: productPage.value,
      page_size: PAGE_SIZE,
    })
    productList.value = res?.list ?? []
    productTotal.value = res?.total ?? 0
  } finally {
    productLoading.value = false
  }
}

function onProductSearch() {
  productPage.value = 1
  fetchProducts()
}

function onProductPageChange(page: number) {
  productPage.value = page
  fetchProducts()
}

// --- Category tab ---
const categoryLoading = ref(false)
const categoryList = ref<any[]>([])

async function fetchCategories() {
  categoryLoading.value = true
  try {
    const res = await getCategoryList()
    categoryList.value = Array.isArray(res) ? res : []
  } finally {
    categoryLoading.value = false
  }
}

// --- Article tab ---
const articleKeyword = ref('')
const articleLoading = ref(false)
const articleList = ref<any[]>([])
const articlePage = ref(1)
const articleTotal = ref(0)

async function fetchArticles() {
  articleLoading.value = true
  try {
    const res = await getArticles({
      keyword: articleKeyword.value || undefined,
      status: 'published',
      page: articlePage.value,
      page_size: PAGE_SIZE,
    })
    articleList.value = res?.list ?? []
    articleTotal.value = res?.total ?? 0
  } finally {
    articleLoading.value = false
  }
}

function onArticleSearch() {
  articlePage.value = 1
  fetchArticles()
}

function onArticlePageChange(page: number) {
  articlePage.value = page
  fetchArticles()
}

// --- Custom URL ---
const customUrl = ref('')

// --- Open dialog ---
function openDialog() {
  dialogVisible.value = true
  const typeToTab: Record<string, string> = {
    product: 'product',
    category: 'category',
    article: 'article',
    product_list: 'fixed',
    custom: 'product',
  }
  activeTab.value = typeToTab[props.modelValue?.type ?? ''] ?? 'product'
  customUrl.value = ''
  if (productList.value.length === 0) fetchProducts()
}

// Lazy load tabs on switch
watch(activeTab, (name) => {
  if (name === 'category' && categoryList.value.length === 0) fetchCategories()
  if (name === 'article' && articleList.value.length === 0) fetchArticles()
})

// --- Select handlers ---
function selectProduct(row: any) {
  emit('update:modelValue', {
    type: 'product',
    target_id: String(row.id),
    target_name: row.name,
    url: '/pages/product/detail?id=' + row.id,
  })
  dialogVisible.value = false
}

function selectCategory(row: any) {
  emit('update:modelValue', {
    type: 'category',
    target_id: String(row.id),
    target_name: row.name,
    url: '/pages/category?id=' + row.id,
  })
  dialogVisible.value = false
}

function selectArticle(row: any) {
  emit('update:modelValue', {
    type: 'article',
    target_id: String(row.id),
    target_name: row.title,
    url: '/pages/cms/detail?id=' + row.id,
  })
  dialogVisible.value = false
}

function selectFixedPage(row: { name: string; url: string }) {
  emit('update:modelValue', {
    type: 'custom',
    target_name: row.name,
    url: row.url,
  })
  dialogVisible.value = false
}

function useCustomUrl() {
  const url = customUrl.value.trim()
  if (!url) return
  emit('update:modelValue', {
    type: 'custom',
    url,
  })
  dialogVisible.value = false
}
</script>

<template>
  <div class="link-picker">
    <div style="display: flex; gap: 8px; align-items: center">
      <el-input
        :model-value="displayText"
        readonly
        placeholder="未选择链接"
        style="flex: 1"
      />
      <el-button @click="openDialog">选择</el-button>
    </div>

    <el-dialog
      v-model="dialogVisible"
      title="选择链接"
      width="640px"
      destroy-on-close
    >
      <el-tabs v-model="activeTab">
        <!-- 商品 Tab -->
        <el-tab-pane label="商品" name="product">
          <div style="display: flex; gap: 8px; margin-bottom: 12px">
            <el-input
              v-model="productKeyword"
              placeholder="搜索商品名称"
              clearable
              @keyup.enter="onProductSearch"
            />
            <el-button type="primary" :icon="Search" @click="onProductSearch">搜索</el-button>
          </div>
          <el-table
            v-loading="productLoading"
            :data="productList"
            size="small"
            border
            style="width: 100%"
          >
            <el-table-column prop="name" label="商品名称" min-width="180" show-overflow-tooltip />
            <el-table-column label="价格" width="100" align="right">
              <template #default="{ row }">
                ¥{{ ((row.price_cents ?? 0) / 100).toFixed(2) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="selectProduct(row)">选择</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div style="display: flex; justify-content: flex-end; margin-top: 12px">
            <el-pagination
              v-model:current-page="productPage"
              :page-size="PAGE_SIZE"
              :total="productTotal"
              small
              layout="prev, pager, next"
              @current-change="onProductPageChange"
            />
          </div>
        </el-tab-pane>

        <!-- 分类 Tab -->
        <el-tab-pane label="分类" name="category">
          <el-table
            v-loading="categoryLoading"
            :data="categoryList"
            size="small"
            border
            max-height="300"
            style="width: 100%"
          >
            <el-table-column prop="name" label="分类名称" min-width="200" show-overflow-tooltip />
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="selectCategory(row)">选择</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 文章 Tab -->
        <el-tab-pane label="文章" name="article">
          <div style="display: flex; gap: 8px; margin-bottom: 12px">
            <el-input
              v-model="articleKeyword"
              placeholder="搜索文章标题"
              clearable
              @keyup.enter="onArticleSearch"
            />
            <el-button type="primary" :icon="Search" @click="onArticleSearch">搜索</el-button>
          </div>
          <el-table
            v-loading="articleLoading"
            :data="articleList"
            size="small"
            border
            style="width: 100%"
          >
            <el-table-column prop="title" label="文章标题" min-width="200" show-overflow-tooltip />
            <el-table-column label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag type="success" size="small">{{ row.status === 'published' ? '已发布' : row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="selectArticle(row)">选择</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div style="display: flex; justify-content: flex-end; margin-top: 12px">
            <el-pagination
              v-model:current-page="articlePage"
              :page-size="PAGE_SIZE"
              :total="articleTotal"
              small
              layout="prev, pager, next"
              @current-change="onArticlePageChange"
            />
          </div>
        </el-tab-pane>

        <!-- 固定页面 Tab -->
        <el-tab-pane label="固定页面" name="fixed">
          <el-table
            :data="FIXED_PAGES"
            size="small"
            border
            style="width: 100%"
          >
            <el-table-column prop="name" label="页面名称" width="160" />
            <el-table-column prop="url" label="路径" min-width="200" />
            <el-table-column label="操作" width="80" align="center">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="selectFixedPage(row)">选择</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

      <!-- 自定义 URL -->
      <el-divider />
      <div style="display: flex; gap: 8px; align-items: center">
        <span style="white-space: nowrap; font-size: 14px; color: var(--el-text-color-regular)">自定义 URL：</span>
        <el-input v-model="customUrl" placeholder="输入自定义跳转路径" clearable />
        <el-button type="primary" @click="useCustomUrl">使用此链接</el-button>
      </div>
    </el-dialog>
  </div>
</template>
