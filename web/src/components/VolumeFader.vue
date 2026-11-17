<script setup lang="ts">
import { computed, ref } from 'vue'
import { A11Y } from '../copy/messages'
import { clamp, toPercent } from '../utils/format'

/** 音量推子 + 静音记忆（两态图标 + aria-pressed）；EQ 呼吸灯属于状态层，由 PlayerStatus 统一提供。 */
const props = withDefaults(defineProps<{
  volume: number
  muted: boolean
  /** 空播放态：控件禁用但可见 */
  disabled?: boolean
}>(), { disabled: false })
const emit = defineEmits<{ change: [ratio: number]; mute: [] }>()

const trackRef = ref<HTMLElement | null>(null)
const dragging = ref(false)
const pct = computed(() => toPercent(props.muted ? 0 : props.volume))
const valueText = computed(() => A11Y.volumeValue(pct.value))

function ratioFromEvent(e: PointerEvent): number {
  const el = trackRef.value
  if (!el) return 0
  const rect = el.getBoundingClientRect()
  if (rect.width <= 0) return 0
  return clamp((e.clientX - rect.left) / rect.width, 0, 1)
}
function onDown(e: PointerEvent): void {
  if (props.disabled) return
  dragging.value = true
  trackRef.value?.setPointerCapture(e.pointerId)
  emit('change', ratioFromEvent(e))
  e.preventDefault()
}
function onMove(e: PointerEvent): void {
  if (!props.disabled && dragging.value) emit('change', ratioFromEvent(e))
}
function onUp(e: PointerEvent): void {
  if (!dragging.value) return
  dragging.value = false
  const el = trackRef.value
  if (el?.hasPointerCapture(e.pointerId)) el.releasePointerCapture(e.pointerId)
}
function onKeydown(e: KeyboardEvent): void {
  if (props.disabled) return
  let next: number | null = null
  if (e.key === 'ArrowLeft' || e.key === 'ArrowDown') next = props.volume - 0.05
  else if (e.key === 'ArrowRight' || e.key === 'ArrowUp') next = props.volume + 0.05
  else if (e.key === 'Home') next = 0
  else if (e.key === 'End') next = 1
  else return
  e.preventDefault()
  emit('change', clamp(next, 0, 1))
}
</script>

<template>
  <div class="fader">
    <button
      type="button"
      class="mute"
      :aria-label="muted ? A11Y.unmute : A11Y.mute"
      :aria-pressed="muted"
      :disabled="disabled"
      @click="emit('mute')"
    >
      <svg viewBox="0 0 20 20" aria-hidden="true" focusable="false">
        <path d="M3 8h3l4-3.4v10.8L6 12H3z" fill="currentColor" />
        <path v-if="!muted" d="M13 7.2a4 4 0 0 1 0 5.6M15.4 5a7 7 0 0 1 0 10" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
        <path v-else d="M13.4 8.2l4 3.6M17.4 8.2l-4 3.6" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" />
      </svg>
    </button>

    <div
      ref="trackRef"
      class="track"
      role="slider"
      :tabindex="disabled ? -1 : 0"
      :aria-label="A11Y.volume"
      :aria-valuemin="0"
      :aria-valuemax="100"
      :aria-valuenow="pct"
      :aria-valuetext="valueText"
      :aria-disabled="disabled ? 'true' : undefined"
      aria-orientation="horizontal"
      @pointerdown="onDown"
      @pointermove="onMove"
      @pointerup="onUp"
      @pointercancel="onUp"
      @keydown="onKeydown"
    >
      <span class="fill" :style="{ width: `${pct}%` }"></span>
      <span class="knob" :style="{ left: `${pct}%` }"></span>
    </div>

    <!-- EQ 呼吸灯统一由 PlayerStatus（状态层）提供，不放进音量推子的装饰层 -->
  </div>
</template>

<style scoped>
.fader {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 96px;
  flex: none;
}
.mute {
  display: grid;
  place-items: center;
  width: 44px; height: 44px;
  flex: none;
  border-radius: var(--r-sm);
  color: var(--text-secondary);
}
.mute:hover:not(:disabled) { color: var(--signal); background: var(--signal-soft); }
.mute:disabled { color: var(--text-disabled); }
.mute svg { width: 20px; height: 20px; }
.track {
  position: relative;
  flex: 1;
  min-width: 56px;
  height: 44px;
  display: flex;
  align-items: center;
  cursor: pointer;
  touch-action: none;
}
.track::before {
  content: '';
  position: absolute;
  left: 0; right: 0;
  height: 4px;
  border-radius: 2px;
  background: var(--track);
}
.fill {
  position: absolute;
  left: 0;
  height: 4px;
  border-radius: 2px;
  background: var(--signal);
  transition: width var(--dur-fast) var(--ease-out);
}
.knob {
  position: absolute;
  width: 10px; height: 10px;
  margin-left: -5px;
  border-radius: 50%;
  background: var(--progress-thumb);
}
</style>
