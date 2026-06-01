import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import { VueQueryPlugin } from '@tanstack/vue-query'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import App from './App.vue'
import { router } from './router'
import { setupPermissionDirective } from './directives/permission'
import { useAuthStore } from './stores/auth'
import { configureAdminIdGetter, installClogLifecycle, report } from './utils/clog'
import './styles/index.scss'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.use(VueQueryPlugin)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

setupPermissionDirective(app)

// ---- 全局错误上报 ----
const authStore = useAuthStore()
configureAdminIdGetter(() => authStore.user?.id)

app.config.errorHandler = (err, instance, info) => {
  const e = err instanceof Error ? err : new Error(String(err))
  report('error', e.message, {
    stack: e.stack,
    extra: {
      info,
      component: (instance?.$options as { name?: string } | undefined)?.name,
    },
  })
  // eslint-disable-next-line no-console
  console.error('[vue:errorHandler]', err, info)
}

window.addEventListener('error', (ev) => {
  const e = ev.error instanceof Error ? ev.error : null
  report('error', e?.message || ev.message || 'window error', {
    stack: e?.stack,
    extra: {
      filename: ev.filename,
      lineno: ev.lineno,
      colno: ev.colno,
    },
  })
})

window.addEventListener('unhandledrejection', (ev) => {
  const reason: unknown = ev.reason
  const e = reason instanceof Error ? reason : null
  report('error', e?.message || String(reason ?? 'unhandledrejection'), {
    stack: e?.stack,
    extra: { kind: 'unhandledrejection' },
  })
})

installClogLifecycle()

// 恢复本地 token 后再挂载
authStore.init().then(() => {
  app.mount('#app')
})
