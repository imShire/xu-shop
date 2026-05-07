<script setup lang="ts">
import { ref, onMounted, computed, watch, shallowRef, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { Editor, Toolbar } from '@wangeditor/editor-for-vue'
import '@wangeditor/editor/dist/css/style.css'
import UploadImage from '@/components/UploadImage/index.vue'
import PriceInput from '@/components/PriceInput/index.vue'
import {
  getProduct,
  createProduct,
  updateProduct,
  getCategoryList,
  uploadImage,
} from '@/api/product'
import { getFreightTemplates } from '@/api/order'

const route = useRoute()
const router = useRouter()
const productId = computed(() => (route.params.id ? String(route.params.id) : null))
const isEdit = computed(() => !!productId.value)

const formRef = ref<FormInstance>()
const saving = ref(false)
const categories = ref<any[]>([])
const freightTemplates = ref<any[]>([])

// wangEditor
const editorRef = shallowRef()
const editorConfig = {
  MENU_CONF: {
    uploadImage: {
      async customUpload(file: File, insertFn: (url: string) => void) {
        const res = await uploadImage(file)
        insertFn(res.url)
      },
    },
  },
}
onBeforeUnmount(() => editorRef.value?.destroy())

const UNIT_OPTIONS = ['件', '个', '套', 'kg', '份', '盒']

const form = ref<{
  title: string
  subtitle: string
  category_id: string | undefined
  unit: string
  tags: string[]
  sort: number
  status: string
  main_image: string
  images: string[]
  video_url: string
  specs: { name: string; values: string[] }[]
  skus: {
    id?: string
    spec_values: string[]
    price_cents: number
    original_price_cents: number
    stock: number
    weight_g: number
    sku_code: string
    barcode: string
    image: string
    status: 'active' | 'inactive'
  }[]
  is_virtual: boolean
  freight_template_id: string | null
  detail_html: string
  virtual_sales: number
  on_sale_at: string | null
}>({
  title: '',
  subtitle: '',
  category_id: undefined,
  unit: '件',
  tags: [],
  sort: 0,
  status: 'draft',
  main_image: '',
  images: [],
  video_url: '',
  specs: [],
  skus: [],
  is_virtual: false,
  freight_template_id: null,
  detail_html: '',
  virtual_sales: 0,
  on_sale_at: null,
})

// Cartesian product for SKU generation
function cartesian(specs: { name: string; values: string[] }[]): string[][] {
  if (!specs.length) return [[]]
  const [head, ...tail] = specs
  const tailCombinations = cartesian(tail)
  return head.values.flatMap((v) => tailCombinations.map((combo) => [v, ...combo]))
}

function rebuildSkus() {
  if (!form.value.specs.length) return
  const combinations = cartesian(form.value.specs)
  const existingMap = new Map(form.value.skus.map((s) => [s.spec_values.join('|'), s]))
  form.value.skus = combinations.map((combo) => {
    const key = combo.join('|')
    const existing = existingMap.get(key)
    return {
      id: existing?.id,
      spec_values: combo,
      price_cents: existing?.price_cents ?? 0,
      original_price_cents: existing?.original_price_cents ?? 0,
      stock: existing?.stock ?? 0,
      weight_g: existing?.weight_g ?? 0,
      sku_code: existing?.sku_code ?? '',
      barcode: existing?.barcode ?? '',
      image: existing?.image ?? '',
      status: existing?.status ?? 'active',
    }
  })
}

watch(() => form.value.specs, rebuildSkus, { deep: true })

function addSpec() {
  if (form.value.specs.length >= 3) return
  form.value.specs.push({ name: '', values: [] })
}

function removeSpec(idx: number) {
  form.value.specs.splice(idx, 1)
  if (!form.value.specs.length) {
    form.value.skus = []
  }
}

const specInputs = ref<string[]>([])

function addSpecValue(specIdx: number, val: string) {
  if (val && !form.value.specs[specIdx].values.includes(val)) {
    form.value.specs[specIdx].values.push(val)
    specInputs.value[specIdx] = ''
  }
}

function removeSpecValue(specIdx: number, valIdx: number) {
  form.value.specs[specIdx].values.splice(valIdx, 1)
}

// 无规格单 SKU 快捷模式
const singleSku = computed(() => {
  if (form.value.specs.length > 0) return null
  return form.value.skus[0] ?? null
})

watch(
  () => form.value.specs.length,
  (len) => {
    if (len === 0) {
      if (!form.value.skus.length) {
        form.value.skus = [
          {
            spec_values: [],
            price_cents: 0,
            original_price_cents: 0,
            stock: 0,
            weight_g: 0,
            sku_code: '',
            barcode: '',
            image: '',
            status: 'active',
          },
        ]
      }
    }
  },
  { immediate: true }
)

const formRules = {
  title: [
    { required: true, message: '请输入商品标题', trigger: 'blur' },
    { max: 60, message: '标题最多 60 个字符', trigger: 'blur' },
  ],
  category_id: [{ required: true, message: '请选择商品分类', trigger: 'change' }],
  main_image: [{ required: true, message: '请上传商品主图', trigger: 'change' }],
}

async function handleSave(targetStatus?: string) {
  await formRef.value?.validate()

  // Validate SKU prices
  const hasInvalidSku = form.value.skus.some((sku) => !(sku.price_cents > 0))
  if (hasInvalidSku) {
    ElMessage.warning('请为所有 SKU 设置售价')
    return
  }

  saving.value = true
  try {
    const payload: any = {
      ...form.value,
      skus: form.value.skus.map((sku) => ({
        ...sku,
        // spec_values 即为 attrs 格式（字符串数组），后端以此关联规格
        attrs: sku.spec_values,
        price_cents: Math.round(sku.price_cents),
        original_price_cents: Math.round(sku.original_price_cents),
      })),
    }
    if (targetStatus) {
      payload.status = targetStatus
    }

    if (isEdit.value) {
      await updateProduct(productId.value!, payload)
      ElMessage.success('保存成功')
      if (targetStatus) form.value.status = targetStatus
    } else {
      await createProduct(payload)
      ElMessage.success('创建成功')
      router.push('/product/list')
    }
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  ;[categories.value, freightTemplates.value] = await Promise.all([
    getCategoryList().catch(() => []),
    getFreightTemplates().catch(() => []),
  ])

  if (isEdit.value) {
    const product = await getProduct(productId.value!)
    Object.assign(form.value, {
      ...product,
      unit: product.unit || '件',
      sort: product.sort ?? 0,
      virtual_sales: product.virtual_sales ?? 0,
      is_virtual: product.is_virtual ?? false,
      freight_template_id: product.freight_template_id ?? null,
      on_sale_at: product.on_sale_at ?? null,
      detail_html: product.detail_html ?? '',
      video_url: product.video_url ?? '',
      // 将 API 返回的 SpecResp[]（values 为对象数组）转换为表单所需格式（values 为字符串数组）
      specs: (product.specs || []).map((spec: any) => ({
        name: spec.name,
        values: (spec.values || []).map((v: any) => (typeof v === 'string' ? v : v.value)),
      })),
      // 将 attrs（字符串数组）映射为 spec_values，保留现有 SKU 的 id
      skus: (product.skus || []).map((sku: any) => ({
        ...sku,
        spec_values: Array.isArray(sku.attrs) ? sku.attrs : [],
        price_cents: sku.price_cents,
        original_price_cents: sku.original_price_cents,
        weight_g: sku.weight_g ?? 0,
        barcode: sku.barcode ?? '',
        image: sku.image ?? '',
        status: sku.status ?? 'active',
      })),
    })
  } else {
    // Init single SKU for no-spec mode
    form.value.skus = [
      {
        spec_values: [],
        price_cents: 0,
        original_price_cents: 0,
        stock: 0,
        weight_g: 0,
        sku_code: '',
        barcode: '',
        image: '',
        status: 'active',
      },
    ]
  }
})
</script>

<template>
  <div class="page-container">
    <!-- Page header -->
    <div class="page-card" style="margin-bottom: 16px; display: flex; align-items: center; gap: 12px">
      <el-button @click="router.back()">返回</el-button>
      <h3 style="flex: 1; font-size: 16px; font-weight: 600; margin: 0">
        {{ isEdit ? '编辑商品' : '新建商品' }}
      </h3>
      <el-button :loading="saving" @click="handleSave()">保存草稿</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave('onsale')">立即上架</el-button>
    </div>

    <el-form ref="formRef" :model="form" :rules="formRules" label-width="110px">
      <!-- Card 1: 基本信息 -->
      <el-card style="margin-bottom: 16px">
        <template #header>基本信息</template>
        <el-form-item label="商品标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入商品标题（最多 60 字）" maxlength="60" show-word-limit />
        </el-form-item>
        <el-form-item label="副标题" prop="subtitle">
          <el-input v-model="form.subtitle" placeholder="请输入副标题" />
        </el-form-item>
        <el-form-item label="商品分类" prop="category_id">
          <el-select v-model="form.category_id" placeholder="请选择分类" filterable style="width: 240px">
            <el-option
              v-for="cat in categories"
              :key="cat.id"
              :label="cat.name"
              :value="cat.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="单位" prop="unit">
          <el-select
            v-model="form.unit"
            allow-create
            filterable
            placeholder="选择或输入单位"
            style="width: 160px"
          >
            <el-option v-for="u in UNIT_OPTIONS" :key="u" :label="u" :value="u" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="form.tags"
            multiple
            allow-create
            filterable
            placeholder="输入标签按回车添加"
            style="width: 360px"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="form.sort" :min="0" controls-position="right" style="width: 120px" />
          <span style="margin-left: 8px; color: var(--text-secondary); font-size: 12px">数值越大越靠前</span>
        </el-form-item>
      </el-card>

      <!-- Card 2: 商品图片 -->
      <el-card style="margin-bottom: 16px">
        <template #header>商品图片</template>
        <el-form-item label="主图" prop="main_image">
          <UploadImage v-model="form.main_image" />
        </el-form-item>
        <el-form-item label="轮播图">
          <UploadImage v-model="form.images" multiple :limit="9" />
        </el-form-item>
        <el-form-item label="视频 URL">
          <el-input v-model="form.video_url" placeholder="视频 URL，可选" style="width: 100%" />
        </el-form-item>
      </el-card>

      <!-- Card 3: 规格与 SKU -->
      <el-card style="margin-bottom: 16px">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center">
            <span>规格与 SKU</span>
            <el-tooltip
              :content="form.specs.length >= 3 ? '最多 3 层规格' : '添加规格'"
              placement="top"
            >
              <el-button
                size="small"
                :disabled="form.specs.length >= 3"
                @click="addSpec"
              >
                + 添加规格
              </el-button>
            </el-tooltip>
          </div>
        </template>

        <!-- 无规格快捷模式 -->
        <template v-if="form.specs.length === 0">
          <el-table v-if="singleSku" :data="[singleSku]" border>
            <el-table-column label="售价(分)" width="160">
              <template #default="{ row }">
                <PriceInput v-model="row.price_cents" />
              </template>
            </el-table-column>
            <el-table-column label="原价(分)" width="160">
              <template #default="{ row }">
                <PriceInput v-model="row.original_price_cents" />
              </template>
            </el-table-column>
            <el-table-column label="库存" width="120">
              <template #default="{ row }">
                <el-input-number v-model="row.stock" :min="0" controls-position="right" style="width: 100px" />
              </template>
            </el-table-column>
            <el-table-column label="重量(g)" width="120">
              <template #default="{ row }">
                <el-input-number v-model="row.weight_g" :min="0" controls-position="right" style="width: 100px" />
              </template>
            </el-table-column>
            <el-table-column label="SKU码" width="160">
              <template #default="{ row }">
                <el-input v-model="row.sku_code" placeholder="可选" />
              </template>
            </el-table-column>
          </el-table>
        </template>

        <!-- 有规格：规格配置区 -->
        <template v-else>
          <div
            v-for="(spec, si) in form.specs"
            :key="si"
            style="margin-bottom: 16px; padding: 12px; border: 1px solid var(--border-color); border-radius: 6px"
          >
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 10px">
              <el-input
                v-model="spec.name"
                placeholder="规格名（如：颜色）"
                style="width: 160px"
              />
              <el-button text type="danger" size="small" @click="removeSpec(si)">删除规格</el-button>
            </div>
            <div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center">
              <el-tag
                v-for="(val, vi) in spec.values"
                :key="vi"
                closable
                @close="removeSpecValue(si, vi)"
              >
                {{ val }}
              </el-tag>
              <el-input
                v-model="specInputs[si]"
                placeholder="输入值后回车"
                style="width: 120px"
                @keyup.enter="addSpecValue(si, specInputs[si])"
              />
            </div>
          </div>

          <!-- 有规格 SKU 表格 -->
          <el-table v-if="form.skus.length" :data="form.skus" border style="margin-top: 16px" max-height="400">
            <el-table-column label="规格图" width="100">
              <template #default="{ row }">
                <UploadImage v-model="row.image" />
              </template>
            </el-table-column>
            <el-table-column label="规格组合" min-width="120">
              <template #default="{ row }">
                <el-tag
                  v-for="(v, i) in row.spec_values"
                  :key="i"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 2px"
                >
                  {{ v }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="售价(分)" width="150">
              <template #default="{ row }">
                <PriceInput v-model="row.price_cents" />
              </template>
            </el-table-column>
            <el-table-column label="原价(分)" width="150">
              <template #default="{ row }">
                <PriceInput v-model="row.original_price_cents" />
              </template>
            </el-table-column>
            <el-table-column label="库存" width="100">
              <template #default="{ row }">
                <el-input-number v-model="row.stock" :min="0" controls-position="right" style="width: 85px" />
              </template>
            </el-table-column>
            <el-table-column label="重量(g)" width="100">
              <template #default="{ row }">
                <el-input-number v-model="row.weight_g" :min="0" controls-position="right" style="width: 85px" />
              </template>
            </el-table-column>
            <el-table-column label="SKU码" width="140">
              <template #default="{ row }">
                <el-input v-model="row.sku_code" placeholder="可选" />
              </template>
            </el-table-column>
            <el-table-column label="条形码" width="140">
              <template #default="{ row }">
                <el-input v-model="row.barcode" placeholder="可选" />
              </template>
            </el-table-column>
            <el-table-column label="状态" width="90" align="center">
              <template #default="{ row }">
                <el-switch
                  v-model="row.status"
                  active-value="active"
                  inactive-value="inactive"
                />
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="请先添加规格值以生成 SKU" :image-size="60" style="margin-top: 16px" />
        </template>
      </el-card>

      <!-- Card 4: 物流设置 -->
      <el-card style="margin-bottom: 16px">
        <template #header>物流设置</template>
        <el-form-item label="虚拟商品">
          <el-switch
            v-model="form.is_virtual"
            active-text="虚拟商品（数字/卡券，无需发货）"
          />
        </el-form-item>
        <el-form-item v-if="!form.is_virtual" label="运费模板">
          <el-select
            v-model="form.freight_template_id"
            placeholder="请选择运费模板"
            style="width: 240px"
          >
            <el-option label="包邮（不设运费模板）" :value="null" />
            <el-option
              v-for="tpl in freightTemplates"
              :key="tpl.id"
              :label="tpl.name"
              :value="tpl.id"
            />
          </el-select>
        </el-form-item>
      </el-card>

      <!-- Card 5: 图文详情 -->
      <el-card style="margin-bottom: 16px">
        <template #header>图文详情</template>
        <el-form-item label="图文详情" label-width="110px">
          <div style="border: 1px solid #ccc; width: 100%">
            <Toolbar
              :editor="editorRef"
              :defaultConfig="{}"
              mode="default"
              style="border-bottom: 1px solid #ccc"
            />
            <Editor
              v-model="form.detail_html"
              :defaultConfig="editorConfig"
              mode="default"
              style="height: 400px; overflow-y: hidden"
              @onCreated="(e: any) => (editorRef = e)"
            />
          </div>
        </el-form-item>
      </el-card>

      <!-- Card 6: 营销与发布 -->
      <el-card style="margin-bottom: 16px">
        <template #header>营销与发布</template>
        <el-form-item label="虚拟销量">
          <el-input-number v-model="form.virtual_sales" :min="0" controls-position="right" style="width: 120px" />
          <span style="margin-left: 8px; color: var(--text-secondary); font-size: 12px">展示用，不影响真实数据</span>
        </el-form-item>
        <el-form-item v-if="form.status === 'draft'" label="定时上架">
          <el-date-picker
            v-model="form.on_sale_at"
            type="datetime"
            placeholder="选择定时上架时间（可选）"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="draft">草稿</el-radio>
            <el-radio value="onsale">上架中</el-radio>
            <el-radio value="offsale">已下架</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-card>
    </el-form>
  </div>
</template>

<style scoped>
.page-container {
  padding: 20px;
  max-width: 1400px;
}
</style>
