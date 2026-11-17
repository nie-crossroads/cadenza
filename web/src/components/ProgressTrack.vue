<script setup lang="ts">
import { computed, ref } from 'vue'
import { A11Y } from '../copy/messages'
import { clamp, formatTime } from '../utils/format'

/** 原生进度轨道：点击 seek + pointer 拖动 + 完整键盘支持。
 *  ★ 已播放填充段只用 --signal 纯色，且宽度绝对不加 transition（必须瞬时跟随）。 */
const props = withDefaults(defineProps<{
  current: number
  duration: number
  buffered: number
  /** Range 不可用：顺序播放降级，禁止拖动 */
  disabled?: boolean
}>(), { disabled: false })

const emit = defineEmits<{
  seek: [seconds: number]
  seekStart: []
  seekEnd: []
}>()

const hitRef = ref<HTMLElement | null>(null)
const dragging = ref(false)
const hoverRatio = ref(0)

const denom = computed(() => (props.duration > 0 ? props.duration : 0))
const fillPct = computed(() => `${denom.value > 0 ? clamp(props.current / denom.value, 0, 1) * 100 : 0}%`)
const bufferedPct = computed(() => `${denom.value > 0 ? clamp(props.buffered / denom.value, 0, 1) * 100 : 0}%`)
const valueText = computed(() => A11Y.timecode(`${formatTime(props.current)} / ${formatTime(props.duration)}`))
const tipText = computed(() => formatTime(hoverRatio.value * props.duration))
const tipLeft = computed(() => `${clamp(hoverRatio.value, 0, 1) * 100}%`)

function ratioFromEvent(e: PointerEvent): number {
  const el = hitRef.value
  if (!el) return 0
  const rect = el.getBoundingClientRect()
  if (rect.width <= 0) return 0
  return clamp((e.clientX - rect.left) / rect.width, 0, 1)
}

function onPointerDown(e: PointerEvent): void {
  if (props.disabled) return
  dragging.value = true
  hoverRatio.value = ratioFromEvent(e)
  hitRef.value?.setPointerCapture(e.pointerId)
  emit('seekStart')
  emit('seek', hoverRatio.value * props.duration)
  e.preventDefault()
}

function onPointerMove(e: PointerEvent): void {
  if (!dragging.value || props.disabled) return
  hoverRatio.value = ratioFromEvent(e)
  emit('seek', hoverRatio.value * props.duration)
}

function onPointerUp(e: PointerEvent): void {
  if (!dragging.value) return
  hoverRatio.value = ratioFromEvent(e)
  emit('seek', hoverRatio.value * props.duration)
  emit('seekEnd')
  dragging.value = false
  const el = hitRef.value
  if (el?.hasPointerCapture(e.pointerId)) el.releasePointerCapture(e.pointerId)
}

function onKeydown(e: KeyboardEvent): void {
  if (props.disabled) return
  let target: number | null = null
  switch (e.key) {
    case 'ArrowLeft':
    case 'ArrowDown':  target = props.current - 5; break
    case 'ArrowRight':
    case 'ArrowUp':    target = props.current + 5; break
    case 'PageDown':   target = props.current - 10; break
    case 'PageUp':     target = props.current + 10; break
    case 'Home':       target = 0; break
    case 'End':        target = props.duration; break
    default: return
  }
  e.preventDefault()
  emit('seekStart')
  emit('seek', clamp(target, 0, props.duration))
  emit('seekEnd')
}
</script>

<template>
  <div class="pt" :class="{ 'is-disabled': disabled, 'is-dragging': dragging }">
    <div
      ref="hitRef"
      class="hit"
      role="slider"
      :tabindex="disabled ? -1 : 0"
      :aria-label="A11Y.progress"
      :aria-valuemin="0"
      :aria-valuemax="Math.floor(duration)"
      :aria-valuenow="Math.floor(current)"
      :aria-valuetext="valueText"
      :aria-disabled="disabled ? 'true' : undefined"
      aria-orientation="horizontal"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerUp"
      @keydown="onKeydown"
    >
      <span class="track" aria-hidden="true">
        <span class="buffered" :style="{ width: bufferedPct }"></span>
        <!-- 禁止给 .fill 加任何 transition / animation -->
        <span class="fill" :style="{ width: fillPct }"></span>
        <span class="thumb" :style="{ left: fillPct }"></span>
      </span>
      <!-- 节拍刻度：轨道下 5px，16px 周期 + 32px 稍亮 -->
      <span class="ticks decor" aria-hidden="true"></span>
    </div>
    <span v-if="dragging" class="tip tabular" :style="{ left: tipLeft }">{{ tipText }}</span>
  </div>
</template>

<style scoped>
.pt { position: relative; width: 100%; overflow: visible; }
.hit {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 5px;
  width: 100%;
  /* ★ 热区不占常规流高度：手机两段播放条实测高度由 --bar-h 决定，不被 44px 撑破 */
  min-height: auto;
  cursor: pointer;
  touch-action: none;
  overflow: visible;
}
/* 拖动热区：绝对定位透明伪元素，向外扩 16px，不参与布局高度 */
.hit::before {
  content: '';
  position: absolute;
  inset: -16px 0;
  pointer-events: auto;
  background: transparent;
}
.pt.is-disabled .hit { cursor: default; }
.track {
  position: relative;
  display: block;
  width: 100%;
  /* ★ 轨道本体 4px（hover/拖动 6px）；节拍刻度 5px 是另一个量，见 .ticks */
  height: 4px;
  border-radius: var(--r-xs);
  background: var(--track);
  overflow: visible;
  transition: height var(--dur-instant) var(--ease-out);
}
.pt:hover .track,
.pt.is-dragging .track { height: 6px; }
.pt.is-disabled .track,
.pt.is-disabled:hover .track { height: 4px; }
.buffered {
  position: absolute;
  inset: 0 auto 0 0;
  height: 100%;
  border-radius: var(--r-xs);
  background: var(--track-buffered);
  transition: width var(--dur-base) var(--ease-out);
}
.fill {
  position: absolute;
  inset: 0 auto 0 0;
  height: 100%;
  border-radius: var(--r-xs);
  background: var(--progress-fill); /* 纯色，不用渐变 */
  /* ★ 播放进度必须瞬时跟随：绝对不加 transition */
  transition: none !important;
  animation: none !important;
}
.thumb {
  position: absolute;
  top: 50%;
  width: 12px; height: 12px;
  margin-left: -6px;
  border-radius: 50%;
  background: var(--progress-thumb);
  transform: translateY(-50%);
  box-shadow: 0 0 0 3px var(--signal-soft);
  transition: width var(--dur-instant) var(--ease-out), height var(--dur-instant) var(--ease-out);
}
.pt.is-dragging .thumb { width: 16px; height: 16px; margin-left: -8px; }
.ticks {
  display: block;
  width: 100%;
  height: 5px;
  background-image:
    repeating-linear-gradient(90deg, var(--border-default) 0 1px, transparent 1px 16px),
    repeating-linear-gradient(90deg, var(--signal-dim) 0 1px, transparent 1px 32px);
  opacity: .65;
}
.tip {
  position: absolute;
  bottom: calc(100% + 6px);
  transform: translateX(-50%);
  padding: 3px 7px;
  border: 1px solid var(--border-default);
  border-radius: var(--r-xs);
  background: var(--surface-active);
  color: var(--text-primary);
  font-size: var(--fs-meta);
  pointer-events: none;
  white-space: nowrap;
}
</style>
