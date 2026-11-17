import { reactive } from 'vue'
import type { Track, PlayerStatus, PlayerInteraction } from '../api/types'
import { previewUrl } from '../api'
import { playerStatusText } from '../copy/messages'
import { clamp } from '../utils/format'

/** 无 Go 后端时用可控模拟时钟代替真实音频，便于本机 dev 自测。 */
const USE_MOCK = import.meta.env.VITE_API_MODE === 'mock'

interface PlayerState {
  track: Track | null
  /** empty / buffering / playing / paused / ended / error */
  status: PlayerStatus
  /** seek / volume 为瞬时交互态，不覆盖播放主状态 */
  interaction: PlayerInteraction
  currentTime: number
  duration: number
  buffered: number
  volume: number
  muted: boolean
  /** 静音记忆 */
  lastVolume: number
  /** Go 转发接口支持 Range 时为 true；不可用时降级顺序播放，绝不回退源站直连 */
  rangeSupported: boolean
  errorCode: string
}

export const state = reactive<PlayerState>({
  track: null,
  status: 'empty',
  interaction: null,
  currentTime: 0,
  duration: 0,
  buffered: 0,
  volume: 0.8,
  muted: false,
  lastVolume: 0.8,
  rangeSupported: true,
  errorCode: '',
})

let audio: HTMLAudioElement | null = null
let switching = false
let simTimer: number | null = null
let interactionTimer: number | null = null

function flashInteraction(kind: 'seek' | 'volume'): void {
  state.interaction = kind
  if (interactionTimer !== null) window.clearTimeout(interactionTimer)
  interactionTimer = window.setTimeout(() => { state.interaction = null }, 600)
}

function resetState(): void {
  state.track = null
  state.status = 'empty'
  state.interaction = null
  state.currentTime = 0
  state.duration = 0
  state.buffered = 0
  state.volume = 0.8
  state.muted = false
  state.lastVolume = 0.8
  state.rangeSupported = true
  state.errorCode = ''
}

/* ---------------------- 真实 <audio> 分支 ---------------------- */

function detectRange(): void {
  if (!audio) return
  const seekable = audio.seekable
  const hasRange = seekable.length > 0 && seekable.end(seekable.length - 1) > 0
  const finiteDuration = Number.isFinite(audio.duration) && audio.duration > 0
  state.rangeSupported = hasRange && finiteDuration
}

function onLoadedMetadata(): void {
  if (!audio) return
  const d = audio.duration
  state.duration = Number.isFinite(d) && d > 0 ? d : (state.track?.duration ?? 0)
  detectRange()
}

function onTimeUpdate(): void {
  if (audio) state.currentTime = audio.currentTime
}

function onProgress(): void {
  if (!audio) return
  const b = audio.buffered
  state.buffered = b.length > 0 ? b.end(b.length - 1) : 0
  detectRange()
}

function onPlaying(): void { state.status = 'playing' }
function onWaiting(): void { if (!switching) state.status = 'buffering' }
function onPause(): void {
  if (!switching && state.status === 'playing') state.status = 'paused'
}
function onEnded(): void {
  state.status = 'ended'
  state.currentTime = state.duration
  // ended 不自动下一首
}
function onAudioError(): void {
  if (switching) return
  state.status = 'error'
  state.errorCode = 'PREVIEW_FAILED'
}

function ensureAudio(): HTMLAudioElement {
  if (audio) return audio
  const a = new Audio()
  a.preload = 'metadata'
  a.addEventListener('loadedmetadata', onLoadedMetadata)
  a.addEventListener('durationchange', onLoadedMetadata)
  a.addEventListener('timeupdate', onTimeUpdate)
  a.addEventListener('progress', onProgress)
  a.addEventListener('canplay', detectRange)
  a.addEventListener('playing', onPlaying)
  a.addEventListener('waiting', onWaiting)
  a.addEventListener('pause', onPause)
  a.addEventListener('ended', onEnded)
  a.addEventListener('error', onAudioError)
  audio = a
  return a
}

/* ---------------------- 模拟时钟分支 ---------------------- */

function stopSim(): void {
  if (simTimer !== null) {
    window.clearInterval(simTimer)
    simTimer = null
  }
}

function startSim(): void {
  stopSim()
  simTimer = window.setInterval(() => {
    if (state.status !== 'playing' || !state.track) return
    state.currentTime = Math.min(state.duration, state.currentTime + 0.2)
    state.buffered = Math.min(state.duration, Math.max(state.buffered, state.currentTime + 12))
    if (state.currentTime >= state.duration) onEnded()
  }, 200)
}

/* ---------------------- 对外动作 ---------------------- */

/** 切换试听：重置进度、换 playing 行；不重置任何行的添加态、不打断搜索。 */
export function playTrack(track: Track): void {
  switching = true
  stopSim()
  state.track = track
  state.currentTime = 0
  state.buffered = 0
  state.duration = track.duration
  state.errorCode = ''
  state.rangeSupported = true
  state.status = 'buffering'

  if (USE_MOCK) {
    switching = false
    window.setTimeout(() => {
      if (state.track?.sourceId === track.sourceId) {
        state.status = 'playing'
        startSim()
      }
    }, 260)
    return
  }

  const a = ensureAudio()
  a.pause()
  // ★ audio.src 只允许指向 Go 的转发接口；sourceId 只出现在这条同源内部请求里。
  a.src = previewUrl(track.sourceId)
  a.volume = state.muted ? 0 : state.volume
  a.muted = state.muted
  a.load()
  switching = false
  void a.play().catch(() => {
    // 浏览器拒绝自动播放时保持 paused，由用户点击播放钮继续。
    if (state.track?.sourceId === track.sourceId) state.status = 'paused'
  })
}

export function toggleAudition(track: Track): void {
  if (state.track?.sourceId === track.sourceId) {
    if (state.status === 'playing' || state.status === 'buffering') pausePlayback()
    else resumePlayback()
    return
  }
  playTrack(track)
}

export function togglePlay(): void {
  if (!state.track) return
  if (state.status === 'playing' || state.status === 'buffering') pausePlayback()
  else resumePlayback()
}

export function pausePlayback(): void {
  if (USE_MOCK) {
    if (state.status === 'playing' || state.status === 'buffering') state.status = 'paused'
    return
  }
  audio?.pause()
}

export function resumePlayback(): void {
  if (state.status === 'ended') {
    state.currentTime = 0
    if (audio && !USE_MOCK) audio.currentTime = 0
  }
  if (USE_MOCK) {
    state.status = 'playing'
    startSim()
    return
  }
  if (!audio) return
  void audio.play().catch(() => { state.status = 'paused' })
}

export function retryPreview(): void {
  const track = state.track
  if (!track) return
  state.track = null
  playTrack(track)
}

/** Range 不可用时 seek 为空操作：降级为顺序播放 + 中性提示，绝不直连源站。 */
export function seek(seconds: number): void {
  if (!state.track || !state.rangeSupported) return
  const t = clamp(seconds, 0, state.duration || 0)
  state.currentTime = t
  if (!USE_MOCK && audio) {
    try { audio.currentTime = t } catch { /* 非 seekable 流忽略 */ }
  }
  flashInteraction('seek')
}

export function seekBy(delta: number): void {
  seek(state.currentTime + delta)
}

export function setVolume(ratio: number): void {
  const v = clamp(ratio, 0, 1)
  state.volume = v
  if (v > 0) state.lastVolume = v
  state.muted = v === 0
  if (audio) {
    audio.volume = v
    audio.muted = false
  }
  flashInteraction('volume')
}

export function toggleMute(): void {
  if (state.muted) {
    const v = state.lastVolume > 0 ? state.lastVolume : 0.8
    state.muted = false
    state.volume = v
    if (audio) { audio.muted = false; audio.volume = v }
  } else {
    state.lastVolume = state.volume
    state.muted = true
    if (audio) audio.muted = true
  }
  flashInteraction('volume')
}

/** 会话失效 / 退出时调用：停播 + 清空全部播放态。 */
export function stopPlayback(): void {
  stopSim()
  if (!USE_MOCK && audio) {
    audio.pause()
    audio.removeAttribute('src')
    audio.load()
  }
  resetState()
}

export function statusText(): string {
  return playerStatusText(state.status)
}

export function usePlayer() {
  return {
    state,
    playTrack,
    toggleAudition,
    togglePlay,
    pausePlayback,
    resumePlayback,
    retryPreview,
    seek,
    seekBy,
    setVolume,
    toggleMute,
    stopPlayback,
    statusText,
  }
}
