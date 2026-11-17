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
  /** 该来源是否还有未取回的结果 */
  more: Record<string, boolean>
  loadingMore: boolean
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
  more: {},
  loadingMore: false,
})

let controller: AbortController | null = null
let pageController: AbortController | null = null
/** 只认最后一次请求，杜绝竞态覆盖 */
let requestSeq = 0
let flip = false

function abortActive(): void {
  if (controller) {
    controller.abort()
    controller = null
  }
  if (pageController) {
    pageController.abort()
    pageController = null
  }
}

/** 交替附加零宽字符，保证连续两次相同播报也能被读屏重新播报 */
function announce(text: string): void {
  flip = !flip
  state.announcement = flip ? text : `${text}\u200B`
}

async function runSearch(): Promise<void> {
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
    const items = await apiSearch(q, ctrl.signal, { limit: 5 })
    if (seq !== requestSeq) return
    state.items = items
    state.more = moreFlags(items, 5)
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

/** 只改搜索框里的字，不发请求。 */
export function setQuery(value: string): void {
  state.query = value
}

/** 点击搜索或按 Enter 后才提交。 */
export function submitNow(): void {
  void runSearch()
}

/** Esc / ×：清空 + abort + 回初始空态；★ 不停止播放、不登出 */
export function clearSearch(): void {
  abortActive()
  requestSeq += 1
  state.query = ''
  state.phase = 'empty'
  state.items = []
  state.keyword = ''
  state.errorCode = ''
  state.errorHint = ''
  state.announcement = ''
  state.more = {}
  state.loadingMore = false
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

function moreFlags(items: Track[], pageSize: number): Record<string, boolean> {
  const counts: Record<string, number> = {}
  for (const item of items) counts[item.source] = (counts[item.source] ?? 0) + 1
  const more: Record<string, boolean> = {}
  for (const [name, count] of Object.entries(counts)) more[name] = count >= pageSize
  return more
}

function orderedItems(name: string, next: Track[], mode: 'replace' | 'append'): Track[] {
  const order: string[] = []
  for (const item of state.items) {
    if (!order.includes(item.source)) order.push(item.source)
  }
  if (!order.includes(name)) order.push(name)
  const buckets = new Map<string, Track[]>()
  for (const source of order) buckets.set(source, [])
  for (const item of state.items) {
    if (mode === 'replace' && item.source === name) continue
    buckets.get(item.source)?.push(item)
  }
  const current = buckets.get(name) ?? []
  buckets.set(name, mode === 'append' ? [...current, ...next] : next)
  return order.flatMap((source) => buckets.get(source) ?? [])
}

async function fetchPage(name: string, offset: number, limit: number): Promise<Track[] | null> {
  const seq = requestSeq
  const q = state.keyword.trim()
  if (!q) return null
  pageController?.abort()
  const ctrl = new AbortController()
  pageController = ctrl
  state.loadingMore = true
  try {
    const items = await apiSearch(q, ctrl.signal, { source: name, offset, limit })
    if (seq !== requestSeq) return null
    return items
  } catch (err) {
    if (isAbort(err) || seq !== requestSeq) return null
    return []
  } finally {
    if (pageController === ctrl) pageController = null
    if (seq === requestSeq) state.loadingMore = false
  }
}

/** 切到单个来源时补到至少 10 首。 */
export async function ensureSource(name: string): Promise<void> {
  const have = state.items.filter((item) => item.source === name).length
  if (have >= 10 || state.more[name] === false) return
  const items = await fetchPage(name, 0, 10)
  if (!items) return
  state.items = orderedItems(name, items, 'replace')
  state.more[name] = items.length >= 10
}

/** 单个来源继续取后面的结果。 */
export async function loadSourceMore(name: string): Promise<void> {
  if (!state.more[name]) return
  const have = state.items.filter((item) => item.source === name).length
  const items = await fetchPage(name, have, 20)
  if (!items) return
  state.items = orderedItems(name, items, 'append')
  state.more[name] = items.length >= 20
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
    ensureSource,
    loadSourceMore,
  }
}
