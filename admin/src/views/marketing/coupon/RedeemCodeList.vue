<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ProTable from '@/components/ProTable/index.vue'
import type { ColumnDef } from '@/components/ProTable/index.vue'
import { useTable } from '@/composables/useTable'
import CouponPicker from '@/components/CouponPicker.vue'
import {
  listCouponRedeemBatches,
  createCouponRedeemBatch,
  type CouponRedeemBatch,
} from '@/api/marketing'
import { formatTime } from '@/utils/format'

const searchForm = ref({})
const dialogVisible = ref(false)
const genForm = ref<{ template_id: string | null; count: number }>({ template_id: null, count: 100 })

const { list, total, page, pageSize, loading, fetch } = useTable<CouponRedeemBatch>((p) =>
  listCouponRedeemBatches({ ...p, ...searchForm.value })
)

const columns: ColumnDef[] = [
  { label: '批次 ID', prop: 'id', width: 130 },
  { label: '券模板', prop: 'template_name', minWidth: 200 },
  { label: '生成数量', prop: 'count', width: 100, align: 'right' },
  { label: '已使用', prop: 'used_count', width: 100, align: 'right' },
  { label: '导出', slot: 'csv', width: 110 },
  { label: '创建时间', prop: 'created_at', width: 160, formatter: (r) => formatTime(r.created_at) },
]

async function generate() {
  if (!genForm.value.template_id) {
    ElMessage.warning('请选择券模板')
    return
  }
  if (genForm.value.count <= 0 || genForm.value.count > 100000) {
    ElMessage.warning('数量必须在 1 ~ 100000 之间')
    return
  }
  await createCouponRedeemBatch({
    template_id: genForm.value.template_id,
    count: genForm.value.count,
  })
  ElMessage.success('已提交，生成完成后可下载 CSV')
  dialogVisible.value = false
  fetch(searchForm.value)
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
      <template #toolbar>
        <el-button type="primary" @click="dialogVisible = true">生成兑换码</el-button>
      </template>

      <template #csv="{ row }">
        <el-link v-if="row.csv_url" :href="row.csv_url" target="_blank" type="primary">下载 CSV</el-link>
        <span v-else>生成中</span>
      </template>
    </ProTable>

    <el-dialog v-model="dialogVisible" title="生成兑换码" width="480px">
      <el-form label-width="100px">
        <el-form-item label="券模板">
          <CouponPicker v-model="genForm.template_id" />
        </el-form-item>
        <el-form-item label="生成数量">
          <el-input-number v-model="genForm.count" :min="1" :max="100000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="generate">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>
