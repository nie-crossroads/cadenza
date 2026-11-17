<script setup lang="ts">
import { computed } from 'vue'
import type { WaveStatus } from '../api/types'
import { A11Y, staticStatusText } from '../copy/messages'
import EqLamp from './EqLamp.vue'
import WaveBars from './WaveBars.vue'

/**
 * 播放条状态区：波形柱 + EQ 状态灯 + 可见静态状态文本。
 * ★ 静态文本固定三值「播放中 / 已暂停 / 缓冲中」，--fs-meta 12px --text-secondary；
 *   不用 aria-live 高频播报播放状态（读屏会刷屏）。
 * ★ EQ 呼吸灯是状态层，通过 EqLamp 的 role=img + aria-label 留在无障碍树。
 * 桌面 7 柱 / 手机 5 柱，由父级以 --desktop-status-display / --mobile-status-display 切换。
 */
const props = withDefaults(defineProps<{
  status: WaveStatus
  bars: number
  variant: 'desktop' | 'mobile'
}>(), { bars: 7 })

const text = computed(() => staticStatusText(props.status))
const lampLabel = computed(() => A11Y.eqLamp(text.value))
const active = computed(() => props.status === 'playing')
</script>

<template>
  <span class="pstatus" :class="`pstatus-${variant}`">
    <WaveBars :status="status" :bars="bars" />
    <EqLamp :active="active" :label="lampLabel" />
    <span class="ptext">{{ text }}</span>
  </span>
</template>

<style scoped>
.pstatus {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.pstatus-desktop { display: var(--desktop-status-display); }
.pstatus-mobile { display: var(--mobile-status-display); }
.ptext {
  font-size: var(--fs-meta);
  color: var(--text-secondary); /* 实色 token，不用 opacity / rgba / color-mix */
  white-space: nowrap;
}
</style>
