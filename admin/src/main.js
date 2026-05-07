import { createApp } from 'vue';
import { createPinia } from 'pinia';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import { VueQueryPlugin } from '@tanstack/vue-query';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import App from './App.vue';
import { router } from './router';
import { setupPermissionDirective } from './directives/permission';
import { useAuthStore } from './stores/auth';
import './styles/index.scss';
const app = createApp(App);
const pinia = createPinia();
app.use(pinia);
app.use(router);
app.use(ElementPlus, { locale: zhCn });
app.use(VueQueryPlugin);
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component);
}
setupPermissionDirective(app);
// 恢复本地 token 后再挂载
const authStore = useAuthStore();
authStore.init().then(() => {
    app.mount('#app');
});
