<script setup lang="ts">
/** 旋转黑胶：纯 CSS 装饰层，不可聚焦、不进 a11y 树。
 *  暂停只改 animation-play-state，保留当前旋转角度；reduced-motion 停转但保留沟槽。 */
withDefaults(defineProps<{ spinning?: boolean }>(), { spinning: false })
</script>

<template>
  <div class="vinyl" :class="{ 'is-spinning': spinning }" aria-hidden="true">
    <span class="grooves"></span>
    <span class="label"></span>
    <span class="hole"></span>
    <span class="shine"></span>
  </div>
</template>

<style scoped>
.vinyl {
  position: relative;
  flex: none;
  width: var(--disc-size);
  height: var(--disc-size);
  border-radius: 50%;
  background: radial-gradient(circle at 32% 28%, #2A2D34 0%, #101216 58%, #08090B 100%);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, .06);
  pointer-events: none;
  animation: vinyl-spin var(--dur-spin) linear infinite;
  animation-play-state: paused;
}
.vinyl.is-spinning { animation-play-state: running; }

.grooves {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: repeating-radial-gradient(circle at 50% 50%, rgba(255, 255, 255, .06) 0 2px, transparent 2px 3px);
  -webkit-mask-image: radial-gradient(circle, transparent 0 21%, #000 24% 95%, transparent 100%);
  mask-image: radial-gradient(circle, transparent 0 21%, #000 24% 95%, transparent 100%);
}
.label {
  position: absolute;
  inset: 33%;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: inset 0 0 0 1px rgba(23, 18, 10, .35);
}
.hole {
  position: absolute;
  top: 50%; left: 50%;
  width: 4px; height: 4px;
  border-radius: 50%;
  background: var(--bg-base);
  transform: translate(-50%, -50%);
}
.shine {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: conic-gradient(from 205deg, transparent 0 18%, rgba(255, 255, 255, .055) 24%, transparent 31% 100%);
}
@keyframes vinyl-spin { to { transform: rotate(360deg); } }
</style>
