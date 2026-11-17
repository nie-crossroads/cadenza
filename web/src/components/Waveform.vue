<script setup lang="ts">
import { computed } from 'vue'
import { A11Y } from '../copy/messages'

const props = withDefaults(defineProps<{
  seed: string
  /** 0 到 1 */
  progress?: number
  active?: boolean
}>(), { progress: 0, active: false })

const emit = defineEmits<{ seek: [ratio: number] }>()

const bars = computed(() => {
  const n = 40
  const out: number[] = []
  let s = 0
  for (let i = 0; i < props.seed.length; i += 1) s = (s * 33 + props.seed.charCodeAt(i)) % 997
  for (let i = 0; i < n; i += 1) {
    s = (s * 9301 + 49297) % 233280
    const r = s / 233280
    const wave = 0.18 + 0.5 * Math.abs(Math.sin(i * 0.63 + r)) + 0.28 * Math.abs(Math.sin(i * 1.7))
    out.push(Math.round(Math.min(1, wave) * 100))
  }
  return out
})

function onClick(e: MouseEvent): void {
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  if (rect.width <= 0) return
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
  emit('seek', ratio)
}
</script>

<template>
  <div
    class="wf"
    :class="{ 'is-on': active }"
    role="slider"
    :aria-label="A11Y.progress"
    :aria-valuemin="0"
    :aria-valuemax="100"
    :aria-valuenow="Math.round(progress * 100)"
    tabindex="0"
    @click="onClick"
    @keydown.left.prevent="emit('seek', Math.max(0, progress - 0.05))"
    @keydown.right.prevent="emit('seek', Math.min(1, progress + 0.05))"
  >
    <i
      v-for="(h, i) in bars"
      :key="i"
      :style="{ height: `${Math.max(12, h)}%` }"
      :class="{ on: (i + 1) / bars.length <= progress }"
    />
  </div>
</template>

<style scoped>
.wf {
  display: flex;
  align-items: center;
  gap: 2px;
  height: 26px;
  cursor: pointer;
  opacity: .72;
}
.wf:hover { opacity: 1; }
.wf i {
  flex: 1;
  background: var(--line-3);
  border-radius: 1px;
  min-height: 2px;
}
.wf.is-on i { background: color-mix(in srgb, var(--acc) 28%, transparent); }
.wf i.on,
.wf.is-on i.on { background: var(--acc); }
@media (prefers-reduced-motion: no-preference) {
  .wf.is-on i.on { animation: barPulse 1.1s ease-in-out infinite; }
}
@keyframes barPulse {
  0%, 100% { transform: scaleY(1); }
  50% { transform: scaleY(.72); }
}
</style>
