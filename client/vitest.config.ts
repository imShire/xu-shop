import path from 'node:path'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  define: {
    'process.env.TARO_ENV': JSON.stringify(process.env.TARO_ENV || 'h5'),
    'process.env.TARO_APP_API_BASE': JSON.stringify(''),
    'process.env.TARO_APP_VERSION': JSON.stringify('test-1.0.0'),
  },
  test: {
    environment: 'jsdom',
    globals: false,
    setupFiles: ['./tests/setup.ts'],
    include: ['tests/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
    },
  },
})
