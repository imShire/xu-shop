<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, updateSettings } from '@/api/system'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const canEdit = computed(() => auth.isSuperAdmin || auth.perms.includes('system.setting.edit'))

type Group = 'basic' | 'wxpay' | 'wxlogin' | 'qywx' | 'kdniao' | 'sms' | 'security'

const tabs: { name: Group; label: string }[] = [
  { name: 'basic', label: '基本设置' },
  { name: 'wxpay', label: '微信支付' },
  { name: 'wxlogin', label: '微信登录' },
  { name: 'qywx', label: '企业微信' },
  { name: 'kdniao', label: '快递鸟' },
  { name: 'sms', label: '短信' },
  { name: 'security', label: '安全设置' },
]

// Secret fields per group — when backend returns "", it means "value is set but masked"
const secretFields: Record<Group, string[]> = {
  basic: [],
  wxpay: ['api_v3_key', 'cert_pem', 'key_pem'],
  wxlogin: ['mp_app_secret'],
  qywx: ['app_secret', 'robot_webhook'],
  kdniao: ['api_key'],
  sms: ['secret_key'],
  security: [],
}

// Textarea fields (large content)
const textareaFields: Record<Group, string[]> = {
  basic: [],
  wxpay: ['cert_pem', 'key_pem'],
  wxlogin: [],
  qywx: [],
  kdniao: [],
  sms: [],
  security: [],
}

// Forms per group
const forms = reactive<Record<Group, Record<string, string>>>({
  basic: { shop_name: '', shop_logo: '', contact_phone: '', icp_no: '' },
  wxpay: { mch_id: '', app_id: '', api_v3_key: '', serial_no: '', cert_pem: '', key_pem: '' },
  wxlogin: { mp_app_id: '', mp_app_secret: '' },
  qywx: { corp_id: '', agent_id: '', app_secret: '', robot_webhook: '' },
  kdniao: { e_business_id: '', api_key: '', env: 'sandbox' },
  sms: { provider: 'aliyun', access_key: '', secret_key: '', sign_name: '', verify_template: '' },
  security: { session_hours: '24', max_login_attempts: '5', admin_pw_min_len: '12' },
})

// Track which groups have been loaded and which are currently being saved
const loaded = reactive<Record<Group, boolean>>({
  basic: false, wxpay: false, wxlogin: false, qywx: false, kdniao: false, sms: false, security: false,
})
const loadingGroup = ref<Group | null>(null)
const savingGroup = ref<Group | null>(null)

// Track which secret fields already have a value on the server (mask display)
const secretSet = reactive<Record<Group, Record<string, boolean>>>({
  basic: {},
  wxpay: { api_v3_key: false, cert_pem: false, key_pem: false },
  wxlogin: { mp_app_secret: false },
  qywx: { app_secret: false, robot_webhook: false },
  kdniao: { api_key: false },
  sms: { secret_key: false },
  security: {},
})

const activeTab = ref<Group>('basic')

async function loadGroup(group: Group) {
  if (loaded[group]) return
  loadingGroup.value = group
  try {
    const data = await getSettings(group)
    const secrets = secretFields[group]
    const newForm: Record<string, string> = { ...forms[group] }
    for (const [k, v] of Object.entries(data)) {
      if (secrets.includes(k)) {
        // Backend returns "" for masked secrets
        secretSet[group][k] = v !== '' // truly configured if backend returned non-empty
        newForm[k] = '' // always keep local field empty so user must retype to change
      } else {
        newForm[k] = v
      }
    }
    Object.assign(forms[group], newForm)
    loaded[group] = true
  } catch {
    ElMessage.error(`加载 ${group} 配置失败`)
  } finally {
    loadingGroup.value = null
  }
}

function handleTabChange(name: string) {
  activeTab.value = name as Group
  loadGroup(name as Group)
}

async function handleSave(group: Group) {
  savingGroup.value = group
  try {
    const payload: Record<string, string> = {}
    const secrets = secretFields[group]
    for (const [k, v] of Object.entries(forms[group])) {
      // Skip empty secret fields — leave backend value unchanged
      if (secrets.includes(k) && v === '') continue
      payload[k] = v
    }
    await updateSettings(group, payload)
    ElMessage.success('保存成功')
    // Reload to sync
    loaded[group] = false
    await loadGroup(group)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    savingGroup.value = null
  }
}

// Initial load for default tab
loadGroup('basic')
</script>

<template>
  <div class="page-card">
    <div class="page-head">
      <div>
        <h3>系统设置</h3>
        <p>管理店铺基本信息、第三方服务接入与安全策略。</p>
      </div>
    </div>

    <el-alert
      v-if="!canEdit"
      title="当前账号只有查看权限，不能修改系统设置。"
      type="warning"
      show-icon
      :closable="false"
      style="margin-bottom: 16px"
    />

    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <el-tab-pane
        v-for="tab in tabs"
        :key="tab.name"
        :label="tab.label"
        :name="tab.name"
      >
        <div
          v-loading="loadingGroup === tab.name"
          class="tab-content"
        >
          <!-- 基本设置 -->
          <template v-if="tab.name === 'basic'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="店铺名称">
                <el-input v-model="forms.basic.shop_name" placeholder="请输入店铺名称" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="店铺 Logo URL">
                <el-input v-model="forms.basic.shop_logo" placeholder="请输入图片地址" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="联系电话">
                <el-input v-model="forms.basic.contact_phone" placeholder="请输入联系电话" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="ICP 备案号">
                <el-input v-model="forms.basic.icp_no" placeholder="例如 粤ICP备XXXXXXXX号" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'basic'" @click="handleSave('basic')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>

          <!-- 微信支付 -->
          <template v-else-if="tab.name === 'wxpay'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="商户号 (mch_id)">
                <el-input v-model="forms.wxpay.mch_id" placeholder="请输入商户号" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="AppID">
                <el-input v-model="forms.wxpay.app_id" placeholder="微信支付关联 AppID" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="APIv3 密钥">
                <el-input
                  v-model="forms.wxpay.api_v3_key"
                  :placeholder="secretSet.wxpay.api_v3_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 APIv3 密钥'"
                  :disabled="!canEdit"
                  show-password
                  style="width: 320px"
                />
              </el-form-item>
              <el-form-item label="证书序列号">
                <el-input v-model="forms.wxpay.serial_no" placeholder="请输入证书序列号" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="cert.pem">
                <el-input
                  v-model="forms.wxpay.cert_pem"
                  type="textarea"
                  :rows="4"
                  :placeholder="secretSet.wxpay.cert_pem ? '●●●●●● 已配置，留空则保持不变' : '请粘贴 apiclient_cert.pem 内容'"
                  :disabled="!canEdit"
                />
              </el-form-item>
              <el-form-item label="key.pem">
                <el-input
                  v-model="forms.wxpay.key_pem"
                  type="textarea"
                  :rows="4"
                  :placeholder="secretSet.wxpay.key_pem ? '●●●●●● 已配置，留空则保持不变' : '请粘贴 apiclient_key.pem 内容'"
                  :disabled="!canEdit"
                />
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'wxpay'" @click="handleSave('wxpay')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>

          <!-- 微信登录 -->
          <template v-else-if="tab.name === 'wxlogin'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="小程序 AppID">
                <el-input v-model="forms.wxlogin.mp_app_id" placeholder="微信小程序 AppID" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="小程序 AppSecret">
                <el-input
                  v-model="forms.wxlogin.mp_app_secret"
                  :placeholder="secretSet.wxlogin.mp_app_secret ? '●●●●●● 已配置，留空则保持不变' : '请输入 AppSecret'"
                  :disabled="!canEdit"
                  show-password
                  style="width: 320px"
                />
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'wxlogin'" @click="handleSave('wxlogin')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>

          <!-- 企业微信 -->
          <template v-else-if="tab.name === 'qywx'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="企业 ID (corp_id)">
                <el-input v-model="forms.qywx.corp_id" placeholder="企业微信 corpid" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="应用 ID (agent_id)">
                <el-input v-model="forms.qywx.agent_id" placeholder="企业微信应用 agentid" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="应用 Secret">
                <el-input
                  v-model="forms.qywx.app_secret"
                  :placeholder="secretSet.qywx.app_secret ? '●●●●●● 已配置，留空则保持不变' : '请输入应用 Secret'"
                  :disabled="!canEdit"
                  show-password
                  style="width: 320px"
                />
              </el-form-item>
              <el-form-item label="机器人 Webhook">
                <el-input
                  v-model="forms.qywx.robot_webhook"
                  :placeholder="secretSet.qywx.robot_webhook ? '●●●●●● 已配置，留空则保持不变' : 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...'"
                  :disabled="!canEdit"
                  show-password
                />
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'qywx'" @click="handleSave('qywx')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>

          <!-- 快递鸟 -->
          <template v-else-if="tab.name === 'kdniao'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="电商 ID">
                <el-input v-model="forms.kdniao.e_business_id" placeholder="快递鸟电商 EBusinessID" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="API Key">
                <el-input
                  v-model="forms.kdniao.api_key"
                  :placeholder="secretSet.kdniao.api_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 API Key'"
                  :disabled="!canEdit"
                  show-password
                  style="width: 320px"
                />
              </el-form-item>
              <el-form-item label="环境">
                <el-select v-model="forms.kdniao.env" style="width: 200px" :disabled="!canEdit">
                  <el-option label="沙箱环境" value="sandbox" />
                  <el-option label="生产环境" value="production" />
                </el-select>
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'kdniao'" @click="handleSave('kdniao')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>

          <!-- 短信 -->
          <template v-else-if="tab.name === 'sms'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="短信服务商">
                <el-select v-model="forms.sms.provider" style="width: 200px" :disabled="!canEdit">
                  <el-option label="阿里云" value="aliyun" />
                  <el-option label="腾讯云" value="tencent" />
                </el-select>
              </el-form-item>
              <el-form-item label="Access Key">
                <el-input v-model="forms.sms.access_key" placeholder="请输入 Access Key ID" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="Secret Key">
                <el-input
                  v-model="forms.sms.secret_key"
                  :placeholder="secretSet.sms.secret_key ? '●●●●●● 已配置，留空则保持不变' : '请输入 Secret Key'"
                  :disabled="!canEdit"
                  show-password
                  style="width: 320px"
                />
              </el-form-item>
              <el-form-item label="短信签名">
                <el-input v-model="forms.sms.sign_name" placeholder="例如 【xu-shop】" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item label="验证码模板 ID">
                <el-input v-model="forms.sms.verify_template" placeholder="短信验证码模板 ID" :disabled="!canEdit" style="width: 320px" />
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'sms'" @click="handleSave('sms')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>

          <!-- 安全设置 -->
          <template v-else-if="tab.name === 'security'">
            <el-form label-width="160px" class="setting-form">
              <el-form-item label="会话时长（小时）">
                <el-input-number
                  v-model.number="forms.security.session_hours"
                  :min="1"
                  :max="720"
                  :disabled="!canEdit"
                  style="width: 120px"
                />
                <span class="field-tip">管理员登录 Token 有效期，单位小时。</span>
              </el-form-item>
              <el-form-item label="最大登录失败次数">
                <el-input-number
                  v-model.number="forms.security.max_login_attempts"
                  :min="1"
                  :max="100"
                  :disabled="!canEdit"
                  style="width: 120px"
                />
                <span class="field-tip">达到上限后锁定账户。</span>
              </el-form-item>
              <el-form-item label="密码最短长度">
                <el-input-number
                  v-model.number="forms.security.admin_pw_min_len"
                  :min="8"
                  :max="64"
                  :disabled="!canEdit"
                  style="width: 120px"
                />
                <span class="field-tip">后台管理员密码最少字符数（≥12 推荐）。</span>
              </el-form-item>
              <el-form-item v-if="canEdit">
                <el-button type="primary" :loading="savingGroup === 'security'" @click="handleSave('security')">保存</el-button>
              </el-form-item>
            </el-form>
          </template>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped lang="scss">
.page-head {
  margin-bottom: 16px;

  h3 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 6px;
  }

  p {
    color: var(--text-secondary);
    line-height: 1.6;
  }
}

.tab-content {
  padding-top: 16px;
  min-height: 200px;
}

.setting-form {
  max-width: 680px;
}

.field-tip {
  margin-left: 12px;
  color: var(--text-secondary);
  font-size: 12px;
}
</style>
