<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { getCaptcha } from '@/api/account'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaId = ref('')
const captchaB64 = ref('')
const captchaError = ref(false)

const signalCards = [
  { label: '待发货订单', value: '23', detail: '高峰时段 3 分钟内响应' },
  { label: '库存预警', value: '05', detail: '补货队列已同步' },
  { label: '私域转化', value: '+18%', detail: '近 24 小时渠道增长' },
]

const operationRows = [
  { module: '订单', metric: '1,284', trend: '+12.6%', tone: 'up' },
  { module: '客户', metric: '486', trend: '+08.4%', tone: 'up' },
  { module: '库存', metric: '32', trend: '-02.1%', tone: 'down' },
  { module: '投放', metric: '74', trend: '+05.7%', tone: 'up' },
]

const checkpoints = [
  '验证码动态刷新，降低撞库风险',
  '角色权限与路由守卫统一校验',
  '登录成功后直达工作台',
]

const form = ref({
  username: '',
  password: '',
  captcha_code: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha_code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

async function loadCaptcha() {
  captchaError.value = false
  try {
    const res = await getCaptcha()
    captchaId.value = res.captcha_id
    captchaB64.value = res.captcha_b64
  } catch {
    captchaB64.value = ''
    captchaError.value = true
  }
}

async function handleLogin() {
  await formRef.value?.validate()
  loading.value = true
  try {
    await authStore.login({
      ...form.value,
      captcha_id: captchaId.value,
    })
    ElMessage.success('登录成功')
    router.push('/workbench')
  } catch {
    await loadCaptcha()
    form.value.captcha_code = ''
  } finally {
    loading.value = false
  }
}

onMounted(loadCaptcha)
</script>

<template>
  <div class="login-shell">
    <div class="login-shell__noise" />
    <div class="login-shell__beam login-shell__beam--left" />
    <div class="login-shell__beam login-shell__beam--right" />

    <div class="login-grid">
      <section class="market-stage">
        <div class="market-stage__masthead">
          <div class="stage-logo">
            <span class="stage-logo__mark" aria-hidden="true">
              <span class="stage-logo__glyph">
                <i class="stage-logo__tile stage-logo__tile--top" />
                <i class="stage-logo__tile stage-logo__tile--right" />
                <i class="stage-logo__tile stage-logo__tile--bottom" />
                <i class="stage-logo__tile stage-logo__tile--left" />
                <i class="stage-logo__core" />
              </span>
            </span>
            <div>
              <p>xu-shop</p>
              <span>管理后台</span>
            </div>
          </div>
          <div class="stage-status">
            <span class="stage-status__dot" />
            安全通道已开启
          </div>
        </div>

        <div class="market-stage__intro">
          <p class="market-stage__eyebrow">私域电商运营中枢</p>
          <h1>商品、订单、库存与私域运营，在一个后台协同处理。</h1>
          <p class="market-stage__lead">
            支持商品管理、订单处理、库存预警与客户运营协同处理，帮助团队在同一后台完成日常业务流转。
          </p>
        </div>

        <div class="signal-strip">
          <article
            v-for="card in signalCards"
            :key="card.label"
            class="signal-card"
          >
            <span>{{ card.label }}</span>
            <strong>{{ card.value }}</strong>
            <small>{{ card.detail }}</small>
          </article>
        </div>

        <div class="ops-board">
          <div class="ops-board__header">
            <div>
              <p>运行概览</p>
              <span>核心模块状态与权限入口一屏可见</span>
            </div>
            <div class="ops-board__tag">
              <el-icon><DataLine /></el-icon>
              实时同步
            </div>
          </div>

          <div class="ops-board__grid">
            <div class="ops-board__ticker">
              <div class="ticker-head">
                <span>模块</span>
                <span>负载</span>
                <span>日环比</span>
              </div>
              <div
                v-for="row in operationRows"
                :key="row.module"
                class="ticker-row"
              >
                <span class="ticker-row__module">{{ row.module }}</span>
                <strong>{{ row.metric }}</strong>
                <span :class="['ticker-row__trend', `ticker-row__trend--${row.tone}`]">
                  {{ row.trend }}
                </span>
              </div>
            </div>

            <div class="ops-board__rail">
              <div class="rail-card">
                <span>权限校验</span>
                <strong>角色守卫已生效</strong>
                <p>角色权限与页面准入已经联动，越权访问会被路由守卫阻断。</p>
              </div>
              <div class="rail-card">
                <span>登录风控</span>
                <strong>验证码已接入</strong>
                <p>每次登录都会刷新验证码，防止后台入口被脚本撞库。</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="auth-panel">
        <div class="auth-panel__header">
          <div class="auth-panel__badge">
            <el-icon><Key /></el-icon>
            管理员入口
          </div>
          <h2>登录管理后台</h2>
          <p>
            仅限已授权运营与管理账号登录。通过验证后直接进入工作台，关键操作全程留痕。
          </p>
        </div>

        <el-form
          ref="formRef"
          class="auth-form"
          :model="form"
          :rules="rules"
          label-position="top"
          size="large"
          @keyup.enter="handleLogin"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              autocomplete="username"
            >
              <template #prefix><el-icon><User /></el-icon></template>
            </el-input>
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              show-password
              autocomplete="current-password"
            >
              <template #prefix><el-icon><Key /></el-icon></template>
            </el-input>
          </el-form-item>

          <el-form-item label="验证码" prop="captcha_code">
            <div class="captcha-row">
              <el-input
                v-model="form.captcha_code"
                placeholder="输入图形验证码"
              >
                <template #prefix><el-icon><Connection /></el-icon></template>
              </el-input>

              <button
                v-if="captchaB64 && !captchaError"
                type="button"
                class="captcha-panel"
                title="点击刷新验证码"
                @click="loadCaptcha"
              >
                <img
                  :src="captchaB64"
                  alt="验证码"
                  @error="captchaError = true"
                >
              </button>

              <button
                v-else
                type="button"
                class="captcha-panel captcha-panel--fallback"
                @click="loadCaptcha"
              >
                <el-icon><Warning /></el-icon>
                {{ captchaError ? '获取失败，点击重试' : '刷新验证码' }}
              </button>
            </div>
          </el-form-item>

          <el-button
            type="primary"
            class="submit-button"
            :loading="loading"
            @click="handleLogin"
          >
            进入后台
            <el-icon><Right /></el-icon>
          </el-button>
        </el-form>

        <div class="auth-panel__footer">
          <div class="footer-note">
            <span class="footer-note__label">登录校验</span>
            <ul>
              <li v-for="item in checkpoints" :key="item">
                <el-icon><CircleCheckFilled /></el-icon>
                <span>{{ item }}</span>
              </li>
            </ul>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped lang="scss">
.login-shell {
  --bg: #f6f1e7;
  --bg-soft: #efe6d7;
  --panel: rgba(255, 252, 246, 0.88);
  --panel-strong: rgba(255, 255, 255, 0.96);
  --line: rgba(22, 26, 31, 0.08);
  --line-strong: rgba(240, 185, 11, 0.32);
  --text: #161a20;
  --muted: #616978;
  --muted-soft: #8a919d;
  --primary: #f0b90b;
  --primary-soft: rgba(240, 185, 11, 0.14);
  --success: #0ecb81;
  --danger: #f6465d;
  --radius-xl: 28px;
  --radius-lg: 20px;
  --radius-md: 14px;
  --shadow-panel: 0 30px 70px rgba(132, 113, 70, 0.16);
  --display: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Noto Sans CJK SC', sans-serif;
  --body: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Noto Sans CJK SC', sans-serif;
  --numeric: 'DIN Alternate', 'Bahnschrift', 'SFMono-Regular', 'Roboto Mono', monospace;

  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at top left, rgba(240, 185, 11, 0.22), transparent 28%),
    radial-gradient(circle at 85% 18%, rgba(240, 185, 11, 0.16), transparent 24%),
    linear-gradient(135deg, #fffdf7 0%, #f7f0e3 44%, #f0e5d3 100%);
  color: var(--text);
  font-family: var(--body);
}

.login-shell__noise,
.login-shell__beam {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.login-shell__noise {
  opacity: 0.2;
  background-image:
    linear-gradient(rgba(17, 24, 39, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(17, 24, 39, 0.035) 1px, transparent 1px);
  background-size: 72px 72px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.9), transparent 92%);
}

.login-shell__beam {
  filter: blur(24px);
  opacity: 0.5;
  animation: beamFloat 12s ease-in-out infinite;
}

.login-shell__beam--left {
  inset: 18% auto auto -8%;
  width: 320px;
  height: 320px;
  background: radial-gradient(circle, rgba(240, 185, 11, 0.24), transparent 70%);
}

.login-shell__beam--right {
  inset: auto -10% 6% auto;
  width: 360px;
  height: 360px;
  background: radial-gradient(circle, rgba(240, 185, 11, 0.18), transparent 68%);
  animation-delay: -4s;
}

.login-grid {
  position: relative;
  z-index: 1;
  height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(360px, 440px);
  gap: 24px;
  align-items: center;
  padding: 20px;
}

.market-stage,
.auth-panel {
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}

.market-stage {
  height: calc(100vh - 40px);
  display: grid;
  grid-template-rows: auto auto auto minmax(0, 1fr);
  align-content: start;
  gap: 16px;
  padding: 24px 26px;
  border: 1px solid var(--line);
  border-radius: var(--radius-xl);
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.9), transparent 45%),
    linear-gradient(180deg, rgba(255, 250, 242, 0.95), rgba(249, 242, 231, 0.95));
  box-shadow: var(--shadow-panel);
  animation: stageReveal 650ms ease-out both;
}

.market-stage__masthead {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.stage-logo {
  display: flex;
  align-items: center;
  gap: 10px;

  p {
    font-family: var(--display);
    font-size: 21px;
    font-weight: 700;
    letter-spacing: 0;
    text-transform: none;
  }

  span {
    display: block;
    margin-top: 2px;
    color: var(--muted);
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: none;
  }
}

.stage-logo__mark {
  width: 38px;
  height: 38px;
  display: grid;
  place-items: center;
  border-radius: 12px;
  border: 1px solid rgba(22, 26, 31, 0.08);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9), rgba(245, 236, 218, 0.9));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.8),
    0 8px 18px rgba(240, 185, 11, 0.12);
}

.stage-logo__glyph {
  position: relative;
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
}

.stage-logo__tile,
.stage-logo__core {
  position: absolute;
  display: block;
}

.stage-logo__tile {
  width: 8px;
  height: 8px;
  border-radius: 3px;
  background: linear-gradient(135deg, #f7cf49 0%, #f0b90b 100%);
  box-shadow: 0 1px 2px rgba(143, 101, 0, 0.18);
}

.stage-logo__tile--top {
  top: 0;
  left: 0;
  right: 0;
  margin-inline: auto;
  transform: rotate(45deg);
}

.stage-logo__tile--right {
  top: 0;
  bottom: 0;
  right: 0;
  margin-block: auto;
  transform: rotate(45deg);
}

.stage-logo__tile--bottom {
  bottom: 0;
  left: 0;
  right: 0;
  margin-inline: auto;
  transform: rotate(45deg);
}

.stage-logo__tile--left {
  top: 0;
  bottom: 0;
  left: 0;
  margin-block: auto;
  transform: rotate(45deg);
}

.stage-logo__core {
  inset: 0;
  margin: auto;
  width: 7px;
  height: 7px;
  border-radius: 2px;
  background: #20252d;
  transform: rotate(45deg);
}

.stage-status {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.66);
  color: #59606d;
  font-size: 12px;
  letter-spacing: 0;
  text-transform: none;
}

.stage-status__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--success);
  box-shadow: 0 0 0 6px rgba(14, 203, 129, 0.14);
}

.market-stage__intro {
  max-width: 760px;
  margin: 4px 0 0;
}

.market-stage__eyebrow {
  margin-bottom: 10px;
  color: var(--primary);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: none;
}

.market-stage__intro h1 {
  font-family: var(--display);
  font-size: clamp(28px, 2.6vw, 40px);
  font-weight: 700;
  line-height: 1.16;
  letter-spacing: -0.03em;
  max-width: none;
  white-space: normal;
  text-wrap: pretty;
}

.market-stage__lead {
  max-width: 560px;
  margin-top: 14px;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.55;
}

.signal-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 0;
}

.signal-card {
  min-height: 96px;
  padding: 14px 16px;
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.86), rgba(251, 247, 239, 0.86));
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  animation: riseIn 700ms ease-out both;

  span,
  small {
    color: var(--muted);
  }

  span {
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: none;
  }

  strong {
    font-family: var(--numeric);
    font-size: clamp(24px, 2vw, 34px);
    font-weight: 700;
    line-height: 1;
  }

  small {
    font-size: 12px;
  }
}

.signal-card:nth-child(2) {
  animation-delay: 90ms;
}

.signal-card:nth-child(3) {
  animation-delay: 180ms;
}

.ops-board {
  border: 1px solid rgba(26, 31, 37, 0.06);
  border-radius: var(--radius-xl);
  background: linear-gradient(180deg, #12161d, #181d26);
  min-height: 0;
  padding: 16px;
  color: rgba(255, 255, 255, 0.92);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
  overflow: hidden;
}

.ops-board__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;

  p {
    color: var(--primary);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: none;
    margin-bottom: 4px;
  }

  span {
    color: rgba(255, 255, 255, 0.58);
    font-size: 12px;
  }
}

.ops-board__tag {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(240, 185, 11, 0.14);
  color: rgba(255, 255, 255, 0.92);
  font-size: 12px;
}

.ops-board__grid {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(240px, 0.9fr);
  gap: 12px;
  min-height: 0;
}

.ops-board__ticker,
.rail-card {
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.03);
}

.ops-board__ticker {
  padding: 8px 12px;
}

.ticker-head,
.ticker-row {
  display: grid;
  grid-template-columns: 1fr 1fr 0.8fr;
  align-items: center;
  gap: 12px;
}

.ticker-head {
  padding: 6px 0 8px;
  color: rgba(255, 255, 255, 0.36);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: none;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.ticker-row {
  padding: 10px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  animation: tickerPulse 6s ease-in-out infinite;

  &:last-child {
    border-bottom: 0;
  }

  strong {
    font-family: var(--numeric);
    font-size: 18px;
    font-weight: 700;
    justify-self: end;
  }
}

.ticker-row__module {
  font-weight: 600;
  letter-spacing: 0;
}

.ticker-row__trend {
  justify-self: end;
  font-family: var(--numeric);
  font-size: 12px;
  font-weight: 600;

  &--up {
    color: var(--success);
  }

  &--down {
    color: var(--danger);
  }
}

.ops-board__rail {
  display: grid;
  gap: 12px;
}

.rail-card {
  padding: 14px;

  span {
    display: block;
    color: var(--primary);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: none;
  }

  strong {
    display: block;
    margin: 8px 0 6px;
    font-family: var(--display);
    font-size: 18px;
    font-weight: 700;
    line-height: 1.1;
  }

  p {
    color: rgba(255, 255, 255, 0.6);
    font-size: 12px;
    line-height: 1.5;
  }
}

.auth-panel {
  position: relative;
  justify-self: end;
  width: min(100%, 440px);
  height: calc(100vh - 40px);
  padding: 22px 22px 18px;
  border: 1px solid rgba(22, 26, 31, 0.08);
  border-radius: var(--radius-xl);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.86), transparent 18%),
    linear-gradient(180deg, rgba(255, 252, 246, 0.98), rgba(255, 247, 237, 0.98));
  box-shadow: var(--shadow-panel);
  animation: panelFloat 750ms ease-out both;
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  overflow: hidden;
}

.auth-panel__header {
  margin-bottom: 16px;

  h2 {
    margin: 14px 0 8px;
    font-family: var(--display);
    font-size: 30px;
    font-weight: 700;
    line-height: 1;
    letter-spacing: -0.04em;
  }

  p {
    color: var(--muted);
    font-size: 13px;
    line-height: 1.55;
  }
}

.auth-panel__badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 999px;
  background: var(--primary-soft);
  color: #9a6d00;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: none;
}

.auth-form {
  align-self: start;

  :deep(.el-form-item) {
    margin-bottom: 16px;
  }

  :deep(.el-form-item__label) {
    margin-bottom: 8px;
    color: #3b414b;
    font-size: 12px;
    font-weight: 500;
    line-height: 1;
  }

  :deep(.el-input__wrapper) {
    min-height: 46px;
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.82);
    box-shadow: inset 0 0 0 1px rgba(22, 26, 31, 0.08);
    transition: box-shadow 0.2s ease, background-color 0.2s ease;
  }

  :deep(.el-input__prefix),
  :deep(.el-input__icon),
  :deep(.el-input__inner) {
    color: #212831;
  }

  :deep(.el-input__inner::placeholder) {
    color: #a0a6af;
  }

  :deep(.el-input__wrapper.is-focus) {
    background: #fff;
    box-shadow:
      inset 0 0 0 1px rgba(240, 185, 11, 0.68),
      0 0 0 4px rgba(240, 185, 11, 0.12);
  }

  :deep(.el-form-item.is-error .el-input__wrapper) {
    box-shadow:
      inset 0 0 0 1px rgba(246, 70, 93, 0.7),
      0 0 0 4px rgba(246, 70, 93, 0.12);
  }
}

.captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 118px;
  gap: 10px;
}

.captcha-panel {
  height: 46px;
  border: 1px solid var(--line);
  border-radius: 14px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.82);
  cursor: pointer;
  transition: transform 0.2s ease, border-color 0.2s ease, background-color 0.2s ease;

  img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  &:hover {
    transform: translateY(-1px);
    border-color: rgba(240, 185, 11, 0.42);
  }
}

.captcha-panel,
.captcha-panel--fallback {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.captcha-panel--fallback {
  gap: 8px;
  padding: 0 12px;
  color: #616978;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
}

.submit-button {
  width: 100%;
  min-height: 48px;
  margin-top: 4px;
  border: 0;
  border-radius: 14px;
  font-family: var(--display);
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.02em;
  background: linear-gradient(135deg, #f8d767 0%, var(--primary) 60%, #d89e00 100%);
  color: #161a1f;
  box-shadow: 0 18px 36px rgba(240, 185, 11, 0.24);

  :deep(.el-icon) {
    margin-left: 8px;
  }
}

.auth-panel__footer {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--line);
}

.footer-note__label {
  display: inline-block;
  margin-bottom: 8px;
  color: var(--muted-soft);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: none;
}

.footer-note ul {
  list-style: none;
  display: grid;
  gap: 8px;
}

.footer-note li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.4;

  .el-icon {
    margin-top: 2px;
    color: var(--primary);
  }
}

@media (max-width: 1120px) {
  .login-grid {
    grid-template-columns: 1fr;
    gap: 20px;
    padding: 20px;
    height: auto;
  }

  .auth-panel {
    order: -1;
    justify-self: stretch;
    width: 100%;
    height: auto;
  }

  .market-stage {
    height: auto;
  }

  .signal-strip,
  .ops-board__grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .login-grid {
    padding: 14px;
    height: auto;
  }

  .market-stage,
  .auth-panel {
    height: auto;
    padding: 22px 18px;
    border-radius: 22px;
  }

  .market-stage__masthead {
    flex-direction: column;
    align-items: flex-start;
  }

  .market-stage__intro {
    margin: 30px 0 24px;
  }

  .market-stage__intro h1 {
    font-size: 38px;
  }

  .captcha-row {
    grid-template-columns: 1fr;
  }

  .captcha-panel,
  .captcha-panel--fallback {
    height: 56px;
  }
}

@media (max-height: 920px) and (min-width: 1121px) {
  .login-grid {
    gap: 18px;
    padding: 16px;
  }

  .market-stage,
  .auth-panel {
    height: calc(100vh - 32px);
  }

  .market-stage {
    padding: 20px 22px;
    gap: 12px;
  }

  .market-stage__intro h1 {
    font-size: clamp(30px, 3.5vw, 52px);
  }

  .market-stage__lead {
    font-size: 13px;
    line-height: 1.45;
  }

  .signal-card {
    min-height: 84px;
  }
}

@media (max-height: 820px) and (min-width: 1121px) {
  .market-stage__lead {
    display: none;
  }

  .signal-card small,
  .rail-card p,
  .auth-panel__header p {
    display: none;
  }

  .auth-panel__header {
    margin-bottom: 10px;
  }
}

@keyframes stageReveal {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes panelFloat {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.985);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes riseIn {
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes tickerPulse {
  0%,
  100% {
    background: transparent;
  }
  50% {
    background: rgba(240, 185, 11, 0.06);
  }
}

@keyframes beamFloat {
  0%,
  100% {
    transform: translate3d(0, 0, 0) scale(1);
  }
  50% {
    transform: translate3d(18px, -14px, 0) scale(1.08);
  }
}
</style>
