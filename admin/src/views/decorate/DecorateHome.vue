<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPageVersions, savePageConfig, activatePageConfig } from '@/api/decorate'
import type { PageModule, PageConfig } from '@/api/decorate'
import UploadImage from '@/components/UploadImage/index.vue'

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
}

async function loadVersions() {
  loading.value = true
  try {
    versions.value = await getPageVersions(pageKey.value)
    const active = versions.value.find((v) => v.is_active)
    if (active) {
      modules.value = JSON.parse(JSON.stringify(active.modules))
    }
  } finally {
    loading.value = false
  }
}

function addModule() {
  const defaults: Record<string, Record<string, unknown>> = {
    product_list: { title: '推荐商品', sort: 'latest', limit: 4 },
    category_entry: { items: [] },
    rich_text: { content: '' },
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
    { type: 'warning' }
  )
  try {
    await activatePageConfig(version.id, pageKey.value)
    ElMessage.success('已激活')
    await loadVersions()
  } catch {
    ElMessage.error('激活失败')
  }
}

onMounted(loadVersions)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <el-row :gutter="24">
      <!-- 左：编辑区 -->
      <el-col :span="16">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px">
          <h3>首页装修</h3>
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

        <el-card v-for="(mod, idx) in modules" :key="idx" style="margin-bottom: 12px">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between">
              <el-tag size="small">{{ MODULE_LABELS[mod.type] || mod.type }}</el-tag>
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
            <el-form-item label="排序方式">
              <el-select v-model="(mod.data as any).sort">
                <el-option value="latest" label="最新" />
                <el-option value="popular" label="最热" />
              </el-select>
            </el-form-item>
            <el-form-item label="展示数量">
              <el-input-number v-model="(mod.data as any).limit" :min="2" :max="8" />
            </el-form-item>
          </template>

          <!-- category_entry 子表单 -->
          <template v-else-if="mod.type === 'category_entry'">
            <div
              v-for="(item, itemIdx) in (mod.data as any).items"
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
                <el-input v-model="item.link_url" placeholder="/pages/category/index?id=1" />
              </el-form-item>
              <el-button text type="danger" size="small" @click="(mod.data as any).items.splice(itemIdx, 1)">
                删除此条
              </el-button>
            </div>
            <el-button size="small" @click="(mod.data as any).items.push({ title: '', image_url: '', link_url: '' })">
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
                placeholder="<p>输入 HTML 内容，禁止使用 <script> 标签</p>"
              />
              <div style="font-size: 12px; color: #e6a23c; margin-top: 4px">
                ⚠️ 内容由服务端做安全过滤，禁止 script/style/iframe 等危险标签
              </div>
            </el-form-item>
          </template>
        </el-card>
      </el-col>

      <!-- 右：版本历史 -->
      <el-col :span="8">
        <h4>历史版本</h4>
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
  </div>
</template>
