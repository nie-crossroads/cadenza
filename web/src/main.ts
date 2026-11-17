import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/tokens.css'
import './styles/base.css'
import './composables/useTheme'

const app = createApp(App)
app.use(router)

// 首屏会话探测由 router.beforeEach 在首次导航内 await 完成（probePromise 全局缓存一次），
// router.isReady() 在首次导航（含异步守卫）结束后才 resolve，因此挂载时会话态已确定。
void router.isReady().then(() => app.mount('#app'))
