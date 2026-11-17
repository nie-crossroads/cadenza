<script setup lang="ts">
import { computed } from 'vue'
import type { WaveStatus } from '../api/types'
import { playerStatusText } from '../copy/messages'

/** 状态层：波形柱绑真实播放状态，不接 Web Audio。
 *  paused 只把 animation-play-state 改为 paused，冻结当前帧、不重置。
 *  reduced-motion 下动画停止，但静态柱高 + 状态文本 + aria 一条不丢。 */
const props = withDefaults(defineProps<{
  status: WaveStatus
  bars?: number
  tone?: 'signal' | 'accent'
}>(), { bars: 5, tone: 'signal' })

const items = computed(() => Array.from({ length: props.bars }, (_, i) => i))
const statusText = computed(() => playerStatusText(props.status))
</script>

<template>
  <span class="wave" :class="[`is-${status}`, `tone-${tone}`]">
    <span class="sr-only">{{ statusText }}</span>
    <span class="bars" aria-hidden="true">
      <i
        v-for="i in items"
        :key="i"
        class="bar"
        :style="{ '--i': i, '--wb-dur': `${600 + ((i * 97) % 400)}ms` }"
      ></i>
    </span>
  </span>
</template>

<style scoped>
.wave {
  --wb-h: 6px;
  --wb-color: var(--signal);
  display: inline-flex;
  align-items: center;
  height: 20px;
  line-height: 0;
}
.wave.tone-accent { --wb-color: var(--accent); }
.wave.is-playing,
.wave.is-paused { --wb-h: 18px; }
.wave.is-buffering { --wb-h: 12px; }
.wave.is-error { --wb-h: 5px; --wb-color: var(--state-error); }

.bars { display: inline-flex; align-items: center; gap: 3px; height: 20px; }
.bar {
  display: block;
  width: 3px;
  height: var(--wb-h);
  border-radius: 2px;
  background: var(--wb-color);
  transform-origin: center bottom;
  animation: wb-pulse var(--wb-dur) var(--ease-in-out) infinite alternate;
  animation-delay: calc(var(--i) * -110ms);
  animation-play-state: paused;
}
/* playing 错峰律动 600–1000ms；paused 冻结当前帧（不改 animation 定义，故不重启） */
.wave.is-playing .bar { animation-play-state: running; }
.wave.is-buffering .bar {
  animation: wb-shimmer var(--dur-shimmer) var(--ease-in-out) infinite alternate;
  animation-delay: calc(var(--i) * -90ms);
}

@keyframes wb-pulse {
  from { transform: scaleY(.32); opacity: .55; }
  to   { transform: scaleY(1);   opacity: 1; }
}
@keyframes wb-shimmer {
  from { transform: scaleY(.4); opacity: .35; }
  to   { transform: scaleY(1);  opacity: .85; }
}

/* reduced-motion：只停动画，静态柱高 + 状态文本 + aria 全部保留；
   状态文本统一三值「播放中 / 已暂停 / 缓冲中」（见 copy/messages.ts） */
@media (prefers-reduced-motion: reduce) {
  .wave.is-playing { --wb-h: 18px; }   /* 静态满格波形 + 「播放中」 */
  .wave.is-paused { --wb-h: 6px; }     /* 静态矮柱 + 「已暂停」 */
  .wave.is-buffering { --wb-h: 10px; } /* 静态骨架 + 「缓冲中」 */
  .wave.is-idle, .wave.is-error { --wb-h: 5px; }
  .bar { transform: none !important; }
}
</style>
