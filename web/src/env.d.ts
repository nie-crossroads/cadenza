/// <reference types="vite/client" />

// VITE_API_MODE=mock 时启用前端内置的离线演示适配器（无 Go 后端时可直接跑 dev）。
// 该开关只影响 api 层，不持久化任何业务状态。
interface ImportMetaEnv {
  readonly VITE_API_MODE?: 'mock' | 'proxy'
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
