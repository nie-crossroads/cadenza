import { reactive } from 'vue'
import type { Track, AddState } from '../api/types'
import { addTrack as apiAdd, ApiError, probePlayable } from '../api'
import { COPY, DUPLICATE_CODE, mapAddError, qualityChoices, qualityLabel } from '../copy/messages'
import {
  beginPlayableCheck,
  cancelPlayableCheck,
  chooseQuality,
  confirmAdd,
  endPlayableCheck,
  playableCheckCurrent,
  settleNotice,
  showUnplayable,
} from './useConfirm'
import { handleApiFailure, useSession } from './useSession'
import { pushToast } from './useToast'

/**
 * 逐行独立的添加五态，模块级内存单例。
 * key = sourceId，因此切换试听 / 重新搜索都不会重置任何行的添加态。
 */
export const states = reactive<Record<string, AddState>>({})
/** 全局序号 + 每行序号，双重防止过期响应回填 */
let globalSeq = 0
const rowSeq: Record<string, number> = {}

export function getAddState(sourceId: string): AddState {
  if (!states[sourceId]) {
    states[sourceId] = { phase: 'idle', reason: '' }
  }
  return states[sourceId]
}

/** 点击添加：先选音质，再确认这一档能播，然后确认，最后才下载上传。 */
export async function requestAdd(track: Track): Promise<void> {
  if (useSession().state.status !== 'authenticated') return
  const s = getAddState(track.sourceId)
  if (s.phase === 'pending' || s.phase === 'success' || s.phase === 'duplicate') return
  cancelPlayableCheck()
  settleNotice()
  const picked = await chooseQuality(track.name, qualityChoices(track.quality))
  if (!picked) return
  const allowed = await ensurePlayable(track, picked)
  if (!allowed) return
  const ok = await confirmAdd(track.name, qualityLabel(picked))
  if (!ok) return
  await add(track, picked)
}

async function ensurePlayable(track: Track, quality: 'std' | 'hi' | 'hr'): Promise<boolean> {
  const ticket = beginPlayableCheck()
  try {
    const playable = await probePlayable(track.sourceId, quality, ticket.signal)
    if (!playableCheckCurrent(ticket.token)) return false
    endPlayableCheck(ticket.token)
    if (playable) return true
    await showUnplayable()
    return false
  } catch (err) {
    if (!playableCheckCurrent(ticket.token)) return false
    endPlayableCheck(ticket.token)
    const code = err instanceof ApiError ? err.code : 'NETWORK_ERROR'
    if (handleApiFailure(code)) return false
    await showUnplayable()
    return false
  }
}

export async function add(track: Track, quality: 'std' | 'hi' | 'hr'): Promise<void> {
  if (useSession().state.status !== 'authenticated') return
  const s = getAddState(track.sourceId)
  // 逻辑去重：pending / success / duplicate 都不允许再次发起
  if (s.phase !== 'idle' && s.phase !== 'error') return

  s.phase = 'pending'
  s.reason = ''
  const seq = ++globalSeq
  rowSeq[track.sourceId] = seq

  try {
    // 添加传 sourceId 和所选音质；name / author / cover 由 Go 复核。
    if (track.name.trim().length > 20 || track.author.trim().length > 20) {
      s.phase = 'error'
      s.reason = mapAddError('BLOG_VALIDATION_FAILED')
      pushToast(s.reason, 'error')
      return
    }
    await apiAdd(track.sourceId, quality)
    if (rowSeq[track.sourceId] !== seq) return
    s.phase = 'success'
    const who = track.author ? `${track.name} · ${track.author}` : track.name
    pushToast(COPY.toastAdded, 'success', who)
  } catch (err) {
    if (rowSeq[track.sourceId] !== seq) return
    const code = err instanceof ApiError ? err.code : 'NETWORK_ERROR'

    if (handleApiFailure(code)) {
      // ★ pending 期间会话失效：该行强制回落 idle，不允许永久卡在「添加中…」
      s.phase = 'idle'
      s.reason = ''
      return
    }
    if (code === DUPLICATE_CODE) {
      // 409 幂等：不用红不用绿、不弹 Toast；补充中性解释，终态保持。
      s.phase = 'duplicate'
      s.reason = COPY.addDuplicateNote
      return
    }
    s.phase = 'error'
    s.reason = mapAddError(code)
    // F4：失败额外给一条 error tone Toast（6s / assertive），行内原因仍保留。
    // 409 duplicate 在此之上已 return，仍不弹 Toast。
    pushToast(COPY.toastAddFailed, 'error', s.reason)
  }
}

export function retry(track: Track): void {
  if (getAddState(track.sourceId).phase !== 'error') return
  void requestAdd(track)
}

/** 会话失效清态：删除所有行状态，并让全部在途响应失效 */
export function resetAddStates(): void {
  globalSeq += 1
  for (const key of Object.keys(states)) delete states[key]
  for (const key of Object.keys(rowSeq)) delete rowSeq[key]
}

export function useAdd() {
  return { states, getAddState, requestAdd, add, retry, resetAddStates }
}
