<script setup lang="ts">
/**
 * EQ 呼吸灯 —— 状态层，不是装饰层。
 * ★ 不设 aria-hidden、不套 .decor 只读装饰容器：
 *   通过 role="img" + aria-label 把播放状态留在无障碍树里。
 * reduced-motion 下停止呼吸动画，但静态指示（实色灯）保留。
 */
withDefaults(defineProps<{ active?: boolean; label?: string }>(), {
  active: false,
  label: '播放状态指示',
})
</script>

<template>
  <span class="eq" :class="{ 'is-on': active }" role="img" :aria-label="label"></span>
</template>

<style scoped>
.eq {
  display: inline-block;
  flex: none;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--signal-dim); /* 实色，不用 opacity < 1 */
}
.eq.is-on {
  background: var(--signal);
  animation: eq-breathe 1600ms var(--ease-in-out) infinite alternate;
}
@keyframes eq-breathe {
  from { box-shadow: 0 0 0 0 var(--signal-soft); }
  to   { box-shadow: 0 0 0 4px var(--signal-soft); }
}
@media (prefers-reduced-motion: reduce) {
  .eq.is-on { animation: none; } /* 停呼吸，静态指示保留 */
}
</style>
