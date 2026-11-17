import { reactive } from 'vue'
import type { TrackQuality } from '../api/types'

interface ConfirmState {
  open: boolean
  name: string
  tone: string
  resolve: ((ok: boolean) => void) | null
}

interface QualityState {
  open: boolean
  name: string
  options: TrackQuality[]
  selected: TrackQuality
  resolve: ((quality: TrackQuality | null) => void) | null
}

interface CheckState {
  open: boolean
  token: number
  controller: AbortController | null
}

interface NoticeState {
  open: boolean
  resolve: (() => void) | null
}

const state = reactive<ConfirmState>({
  open: false,
  name: '',
  tone: '',
  resolve: null,
})

const quality = reactive<QualityState>({
  open: false,
  name: '',
  options: ['std'],
  selected: 'std',
  resolve: null,
})

const checking = reactive<CheckState>({
  open: false,
  token: 0,
  controller: null,
})

const notice = reactive<NoticeState>({
  open: false,
  resolve: null,
})

export function useConfirm() {
  return { state, quality, checking, notice }
}

/** 添加前选择音质。取消或被新的选择替换时返回 null。 */
export function chooseQuality(name: string, options: TrackQuality[]): Promise<TrackQuality | null> {
  if (quality.resolve) quality.resolve(null)
  const choices = options.length > 0 ? options : (['std'] as TrackQuality[])
  return new Promise((resolve) => {
    quality.name = name
    quality.options = choices
    quality.selected = choices.includes('std') ? 'std' : choices[0]
    quality.open = true
    quality.resolve = resolve
  })
}

export function setQualityChoice(code: TrackQuality): void {
  quality.selected = code
}

export function settleQuality(ok: boolean): void {
  const resolve = quality.resolve
  const selected = quality.selected
  quality.open = false
  quality.name = ''
  quality.resolve = null
  resolve?.(ok ? selected : null)
}

/** 开始检查所选音质。新的一次检查会取消上一次。 */
export function beginPlayableCheck(): { signal: AbortSignal, token: number } {
  checking.controller?.abort()
  const controller = new AbortController()
  checking.token += 1
  checking.controller = controller
  checking.open = true
  return { signal: controller.signal, token: checking.token }
}

export function cancelPlayableCheck(): void {
  checking.controller?.abort()
  checking.open = false
  checking.controller = null
}

export function endPlayableCheck(token: number): void {
  if (checking.token !== token) return
  checking.open = false
  checking.controller = null
}

export function playableCheckCurrent(token: number): boolean {
  const controller = checking.controller
  return checking.token === token && controller != null && !controller.signal.aborted
}

/** 所选音质不能播放。关闭后才继续后面的操作。 */
export function showUnplayable(): Promise<void> {
  if (notice.resolve) notice.resolve()
  return new Promise((resolve) => {
    notice.open = true
    notice.resolve = resolve
  })
}

export function settleNotice(): void {
  const resolve = notice.resolve
  notice.open = false
  notice.resolve = null
  resolve?.()
}

/** 添加前的确认。取消或被新的确认替换时返回 false。 */
export function confirmAdd(name: string, tone: string): Promise<boolean> {
  if (state.resolve) state.resolve(false)
  return new Promise((resolve) => {
    state.name = name
    state.tone = tone
    state.open = true
    state.resolve = resolve
  })
}

export function settleConfirm(ok: boolean): void {
  const resolve = state.resolve
  state.open = false
  state.name = ''
  state.tone = ''
  state.resolve = null
  resolve?.(ok)
}
