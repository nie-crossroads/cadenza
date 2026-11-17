import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// 开发期：Vite 把 /api 全部代理给本机 Go 进程。
// 生产期：Vue 产物由 Go 用 go:embed 同源提供，不再需要本代理。
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  preview: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    target: 'es2020',
    outDir: 'dist',
    // 首屏不引入任何第三方运行时资源；拆包只做路由级懒加载。
    rollupOptions: {
      output: {
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router'],
        },
      },
    },
  },
})
