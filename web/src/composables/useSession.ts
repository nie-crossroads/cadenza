import { reactive } from 'vue'
import type { SessionStatus } from '../api/types'
import { probeSession as apiProbeSession } from '../api'
import { isSessionInvalid } from '../copy/messages'

/** 模块级 reactive 单例：全部内存态，刷新即丢；禁止 localStorage / sessionStorage / IndexedDB。 */
interface SessionState {
  status: SessionStatus
  checked: boolean
}

const state = reactive<SessionState>({ status: 'unknown', checked: false })

/** 首屏探测结果缓存：整个生命周期只发一次 /api/session */
let probePromise: Promise<void> | null = null

/** 会话失效时需要执行的清理动作，由 App.vue 统一注册（避免与 router 形成循环依赖） */
const expiredHooks = new Set<() => void>()

export function onSessionExpired(hook: () => void): () => void {
  expiredHooks.add(hook)
  return () => { expiredHooks.delete(hook) }
}

export function probeSession(): Promise<void> {
  if (probePromise) return probePromise
  state.status = 'checking'
  probePromise = (async () => {
    try {
      const ok = await apiProbeSession()
      state.status = ok ? 'authenticated' : 'anonymous'
    } catch {
      // 探测异常按匿名处理，绝不把 guard 卡在 checking
      state.status = 'anonymous'
    } finally {
      state.checked = true
    }
  })()
  return probePromise
}

export function markAuthenticated(): void {
  state.status = 'authenticated'
  state.checked = true
}

export function markAnonymous(): void {
  state.status = 'anonymous'
  state.checked = true
}

/**
 * ★ 401 分流的唯一入口。
 * 只有 SESSION_INVALID / UNAUTHENTICATED 返回 true 并触发清态；
 * 绝不接受裸 HTTP status 作为判断依据。
 */
export function handleApiFailure(code: string): boolean {
  if (!isSessionInvalid(code)) return false
  invalidateSession()
  return true
}

/** 清态 + 停播 + 跳 /login?reason=expired（停播与跳转由 App.vue 注册的钩子完成） */
export function invalidateSession(): void {
  // 即使当前已是匿名态也照常执行钩子：保证 pending 行强制回落、Toast 清空等副作用不丢。
  markAnonymous()
  for (const hook of Array.from(expiredHooks)) hook()
}

export function useSession() {
  return {
    state,
    probeSession,
    markAuthenticated,
    markAnonymous,
    handleApiFailure,
    invalidateSession,
    onSessionExpired,
  }
}
