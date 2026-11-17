/**
 * 数据模型。
 * ★ 搜索结果 Track：name / author / cover / duration / source / sourceId / quality。
 *   quality 只接受 hi、std、hr，页面再映射成音质标签。
 *   绝不含源站播放地址、临时链接、COS 地址或任何凭证。
 * ★ sourceId 只在内存对象与 DOM 的 data-source-id 上出现；
 *   不渲染为文本、不进 aria-label / title / 页面 URL / 日志。
 */

/** hi 高品 / std 标准 / hr Hi-Res。未知值在 pickTrack 里落到 std。 */
export type TrackQuality = 'hi' | 'std' | 'hr'

export interface Track {
  name: string
  author: string
  cover: string
  /** 秒 */
  duration: number
  source: string
  sourceId: string
  quality: TrackQuality
}

export type SearchPhase = 'empty' | 'typing' | 'loading' | 'results' | 'noResults' | 'error'

export type AddPhase = 'idle' | 'pending' | 'success' | 'duplicate' | 'error'

export interface AddState {
  phase: AddPhase
  /** 已映射过的中文原因，只用于 error 态行内显示 */
  reason: string
}

/** empty / buffering / playing / paused / seek / volume / ended / error
 *  seek / volume 为瞬时交互态，存在 state.interaction，不覆盖播放主状态。 */
export type PlayerStatus =
  | 'empty'
  | 'buffering'
  | 'playing'
  | 'paused'
  | 'ended'
  | 'error'
export type PlayerInteraction = 'seek' | 'volume' | null

export type WaveStatus = 'idle' | 'buffering' | 'playing' | 'paused' | 'error'

export type SessionStatus = 'unknown' | 'checking' | 'authenticated' | 'anonymous'

/** 六种字段的防御性裁剪：后端多给任何字段都在这里丢弃 */
export function pickTrack(raw: unknown): Track | null {
  if (typeof raw !== 'object' || raw === null) return null
  const o = raw as Record<string, unknown>
  if (typeof o.name !== 'string' || typeof o.sourceId !== 'string') return null
  const quality: TrackQuality = o.quality === 'hi' || o.quality === 'hr' ? o.quality : 'std'
  return {
    name: o.name,
    author: typeof o.author === 'string' ? o.author : '',
    cover: typeof o.cover === 'string' ? o.cover : '',
    duration: typeof o.duration === 'number' && Number.isFinite(o.duration) ? o.duration : 0,
    source: typeof o.source === 'string' ? o.source : '',
    sourceId: o.sourceId,
    quality,
  }
}
