<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getCarriers, updateCarrier } from '@/api/shipping'

const loading = ref(false)
const carriers = ref<any[]>([])
const saving = ref<Record<string, boolean>>({})

async function loadData() {
  loading.value = true
  try {
    carriers.value = await getCarriers() ?? []
  } finally {
    loading.value = false
  }
}

async function toggleEnabled(row: any) {
  saving.value[row.code] = true
  try {
    await updateCarrier(row.code, { enabled: row.enabled })
    ElMessage.success('更新成功')
  } catch {
    row.enabled = !row.enabled
  } finally {
    delete saving.value[row.code]
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card">
    <div style="margin-bottom: 16px">
      <h3 style="font-size: 15px; font-weight: 600">快递商配置</h3>
    </div>

    <el-table v-loading="loading" :data="carriers" border>
      <el-table-column prop="name" label="快递名称" min-width="120" />
      <el-table-column prop="code" label="快递代码" width="120" />
      <el-table-column label="启用" width="100" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.enabled"
            :loading="!!saving[row.code]"
            @change="toggleEnabled(row)"
          />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
