/**
 * ============================================================================
 * 接口契约【草案 · 需与 Go 侧冻结】  TODO(go): 冻结后移除 mock 分支与 TODO 标记
 * ----------------------------------------------------------------------------
 * 认证：网页登录用「口令 + HttpOnly Cookie」。
 *       前端只发 credentials:'same-origin'，★ 不设置任何认证请求头，
 *       也绝不用 JS 读写 Cookie；页面 / DOM / 日志里不出现任何凭证。
 *
 * GET  /api/session
 *      首屏探测。200 → 已登录；401/419 → 匿名。结果在 useSession 内缓存，不重复请求。
 * POST /api/login      { password }                → 200 + Set-Cookie(HttpOnly)
 * GET  /api/search?q=  → { items: Track[] }
 *      Track：{ name, author, cover, duration(秒), source, sourceId, quality }
 *      ★ 绝不含源站播放地址 / 临时链接 / 存储地址。
 * GET  /api/preview?sourceId=
 *      → 206 Partial Content + Accept-Ranges（由 Go 转发音频并支持 Range）
 *      ★ audio.src 只允许指向这里；浏览器不直连音乐源站。
 * POST /api/music      { sourceId }               → 200 { id } / 409 ADD_DUPLICATE
 *      ★ 添加只传 sourceId；name / author / cover 由 Go 复核，前端不重复提交。
 *      失败体 { code }，code 必须落在 copy/messages.ts 的白名单里。
 *
 * 开发期：Vite 代理 /api → http://localhost:8080（见 vite.config.ts）。
 * 生产期：Vue 产物由 Go go:embed 同源提供，页面与接口同源，无跨域、无第二进程。
 * 离线演示：VITE_API_MODE=mock 时启用 api/mock.ts，便于无 Go 环境自测。
 * ============================================================================
 */
import type { Track } from './types'
import { pickTrack } from './types'
import { isSessionInvalid, SESSION_INVALID_CODES } from '../copy/messages'
import { MOCK_TRACKS, mockProbe, mockLogin, mockLogout, mockSearch, mockAdd } from './mock'

/** 错误只携带白名单码；message 固定为安全字符串，绝不携带后端原文或栈。 */
export class ApiError extends Error {
  readonly code: string
  constructor(code: string) {
    super('api error')
    this.name = 'ApiError'
    this.code = code
  }
}

const USE_MOCK = import.meta.env.VITE_API_MODE === 'mock'

/** 添加类白名单码；任何不在表内的 code 一律降级为 ADD_UNKNOWN */
const ADD_WHITELIST: ReadonlySet<string> = new Set([
  'ADD_DOWNLOAD_FAILED',
  'ADD_COS_UPLOAD_FAILED',
  'ADD_FILE_TOO_LARGE',
  'ADD_TIMEOUT',
  'ADD_BLOG_TOKEN_INVALID',
  'NETWORK_ERROR',
  'BLOG_VALIDATION_FAILED',
  'ADD_UNKNOWN',
])

/**
 * ★ 401 分流（最高风险点）
 * - 只有会话类 401/419（SESSION_INVALID / UNAUTHENTICATED）才允许清态 / 停播 / 跳登录。
 * - 博客侧 401/419 一律包装成 ADD_BLOG_TOKEN_INVALID，只算「这一行」添加失败，
 *   ★ 绝不按裸 HTTP 401 触发登出，也绝不把后端 message 原文透出。
 * - 409 视为幂等提示（duplicate），不是错误。
 */
export function normalizeAddCode(status: number, raw: string): string {
  if (isSessionInvalid(raw)) return raw
  if (status === 401 || status === 419) return 'ADD_BLOG_TOKEN_INVALID'
  if (status === 409) return 'ADD_DUPLICATE'
  if (ADD_WHITELIST.has(raw)) return raw
  return 'ADD_UNKNOWN'
}

/** 只读 code 字段；message / stack / 其它字段一律丢弃 */
async function readCode(res: Response): Promise<string> {
  try {
    const body: unknown = await res.json()
    if (typeof body === 'object' && body !== null) {
      const code = (body as Record<string, unknown>).code
      if (typeof code === 'string') return code
    }
  } catch {
    /* 非 JSON 响应按空码处理 */
  }
  return ''
}

function isAbort(err: unknown): boolean {
  return err instanceof DOMException && err.name === 'AbortError'
}

export { isAbort, SESSION_INVALID_CODES }

/* ------------------------------- 会话 ------------------------------- */

export async function probeSession(): Promise<boolean> {
  if (USE_MOCK) return mockProbe()
  try {
    const res = await fetch('/api/session', {
      credentials: 'same-origin',
      cache: 'no-store',
    })
    return res.ok
  } catch {
    return false
  }
}

export async function login(password: string): Promise<void> {
  if (USE_MOCK) {
    await mockLogin(password)
    return
  }
  let res: Response
  try {
    res = await fetch('/api/login', {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password }),
    })
  } catch {
    throw new ApiError('LOGIN_UNAVAILABLE')
  }
  if (res.ok) return
  const raw = await readCode(res)
  if (isSessionInvalid(raw)) throw new ApiError(raw)
  if (res.status === 401 || res.status === 403) throw new ApiError('LOGIN_INVALID')
  throw new ApiError('LOGIN_UNAVAILABLE')
}

export async function logout(): Promise<void> {
  if (USE_MOCK) {
    mockLogout()
    return
  }
  try {
    await fetch('/api/logout', { method: 'POST', credentials: 'same-origin' })
  } catch {
    /* 退出以清掉本机会话为准，网络失败也离开主页 */
  }
}

/* ------------------------------- 搜索 ------------------------------- */

export async function search(q: string, signal: AbortSignal): Promise<Track[]> {
  if (USE_MOCK) return mockSearch(q, signal)
  let res: Response
  try {
    res = await fetch(`/api/search?q=${encodeURIComponent(q)}`, {
      credentials: 'same-origin',
      signal,
    })
  } catch (err) {
    if (isAbort(err)) throw err
    throw new ApiError('NETWORK_ERROR')
  }
  if (!res.ok) {
    const raw = await readCode(res)
    if (isSessionInvalid(raw)) throw new ApiError(raw)
    throw new ApiError('SEARCH_FAILED')
  }
  let body: unknown
  try {
    body = await res.json()
  } catch {
    throw new ApiError('SEARCH_FAILED')
  }
  const items = (body as { items?: unknown })?.items
  if (!Array.isArray(items)) return []
  // 防御性裁剪：播放地址等多余字段丢弃，音质只保留 hi / std / hr。
  return items.map(pickTrack).filter((t): t is Track => t !== null)
}

/* ------------------------------- 试听 ------------------------------- */

/**
 * audio.src 唯一允许的地址。sourceId 只出现在这条同源内部请求里：
 * 不渲染为文本、不进 aria-label / title / 页面 URL / 日志。
 * Range 不可用由 usePlayer 检测 seekable 后降级，绝不回退到源站直连。
 */
export function previewUrl(sourceId: string): string {
  return `/api/preview?sourceId=${encodeURIComponent(sourceId)}`
}

/* ------------------------------- 添加 ------------------------------- */

/** 只传 sourceId；返回成功或抛 ApiError(白名单码 / ADD_DUPLICATE) */
export async function addTrack(sourceId: string): Promise<void> {
  if (USE_MOCK) {
    const outcome = await mockAdd(sourceId)
    if (outcome !== 'OK') throw new ApiError(outcome)
    return
  }
  let res: Response
  try {
    res = await fetch('/api/music', {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sourceId }),
    })
  } catch {
    throw new ApiError('NETWORK_ERROR')
  }
  if (res.ok) return
  const raw = await readCode(res)
  throw new ApiError(normalizeAddCode(res.status, raw))
}

/** 演示适配器使用的虚构曲目，仅 dev 用；生产构建不会渲染其中任何链接。 */
export const DEMO_TRACKS: readonly Track[] = USE_MOCK ? MOCK_TRACKS : []
