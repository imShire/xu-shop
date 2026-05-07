<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { getRegions, type RegionNode } from '@/api/region'
import {
  getSenderAddresses,
  createSenderAddress,
  updateSenderAddress,
  deleteSenderAddress,
  setDefaultSenderAddress,
} from '@/api/shipping'
import type { SenderAddress } from '@/types'

const loading = ref(false)
const addresses = ref<SenderAddress[]>([])
const dialogVisible = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()
const isEdit = ref(false)
const regionOptions = ref<RegionCascaderNode[]>([])

interface SenderAddressForm {
  id: string
  name: string
  phone: string
  province_code: string
  province: string
  city_code: string
  city: string
  district_code: string
  district: string
  street_code: string
  street: string
  detail: string
  region_codes: string[]
}

interface RegionCascaderNode {
  code: string
  name: string
  leaf: boolean
  children?: RegionCascaderNode[]
}

function createEmptyForm(): SenderAddressForm {
  return {
    id: '',
    name: '',
    phone: '',
    province_code: '',
    province: '',
    city_code: '',
    city: '',
    district_code: '',
    district: '',
    street_code: '',
    street: '',
    detail: '',
    region_codes: [],
  }
}

const form = ref<SenderAddressForm>(createEmptyForm())

const cascaderProps = {
  lazy: true,
  emitPath: true,
  value: 'code',
  label: 'name',
  leaf: 'leaf',
  async lazyLoad(node: { level: number; value?: string | number }, resolve: (data: RegionCascaderNode[]) => void) {
    if (node.level === 0) {
      await ensureRootOptions()
      resolve(regionOptions.value)
      return
    }
    const children = await loadRegionChildren(String(node.value ?? ''))
    resolve(children)
  },
}

async function loadData() {
  loading.value = true
  try {
    addresses.value = await getSenderAddresses() ?? []
  } finally {
    loading.value = false
  }
}

function toRegionOption(region: RegionNode): RegionCascaderNode {
  return {
    code: region.code,
    name: region.name,
    leaf: region.has_children === false || region.level >= 4,
  }
}

function formatAddress(row: Pick<SenderAddressForm, 'province' | 'city' | 'district' | 'street' | 'detail'>) {
  return [row.province, row.city, row.district, row.street, row.detail].filter(Boolean).join('')
}

function applyChildren(nodes: RegionCascaderNode[], targetCode: string, children: RegionCascaderNode[]): boolean {
  for (const node of nodes) {
    if (node.code === targetCode) {
      node.children = children
      return true
    }
    if (node.children && applyChildren(node.children, targetCode, children)) {
      return true
    }
  }
  return false
}

function findRegionPath(nodes: RegionCascaderNode[], codes: string[], depth = 0): RegionCascaderNode[] {
  if (depth >= codes.length) return []
  const current = nodes.find((node) => node.code === codes[depth])
  if (!current) return []
  if (depth === codes.length - 1) return [current]
  return [current, ...findRegionPath(current.children ?? [], codes, depth + 1)]
}

async function ensureRootOptions() {
  if (regionOptions.value.length > 0) return
  regionOptions.value = (await getRegions()).map(toRegionOption)
}

async function loadRegionChildren(parentCode: string) {
  const children = (await getRegions(parentCode)).map(toRegionOption)
  if (parentCode) {
    applyChildren(regionOptions.value, parentCode, children)
    regionOptions.value = [...regionOptions.value]
  }
  return children
}

async function ensureRegionPath(regionCodes: string[]) {
  await ensureRootOptions()
  for (const code of regionCodes.slice(0, -1)) {
    await loadRegionChildren(code)
  }
}

async function openCreate() {
  await ensureRootOptions()
  form.value = createEmptyForm()
  isEdit.value = false
  dialogVisible.value = true
}

async function openEdit(row: SenderAddress) {
  await ensureRootOptions()
  form.value = {
    id: row.id,
    name: row.name,
    phone: row.phone,
    province_code: row.province_code ?? '',
    province: row.province,
    city_code: row.city_code ?? '',
    city: row.city,
    district_code: row.district_code ?? '',
    district: row.district,
    street_code: row.street_code ?? '',
    street: row.street ?? '',
    detail: row.detail,
    region_codes: [
      row.province_code,
      row.city_code,
      row.district_code,
      row.street_code,
    ].filter(Boolean) as string[],
  }
  await ensureRegionPath(form.value.region_codes)
  isEdit.value = true
  dialogVisible.value = true
}

function handleRegionChange(value?: string[]) {
  const regionCodes = (value ?? []).map(String)
  const selected = findRegionPath(regionOptions.value, regionCodes)
  form.value.region_codes = regionCodes
  form.value.province_code = regionCodes[0] ?? ''
  form.value.province = selected[0]?.name ?? ''
  form.value.city_code = regionCodes[1] ?? ''
  form.value.city = selected[1]?.name ?? ''
  form.value.district_code = regionCodes[2] ?? ''
  form.value.district = selected[2]?.name ?? ''
  form.value.street_code = regionCodes[3] ?? ''
  form.value.street = selected[3]?.name ?? ''
}

async function handleSave() {
  await formRef.value?.validate()
  if (!form.value.province) {
    ElMessage.warning('请选择所在地区')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      phone: form.value.phone,
      province_code: form.value.province_code || undefined,
      province: form.value.province,
      city_code: form.value.city_code || undefined,
      city: form.value.city,
      district_code: form.value.district_code || undefined,
      district: form.value.district,
      street_code: form.value.street_code || undefined,
      street: form.value.street || undefined,
      detail: form.value.detail,
    }
    if (isEdit.value) {
      await updateSenderAddress(form.value.id, payload)
    } else {
      await createSenderAddress(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm('确认删除该发件地址？', '提示', { type: 'warning' })
    await deleteSenderAddress(row.id)
    ElMessage.success('删除成功')
    await loadData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.message || '操作失败')
  }
}

async function handleSetDefault(row: any) {
  try {
    await setDefaultSenderAddress(row.id)
    ElMessage.success('已设为默认')
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="display: flex; justify-content: space-between; margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">发件地址</h3>
      <el-button type="primary" @click="openCreate">+ 新增地址</el-button>
    </div>

    <el-table v-loading="loading" :data="addresses" border>
      <el-table-column prop="name" label="联系人" width="100" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column label="地址" min-width="200">
        <template #default="{ row }">
          {{ formatAddress(row) }}
        </template>
      </el-table-column>
      <el-table-column label="默认" width="80" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.is_default" type="success" size="small">默认</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button text type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="!row.is_default" text size="small" @click="handleSetDefault(row)">设默认</el-button>
          <el-button text type="danger" size="small" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑地址' : '新增地址'" width="480px">
      <el-form ref="formRef" :model="form" label-width="80px">
        <el-form-item label="联系人" prop="name" :rules="[{ required: true }]">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone" :rules="[{ required: true }]">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="所在地区" prop="region_codes" :rules="[{ required: true, message: '请选择所在地区' }]">
          <el-cascader
            v-model="form.region_codes"
            :options="regionOptions"
            :props="cascaderProps"
            clearable
            style="width: 100%"
            placeholder="请选择省/市/区/街道"
            @change="handleRegionChange"
          />
          <div v-if="form.province" style="margin-top: 8px; font-size: 12px; color: var(--text-secondary)">
            当前选择：{{ [form.province, form.city, form.district, form.street].filter(Boolean).join(' / ') }}
          </div>
        </el-form-item>
        <el-form-item label="详细地址" prop="detail" :rules="[{ required: true }]">
          <el-input v-model="form.detail" type="textarea" placeholder="楼栋、单元、门牌号等" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
