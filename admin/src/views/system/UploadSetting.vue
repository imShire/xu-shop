<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getUploadSettings, probeUploadSettings, testUploadSettings, updateUploadSettings } from '@/api/system'
import type { UploadSettings } from '@/types'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const probing = ref(false)
const formRef = ref<FormInstance>()
const probeInputRef = ref<HTMLInputElement>()
const probeResult = ref('')

const form = reactive<UploadSettings>({
  driver: 'local',
  public_base_url: 'http://localhost:8080/uploads',
  max_size_mb: 10,
  local_dir: 'uploads',
  s3_vendor: 'generic',
  s3_endpoint: '',
  s3_region: '',
  s3_bucket: '',
  s3_prefix: '',
  s3_force_path_style: false,
  s3_access_key: '',
  s3_secret_key: '',
  s3_access_key_set: false,
  s3_secret_key_set: false,
})

const canEdit = computed(() => auth.isSuperAdmin || auth.perms.includes('system.upload.edit'))

const rules: FormRules = {
  driver: [{ required: true, message: '请选择上传驱动', trigger: 'change' }],
  public_base_url: [{ required: true, message: '请填写上传访问地址', trigger: 'blur' }],
  max_size_mb: [{ required: true, message: '请填写上传大小限制', trigger: 'change' }],
  local_dir: [{
    validator: (_rule, value, callback) => {
      if (form.driver === 'local' && !value) return callback(new Error('请填写本地上传目录'))
      callback()
    },
    trigger: 'blur',
  }],
  s3_endpoint: [{
    validator: (_rule, value, callback) => {
      if (form.driver === 's3' && !value) return callback(new Error('请填写 S3 Endpoint'))
      callback()
    },
    trigger: 'blur',
  }],
  s3_bucket: [{
    validator: (_rule, value, callback) => {
      if (form.driver === 's3' && !value) return callback(new Error('请填写 S3 Bucket'))
      callback()
    },
    trigger: 'blur',
  }],
  s3_access_key: [{
    validator: (_rule, value, callback) => {
      if (form.driver === 's3' && !value && !form.s3_access_key_set) return callback(new Error('请填写 S3 Access Key'))
      callback()
    },
    trigger: 'blur',
  }],
  s3_secret_key: [{
    validator: (_rule, value, callback) => {
      if (form.driver === 's3' && !value && !form.s3_secret_key_set) return callback(new Error('请填写 S3 Secret Key'))
      callback()
    },
    trigger: 'blur',
  }],
}

async function loadData() {
  loading.value = true
  try {
    const res = await getUploadSettings()
    Object.assign(form, res, { s3_access_key: '', s3_secret_key: '' })
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  await formRef.value?.validate()
  saving.value = true
  try {
    await updateUploadSettings(toPayload())
    ElMessage.success('保存成功')
    await loadData()
  } finally {
    saving.value = false
  }
}

async function handleTest() {
  await formRef.value?.validate()
  testing.value = true
  try {
    await testUploadSettings(toPayload())
    ElMessage.success('连接测试成功')
  } finally {
    testing.value = false
  }
}

async function handleProbeChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  probing.value = true
  probeResult.value = ''
  try {
    const res = await probeUploadSettings(file)
    probeResult.value = res.url
    ElMessage.success('测试上传成功')
  } finally {
    probing.value = false
    input.value = ''
  }
}

function openProbePicker() {
  probeInputRef.value?.click()
}

function toPayload() {
  return {
    driver: form.driver,
    public_base_url: form.public_base_url,
    max_size_mb: Number(form.max_size_mb),
    local_dir: form.local_dir,
    s3_vendor: form.s3_vendor,
    s3_endpoint: form.s3_endpoint,
    s3_region: form.s3_region,
    s3_bucket: form.s3_bucket,
    s3_access_key: form.s3_access_key,
    s3_secret_key: form.s3_secret_key,
    s3_prefix: form.s3_prefix,
    s3_force_path_style: form.s3_force_path_style,
  }
}

onMounted(loadData)
</script>

<template>
  <div class="page-card" v-loading="loading">
    <div class="page-head">
      <div>
        <h3>上传设置</h3>
        <p>统一管理后台图片上传方式。七牛请走 S3 兼容模式填写 endpoint 与 bucket。</p>
      </div>
      <div class="page-actions">
        <el-button :loading="testing" @click="handleTest" :disabled="!canEdit">测试连接</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave" :disabled="!canEdit">保存配置</el-button>
      </div>
    </div>

    <el-alert
      v-if="!canEdit"
      title="当前账号只有查看权限，不能修改上传设置。"
      type="warning"
      show-icon
      :closable="false"
      style="margin-bottom: 16px"
    />

    <el-form ref="formRef" :model="form" :rules="rules" label-width="140px" class="upload-form">
      <el-card shadow="never" class="setting-card">
        <template #header>基础配置</template>
        <el-form-item label="上传驱动" prop="driver">
          <el-radio-group v-model="form.driver">
            <el-radio value="local">本地上传</el-radio>
            <el-radio value="s3">S3 兼容上传</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="访问前缀" prop="public_base_url">
          <el-input v-model="form.public_base_url" placeholder="例如 http://localhost:8080/uploads 或 https://cdn.example.com" style="width: 320px" />
        </el-form-item>
        <el-form-item label="大小限制" prop="max_size_mb">
          <el-input-number v-model="form.max_size_mb" :min="1" :max="50" style="width: 120px" />
          <span class="field-tip">单位 MB，当前只用于图片上传。</span>
        </el-form-item>
      </el-card>

      <el-card v-if="form.driver === 'local'" shadow="never" class="setting-card">
        <template #header>本地上传</template>
        <el-form-item label="保存目录" prop="local_dir">
          <el-input v-model="form.local_dir" placeholder="例如 uploads" style="width: 320px" />
        </el-form-item>
        <div class="hint-block">
          本地模式要求你的反向代理或 API 服务对外提供 `/uploads/*` 访问路径。
        </div>
      </el-card>

      <el-card v-else shadow="never" class="setting-card">
        <template #header>S3 / 七牛兼容</template>
        <el-form-item label="存储厂商">
          <el-select v-model="form.s3_vendor" style="width: 220px">
            <el-option label="通用 S3" value="generic" />
            <el-option label="七牛 Kodo S3" value="qiniu" />
          </el-select>
        </el-form-item>
        <el-form-item label="Endpoint" prop="s3_endpoint">
          <el-input v-model="form.s3_endpoint" placeholder="例如 https://s3-cn-east-1.qiniucs.com" style="width: 320px" />
        </el-form-item>
        <el-form-item label="Region">
          <el-input v-model="form.s3_region" placeholder="可选，例如 z0 / us-east-1" style="width: 320px" />
        </el-form-item>
        <el-form-item label="Bucket" prop="s3_bucket">
          <el-input v-model="form.s3_bucket" style="width: 320px" />
        </el-form-item>
        <el-form-item label="Access Key" prop="s3_access_key">
          <el-input v-model="form.s3_access_key" :placeholder="form.s3_access_key_set ? '已配置，留空则保持不变' : '请输入 Access Key'" style="width: 320px" />
        </el-form-item>
        <el-form-item label="Secret Key" prop="s3_secret_key">
          <el-input v-model="form.s3_secret_key" type="password" show-password :placeholder="form.s3_secret_key_set ? '已配置，留空则保持不变' : '请输入 Secret Key'" style="width: 320px" />
        </el-form-item>
        <el-form-item label="对象前缀">
          <el-input v-model="form.s3_prefix" placeholder="可选，例如 media" style="width: 320px" />
        </el-form-item>
        <el-form-item label="Path Style">
          <el-switch v-model="form.s3_force_path_style" />
          <span class="field-tip">MinIO、部分私有 S3 兼容服务通常需要开启。</span>
        </el-form-item>
      </el-card>

      <el-card shadow="never" class="setting-card">
        <template #header>上传调试</template>
        <div class="probe-panel">
          <div class="probe-copy">
            用当前生效配置实际上传一张图片，直接验证切换后的上传路径、访问前缀和回源是否可用。
          </div>
          <div class="probe-actions">
            <input
              ref="probeInputRef"
              type="file"
              accept="image/*"
              style="display: none"
              @change="handleProbeChange"
            />
            <el-button :loading="probing" @click="openProbePicker" :disabled="!canEdit">
              选择图片并测试上传
            </el-button>
          </div>
          <div v-if="probeResult" class="probe-result">
            <el-input :model-value="probeResult" readonly />
            <a :href="probeResult" target="_blank" rel="noreferrer">打开上传结果</a>
            <img :src="probeResult" alt="probe preview" class="probe-preview" />
          </div>
        </div>
      </el-card>
    </el-form>
  </div>
</template>

<style scoped lang="scss">
.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;

  h3 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 6px;
  }

  p {
    color: var(--text-secondary);
    line-height: 1.6;
    max-width: 720px;
  }
}

.page-actions {
  display: flex;
  gap: 8px;
}

.upload-form {
  display: grid;
  gap: 16px;
}

.setting-card {
  border-radius: 10px;
}

.field-tip {
  margin-left: 12px;
  color: var(--text-secondary);
  font-size: 12px;
}

.hint-block {
  padding: 12px 14px;
  border-radius: 8px;
  background: #fffbeb;
  color: #8a5b00;
  font-size: 13px;
  line-height: 1.6;
}

.probe-panel {
  display: grid;
  gap: 14px;
}

.probe-copy {
  color: var(--text-secondary);
  line-height: 1.6;
}

.probe-actions {
  display: flex;
  gap: 8px;
}

.probe-result {
  display: grid;
  gap: 10px;

  a {
    color: var(--el-color-primary);
    width: fit-content;
  }
}

.probe-preview {
  max-width: 320px;
  max-height: 240px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: #fff;
  object-fit: contain;
  padding: 8px;
}
</style>
