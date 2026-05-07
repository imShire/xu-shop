import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';
import path from 'node:path';
// 阶段 0 由 Codex 完善：可能需要按需国际化、CDN、productionSourceMap 配置
export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');
    return {
        plugins: [
            vue(),
            AutoImport({ resolvers: [ElementPlusResolver()] }),
            Components({ resolvers: [ElementPlusResolver()] })
        ],
        resolve: { alias: { '@': path.resolve(__dirname, 'src') } },
        server: {
            port: 5173,
            proxy: {
                '/api': { target: env.VITE_API_BASE?.replace(/\/api\/v1$/, '') ?? 'http://localhost:8080', changeOrigin: true }
            }
        }
    };
});
