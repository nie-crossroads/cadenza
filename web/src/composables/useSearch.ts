import { reactive } from 'vue'
import type { Track, SearchPhase } from '../api/types'
import { search as apiSearch, ApiError, isAbort } from '../api'
import { COPY, mapError } from '../copy/messages'
import { handleApiFailure } from './useSession'

interface SearchState {
  query: string
  /** empty / typing / loading / results / noResults / error */
  phase: SearchPhase
  items: Track[]
  /** 已提交的关键词，用于无结果文案 */
  keyword: string
  errorCode: string
  errorHint: string
  /** role=status 播报内容 */
  announcement: string
  /** 本次会话内最近 3 次搜索，只在内存里 */
  recent: string[]
}

const state = reactive<SearchState>({
  query: '',
  phase: 'empty',
  items: [],
  keyword: '',
  errorCode: '',
  errorHint: '',
  announcement: '',
  recent: [],
})

let debounceTimer: number | null = null
let controller: AbortController | null = null
/** 只认最后一次请求，杜绝竞态覆盖 */
let requestSeq = 0
let flip = false

function cancelDebounce(): void {
  if (debounceTimer !== null) {
    window.clearTimeout(debounceTimer)
    debounceTimer = null
  }
}

function abortActive(): void {
  if (controller) {
    controller.abort()
    controller = null
  }
}

/** 交替附加零宽字符，保证连续两次相同播报也能被读屏重新播报 */
function announce(text: string): void {
  flip = !flip
  state.announcement = flip ? text : `${text}\u200B`
}

async function runSearch(): Promise<void> {
  cancelDebounce()
  const q = state.query.trim()
  if (q.length < 1) return

  const seq = ++requestSeq
  abortActive()
  const ctrl = new AbortController()
  controller = ctrl

  state.keyword = q
  state.recent = [q, ...state.recent.filter((item) => item !== q)].slice(0, 3)
  state.phase = 'loading'
  state.errorCode = ''
  state.errorHint = ''
  announce(COPY.searching)

  try {
    const items = await apiSearch(q, ctrl.signal)
    if (seq !== requestSeq) return
    state.items = items
    state.phase = items.length > 0 ? 'results' : 'noResults'
    announce(items.length > 0 ? COPY.found(items.length) : COPY.noResultsSub)
  } catch (err) {
    if (isAbort(err) || seq !== requestSeq) return
    const code = err instanceof ApiError ? err.code : 'NETWORK_ERROR'
    // 会话类 code 走全局清态；其余一律按搜索失败处理。
    if (handleApiFailure(code)) return
    state.errorCode = code
    state.errorHint = mapError(code, COPY.searchFailedSub)
    state.phase = 'error'
    announce(COPY.searchFailedTitle)
  } finally {
    if (seq === requestSeq) controller = null
  }
}

/** 输入变化：300ms 防抖，trim 后至少 1 字才提交 */
export function setQuery(value: string): void {
  state.query = value
  cancelDebounce()
  if (value.trim().length < 1) {
    abortActive()
    requestSeq += 1
    state.phase = 'empty'
    state.items = []
    state.keyword = ''
    state.announcement = ''
    return
  }
  debounceTimer = window.setTimeout(() => { void runSearch() }, 300)
}

/** Enter：取消防抖立即提交 */
export function submitNow(): void {
  void runSearch()
}

/** Esc / ×：清空 + abort + 回初始空态；★ 不停止播放、不登出 */
export function clearSearch(): void {
  cancelDebounce()
  abortActive()
  requestSeq += 1
  state.query = ''
  state.phase = 'empty'
  state.items = []
  state.keyword = ''
  state.errorCode = ''
  state.errorHint = ''
  state.announcement = ''
}

export function retrySearch(): void {
  void runSearch()
}

/** 会话失效时清空全部搜索态 */
export function resetSearch(): void {
  clearSearch()
}

export function setAnnouncement(text: string): void {
  announce(text)
}

export function useSearch() {
  return {
    state,
    setQuery,
    submitNow,
    clearSearch,
    retrySearch,
    resetSearch,
    setAnnouncement,
  }
}
