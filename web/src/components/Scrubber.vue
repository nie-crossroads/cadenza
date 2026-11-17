<script setup lang="ts">
import { computed } from 'vue'
import { A11Y } from '../copy/messages'

const props = defineProps<{ ratio: number; disabled?: boolean }>()
const emit = defineEmits<{ seek: [ratio: number] }>()
const percent = computed(() => `${Math.min(100, Math.max(0, props.ratio * 100))}%`)

function ratioFrom(el: HTMLElement, clientX: number): number {
  const rect = el.getBoundingClientRect()
  if (rect.width <= 0) return 0
  return Math.min(1, Math.max(0, (clientX - rect.left) / rect.width))
}

function onPointerDown(event: PointerEvent): void {
  if (props.disabled) return
  const el = event.currentTarget as HTMLElement
  el.setPointerCapture(event.pointerId)
  emit('seek', ratioFrom(el, event.clientX))
}

function onPointerMove(event: PointerEvent): void {
  const el = event.currentTarget as HTMLElement
  if (!el.hasPointerCapture(event.pointerId)) return
  emit('seek', ratioFrom(el, event.clientX))
}

function onKey(event: KeyboardEvent): void {
  if (props.disabled) return
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
  event.preventDefault()
  const step = event.key === 'ArrowRight' ? 0.02 : -0.02
  emit('seek', Math.min(1, Math.max(0, props.ratio + step)))
}
</script>

<template>
  <div
    class="scrub"
    role="slider"
    tabindex="0"
    :aria-label="A11Y.progress"
    :aria-valuemin="0"
    :aria-valuemax="100"
    :aria-valuenow="Math.round(ratio * 100)"
    :aria-disabled="disabled ? 'true' : 'false'"
    @pointerdown.stop="onPointerDown"
    @pointermove="onPointerMove"
    @click.stop
    @keydown="onKey"
  >
    <i class="track" />
    <i class="fill" :style="{ width: percent }" />
    <i class="knob" :style="{ left: percent }" />
  </div>
</template>

<style scoped>
.scrub {
  position: relative;
  height: 18px;
  touch-action: none;
  cursor: pointer;
}
.scrub[aria-disabled="true"] { cursor: default; }
.track, .fill {
  position: absolute;
  left: 0;
  top: 50%;
  height: 2px;
  transform: translateY(-50%);
  border-radius: 999px;
  pointer-events: none;
}
.track {
  right: 0;
  background: color-mix(in srgb, var(--t1) 16%, transparent);
}
.fill { background: var(--acc); }
.knob {
  position: absolute;
  top: 50%;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: var(--acc);
  border: 2px solid var(--bg);
  transform: translate(-50%, -50%);
  pointer-events: none;
}
</style>
