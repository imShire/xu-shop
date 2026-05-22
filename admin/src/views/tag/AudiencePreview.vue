<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import AudienceFilterBuilder from '@/components/AudienceFilterBuilder.vue'
import { previewAudience, type AudiencePreviewResp } from '@/api/tag'

const filter = ref<any>({ op: 'and', conditions: [] })
const result = ref<AudiencePreviewResp | null>(null)
const loading = ref(false)

async function preview() {
  loading.value = true
  try {
    result.value = await previewAudience(filter.value)
  } catch (e) {
    ElMessage.error('预估失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page-card">
    <el-alert type="info" :closable="false" style="margin-bottom: 12px">
      构建标签 + 行为条件组合，预估命中人数；可用于优惠券定向发放、召回活动选人前快速验证规则
    </el-alert>

    <AudienceFilterBuilder v-model="filter" />

    <div style="margin-top: 16px; text-align: center">
      <el-button type="primary" :loading="loading" @click="preview">预估命中人数</el-button>
    </div>

    <el-card v-if="result" shadow="never" style="margin-top: 16px">
      <el-descriptions title="预估结果" :column="2" border>
        <el-descriptions-item label="命中人数">
          <span style="font-size: 28px; font-weight: 600; color: var(--el-color-primary)">
            {{ result.total }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="样本量">{{ result.sample?.length || 0 }}</el-descriptions-item>
      </el-descriptions>

      <div style="margin-top: 16px" v-if="result.sample?.length">
        <h4>样本用户（最多 20 个）</h4>
        <el-table :data="result.sample" size="small" border>
          <el-table-column prop="user_id" label="用户 ID" width="160" />
          <el-table-column prop="nickname" label="昵称" />
          <el-table-column prop="phone" label="手机号" width="140" />
        </el-table>
      </div>
    </el-card>
  </div>
</template>
