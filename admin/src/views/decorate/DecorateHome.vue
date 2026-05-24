<script setup lang="ts">
import Sortable from 'sortablejs'
import { ref, onMounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPageVersions, savePageConfig, activatePageConfig } from '@/api/decorate'
import type { PageModule, PageConfig } from '@/api/decorate'
import UploadImage from '@/components/UploadImage/index.vue'
import LinkPicker from '@/components/LinkPicker/index.vue'
import type { LinkConfig } from '@/types/link'
import { getProductList } from '@/api/product'

const pageKey = ref('home')
const versions = ref<PageConfig[]>([])
const loading = ref(false)
const saving = ref(false)

const modules = ref<PageModule[]>([])
const newModuleType = ref<string>('product_list')

const MODULE_LABELS: Record<string, string> = {
  product_list: '商品推荐',
  category_entry: '分类入口',
  rich_text: '富文本',
  image_ad: '广告图片',
}

// ── 拖拽排序 ──────────────────────────────────────────────────────────────────
const moduleListRef = ref<HTMLElement | null>(null)
let sortableInstance: Sortable | null = null

function initSortable() {
  if (sortableInstance) { sortableInstance.destroy(); sortableInstance = null }
  if (!moduleListRef.value) return
  sortableInstance = Sortable.create(moduleListRef.value, {
    animation: 150,
    handle: '.drag-handle',
    onEnd({ oldIndex, newIndex }: { oldIndex?: number; newIndex?: number }) {
      if (oldIndex == null || newIndex == null || oldIndex === newIndex) return
      const arr = [...modules.value]
      const [item] = arr.splice(oldIndex, 1)
      arr.splice(newIndex, 0, item)
      modules.value = arr
    },
  })
}

// 深层 watch：category_entry link_config → 同步 link_url（保持兼容）
watch(
  modules,
  (mods) => {
    mods.forEach((mod) => {
      if (mod.type === 'category_entry') {
        const items = (mod.data as any).items as any[] | undefined
        items?.forEach((item) => {
          const cfg = item.link_config as LinkConfig | null
          if (cfg?.url) {
            item.link_url = cfg.url
          }
        })
      }
    })
  },
  { deep: true },
)

// 模块列表变化后重新初始化 Sortable
watch(modules, async () => {
  await nextTick()
  initSortable()
})

async function loadVersions() {
  loading.value = true
  try {
    versions.value = await getPageVersions(pageKey.value)
    const active = versions.value.find((v) => v.is_active)
    modules.value = active ? JSON.parse(JSON.stringify(active.modules)) : []
  } finally {
    loading.value = false
  }
}

watch(pageKey, loadVersions)

function addModule() {
  const defaults: Record<string, Record<string, unknown>> = {
    product_list: { title: '推荐商品', sort: 'latest', limit: 4, manual: false, product_ids: [] },
    category_entry: { items: [] },
    rich_text: { content: '' },
    image_ad: { image_url: '', alt: '', link_config: null },
  }
  modules.value.push({
    type: newModuleType.value as PageModule['type'],
    data: defaults[newModuleType.value] ?? {},
  })
}

function removeModule(idx: number) {
  modules.value.splice(idx, 1)
}

function moveUp(idx: number) {
  if (idx > 0) {
    const tmp = modules.value[idx - 1]
    modules.value[idx - 1] = modules.value[idx]
    modules.value[idx] = tmp
  }
}

function moveDown(idx: number) {
  if (idx < modules.value.length - 1) {
    const tmp = modules.value[idx + 1]
    modules.value[idx + 1] = modules.value[idx]
    modules.value[idx] = tmp
  }
}

async function handleSave() {
  saving.value = true
  try {
    await savePageConfig({ page_key: pageKey.value, modules: modules.value })
    ElMessage.success('保存成功（新版本未激活）')
    await loadVersions()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleActivate(version: PageConfig) {
  await ElMessageBox.confirm(
    `确认激活 v${version.version}？当前激活版本将被替换。`,
    '激活确认',
    { type: 'warning' },
  )
  try {
    await activatePageConfig(version.id, pageKey.value)
    ElMessage.success('已激活')
    await loadVersions()
  } catch {
    ElMessage.error('激活失败')
  }
}

// ── ProductPicker ─────────────────────────────────────────────────────────────
const pickerVisible = ref(false)
const pickerLoading = ref(false)
const pickerKeyword = ref('')
const pickerProducts = ref<any[]>([])
const pickerPage = ref(1)
const pickerTotal = ref(0)
const pickerSelected = ref<any[]>([])
let pickerTargetData: any = null

async function loadPickerProducts() {
  pickerLoading.value = true
  try {
    const res = await getProductList({
      keyword: pickerKeyword.value,
      page: pickerPage.value,
      page_size: 20,
    } as any)
    pickerProducts.value = res?.items ?? res?.data ?? (Array.isArray(res) ? res : [])
    pickerTotal.value = res?.total ?? pickerProducts.value.length
  } finally {
    pickerLoading.value = false
  }
}

function openProductPicker(modData: any) {
  pickerTargetData = modData
  pickerKeyword.value = ''
  pickerPage.value = 1
  pickerSelected.value = []
  pickerVisible.value = true
  loadPickerProducts()
}

function onPickerPageChange(p: number) {
  pickerPage.value = p
  loadPickerProducts()
}

function onPickerSelectionChange(rows: any[]) {
  pickerSelected.value = rows
}

function confirmPicker() {
  if (!pickerTargetData) return
  const existingIds: string[] = pickerTargetData.product_ids ?? []
  const newIds = pickerSelected.value.map((r: any) => String(r.id))
  pickerTargetData.product_ids = Array.from(new Set([...existingIds, ...newIds])).slice(0, 20)
  pickerVisible.value = false
}

// ── 手机预览摘要 ───────────────────────────────────────────────────────────────
function getModulePreview(mod: PageModule): string {
  const d = mod.data as any
  switch (mod.type) {
    case 'product_list':
      return d.manual
        ? `手动选品 | ${(d.product_ids as string[])?.length ?? 0} 件`
        : `sort: ${d.sort ?? 'latest'} | limit: ${d.limit ?? 4}`
    case 'category_entry':
      return `${(d.items as any[])?.length ?? 0} 条`
    case 'rich_text':
      return String(d.content ?? '')
        .replace(/<[^>]+>/g, '')
        .slice(0, 20)
    case 'image_ad':
      return d.image_url
        ? String(d.image_url).split('/').pop()?.slice(0, 30) ?? '图片已设'
        : '(未设图片)'
    default:
      return mod.type
  }
}

onMounted(loadVersions)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <!-- 页面 key 切换 -->
    <el-tabs v-model="pageKey" style="margin-bottom: 8px">
      <el-tab-pane label="首页" name="home" />
      <el-tab-pane label="分类页" name="category" />
    </el-tabs>

    <el-row :gutter="24">
      <!-- 左：编辑区 -->
      <el-col :span="16">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
          <h3 style="margin: 0">{{ pageKey === 'home' ? '首页' : '分类页' }}装修</h3>
          <div>
            <el-select v-model="newModuleType" style="width: 130px; margin-right: 8px">
              <el-option
                v-for="(label, type) in MODULE_LABELS"
                :key="type"
                :value="type"
                :label="label"
              />
            </el-select>
            <el-button @click="addModule">添加模块</el-button>
            <el-button type="primary" :loading="saving" style="margin-left: 8px" @click="handleSave">
              保存新版本
            </el-button>
          </div>
        </div>

        <div v-if="modules.length === 0" style="padding: 40px; text-align: center; color: #999">
          暂无模块，点击「添加模块」开始配置
        </div>

        <!-- 模块列表（拖拽容器，安装 sortablejs 后自动启用） -->
        <div ref="moduleListRef">
          <el-card v-for="(mod, idx) in modules" :key="idx" style="margin-bottom: 12px">
            <template #header>
              <div style="display: flex; align-items: center; justify-content: space-between">
                <div style="display: flex; align-items: center; gap: 8px">
                  <span class="drag-handle" title="拖拽排序（需安装 sortablejs）">⠿</span>
                  <el-tag size="small">{{ MODULE_LABELS[mod.type] || mod.type }}</el-tag>
                </div>
                <div>
                  <el-button text size="small" :disabled="idx === 0" @click="moveUp(idx)">↑</el-button>
                  <el-button text size="small" :disabled="idx === modules.length - 1" @click="moveDown(idx)">↓</el-button>
                  <el-button text type="danger" size="small" @click="removeModule(idx)">删除</el-button>
                </div>
              </div>
            </template>

            <!-- product_list 子表单 -->
            <template v-if="mod.type === 'product_list'">
              <el-form-item label="标题">
                <el-input v-model="(mod.data as any).title" placeholder="推荐商品" />
              </el-form-item>
              <el-form-item label="手动选品">
                <el-switch v-model="(mod.data as any).manual" />
              </el-form-item>
              <!-- 自动选品字段 -->
              <template v-if="!(mod.data as any).manual">
                <el-form-item label="排序方式">
                  <el-select v-model="(mod.data as any).sort">
                    <el-option value="latest" label="最新" />
                    <el-option value="popular" label="最热" />
                    <el-option value="hot" label="热门" />
                  </el-select>
                </el-form-item>
                <el-form-item label="展示数量">
                  <el-input-number v-model="(mod.data as any).limit" :min="2" :max="8" />
                </el-form-item>
              </template>
              <!-- 手动选品字段 -->
              <template v-else>
                <el-form-item label="已选商品">
                  <div style="width: 100%">
                    <div
                      v-for="(pid, pidIdx) in ((mod.data as any).product_ids as string[])"
                      :key="pid"
                      style="margin-bottom: 4px"
                    >
                      <el-tag
                        closable
                        @close="((mod.data as any).product_ids as string[]).splice(pidIdx, 1)"
                      >
                        ID: {{ pid }}
                      </el-tag>
                    </div>
                    <el-empty
                      v-if="!((mod.data as any).product_ids as string[])?.length"
                      description="暂未选择商品"
                      :image-size="40"
                    />
                    <el-button size="small" style="margin-top: 8px" @click="openProductPicker(mod.data)">
                      选择商品
                    </el-button>
                  </div>
                </el-form-item>
              </template>
            </template>

            <!-- category_entry 子表单 -->
            <template v-else-if="mod.type === 'category_entry'">
              <div
                v-for="(item, itemIdx) in ((mod.data as any).items as any[])"
                :key="itemIdx"
                style="border: 1px solid #eee; padding: 12px; border-radius: 4px; margin-bottom: 8px"
              >
                <el-form-item label="标题">
                  <el-input v-model="item.title" placeholder="如：春季新品" />
                </el-form-item>
                <el-form-item label="图片">
                  <UploadImage v-model="item.image_url" />
                </el-form-item>
                <el-form-item label="跳转链接">
                  <LinkPicker v-model="item.link_config" />
                </el-form-item>
                <el-button
                  text
                  type="danger"
                  size="small"
                  @click="((mod.data as any).items as any[]).splice(itemIdx, 1)"
                >
                  删除此条
                </el-button>
              </div>
              <el-button
                size="small"
                @click="((mod.data as any).items as any[]).push({ title: '', image_url: '', link_url: '', link_config: null })"
              >
                + 添加条目
              </el-button>
            </template>

            <!-- rich_text 子表单 -->
            <template v-else-if="mod.type === 'rich_text'">
              <el-form-item label="内容（HTML）">
                <el-input
                  v-model="(mod.data as any).content"
                  type="textarea"
                  :rows="6"
                  placeholder="输入 HTML 内容，禁止使用 script 标签"
                />
                <div style="font-size: 12px; color: #e6a23c; margin-top: 4px">
                  ⚠️ 内容由服务端做安全过滤，禁止 script/style/iframe 等危险标签
                </div>
              </el-form-item>
            </template>

            <!-- image_ad 子表单 -->
            <template v-else-if="mod.type === 'image_ad'">
              <el-form-item label="标题（可选）">
                <el-input v-model="(mod.data as any).alt" placeholder="广告图描述文字" />
              </el-form-item>
              <el-form-item label="图片">
                <UploadImage v-model="(mod.data as any).image_url" />
              </el-form-item>
              <el-form-item label="链接">
                <LinkPicker v-model="(mod.data as any).link_config" />
              </el-form-item>
            </template>
          </el-card>
        </div>
      </el-col>

      <!-- 右：手机预览 + 版本历史 -->
      <el-col :span="8">
        <!-- 手机预览 -->
        <h4 style="margin: 0 0 8px">手机预览</h4>
        <div class="phone-frame" style="margin-bottom: 24px">
          <div class="phone-screen">
            <div
              v-if="!modules.length"
              style="text-align: center; color: #999; padding: 20px; font-size: 12px"
            >
              暂无模块
            </div>
            <div
              v-for="(mod, idx) in modules"
              :key="idx"
              class="preview-module"
            >
              <el-tag size="small" type="info" style="flex-shrink: 0">
                {{ MODULE_LABELS[mod.type] || mod.type }}
              </el-tag>
              <span class="preview-text">{{ getModulePreview(mod) }}</span>
            </div>
          </div>
        </div>

        <!-- 版本历史 -->
        <h4 style="margin: 0 0 8px">历史版本</h4>
        <el-empty v-if="!versions.length" description="暂无版本" />
        <el-card v-for="v in versions" :key="v.id" style="margin-bottom: 8px">
          <div style="display: flex; align-items: center; justify-content: space-between">
            <div>
              <span>v{{ v.version }}</span>
              <el-tag v-if="v.is_active" type="success" size="small" style="margin-left: 8px">激活中</el-tag>
            </div>
            <el-button
              v-if="!v.is_active"
              text
              type="primary"
              size="small"
              @click="handleActivate(v)"
            >
              激活
            </el-button>
          </div>
          <div style="font-size: 12px; color: #999; margin-top: 4px">{{ v.created_at }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- ProductPicker 对话框 -->
    <el-dialog v-model="pickerVisible" title="选择商品" width="640px" destroy-on-close>
      <div style="display: flex; gap: 8px; margin-bottom: 12px">
        <el-input
          v-model="pickerKeyword"
          placeholder="搜索商品名称"
          clearable
          style="width: 300px"
          @keyup.enter="loadPickerProducts"
        />
        <el-button @click="loadPickerProducts">搜索</el-button>
      </div>
      <el-table
        v-loading="pickerLoading"
        :data="pickerProducts"
        border
        height="320"
        @selection-change="onPickerSelectionChange"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column prop="id" label="ID" width="120" show-overflow-tooltip />
        <el-table-column prop="title" label="商品名称" min-width="180" show-overflow-tooltip />
        <el-table-column label="主图" width="70" align="center">
          <template #default="{ row }">
            <el-image
              v-if="row.main_image"
              :src="row.main_image"
              style="width: 40px; height: 40px; object-fit: cover; border-radius: 4px"
              fit="cover"
            />
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pickerPage"
        :page-size="20"
        :total="pickerTotal"
        layout="prev, pager, next"
        style="margin-top: 12px; display: flex; justify-content: flex-end"
        @current-change="onPickerPageChange"
      />
      <template #footer>
        <el-button @click="pickerVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmPicker">
          确认（已选 {{ pickerSelected.length }} 件）
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.drag-handle {
  cursor: grab;
  color: #999;
  font-size: 16px;
  user-select: none;
}

.phone-frame {
  width: 375px;
  border: 8px solid #333;
  border-radius: 32px;
  padding: 16px 8px;
  background: #fff;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.phone-screen {
  height: 460px;
  overflow-y: auto;
  border-radius: 16px;
  background: #f5f5f5;
  padding: 8px;
}

.preview-module {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  background: #fff;
  border-radius: 6px;
  padding: 8px;
  margin-bottom: 6px;
}

.preview-text {
  font-size: 12px;
  color: #666;
  word-break: break-all;
  line-height: 1.4;
}
</style>
