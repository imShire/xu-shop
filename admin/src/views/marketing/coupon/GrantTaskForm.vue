<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import CouponPicker from '@/components/CouponPicker.vue'
import AudienceFilterBuilder from '@/components/AudienceFilterBuilder.vue'
import type { Group } from '@/components/AudienceGroupNode.vue'
import { createCouponGrantTask } from '@/api/marketing'
import { previewAudience } from '@/api/tag'

const router = useRouter()
const templateId = ref<string | null>(null)
const sourceMode = ref<'audience' | 'phones'>('audience')
const audience = ref<Group>({ op: 'and', conditions: [] })
const phones = ref('')
const previewCount = ref<number | null>(null)

async function preview() {
  if (sourceMode.value === 'phones') {
    const arr = phones.value
      .split(/[\n,]/)
      .map((s) => s.trim())
      .filter(Boolean)
    previewCount.value = arr.length
    return
  }
  const res = await previewAudience({ audience: audience.value })
  previewCount.value = res.total
}

async function submit() {
  if (!templateId.value) {
    ElMessage.warning('请选择券模板')
    return
  }
  await ElMessageBox.confirm('确认创建发放任务？任务一旦提交将异步发送', '提示', { type: 'warning' })
  const filter =
    sourceMode.value === 'audience'
      ? { audience: audience.value }
      : {
          phones: phones.value
            .split(/[\n,]/)
            .map((s) => s.trim())
            .filter(Boolean),
        }
  await createCouponGrantTask({ template_id: templateId.value, filter })
  ElMessage.success('已创建')
  router.push('/marketing/coupon/grants')
}
</script>

<template>
  <div class="page-card">
    <el-page-header content="新建定向发放任务" @back="router.back()" style="margin-bottom: 16px" />

    <el-form label-width="120px">
      <el-form-item label="券模板" required>
        <CouponPicker v-model="templateId" style="max-width: 480px" />
      </el-form-item>

      <el-form-item label="人群来源">
        <el-radio-group v-model="sourceMode">
          <el-radio value="audience">标签人群</el-radio>
          <el-radio value="phones">手机号批量</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="sourceMode === 'audience'" label="人群条件">
        <AudienceFilterBuilder v-model="audience" />
      </el-form-item>

      <el-form-item v-else label="手机号">
        <el-input
          v-model="phones"
          type="textarea"
          :rows="6"
          placeholder="每行一个手机号，或英文逗号分隔"
          style="max-width: 600px"
        />
      </el-form-item>

      <el-form-item label="预估命中">
        <el-button @click="preview">预估</el-button>
        <span v-if="previewCount !== null" style="margin-left: 12px">
          预计发放 <b>{{ previewCount }}</b> 张
        </span>
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="submit">提交任务</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
