<script setup lang="ts">
import { computed } from 'vue'
import type { AddState } from '../api/types'
import { A11Y, COPY } from '../copy/messages'

/** 添加五态：色 + 图标 + 文案三重编码；success / duplicate 为终态。 */
const props = defineProps<{ state: AddState; name: string }>()
const emit = defineEmits<{ add: []; retry: [] }>()

const label = computed(() => {
  switch (props.state.phase) {
    case 'pending':   return COPY.addPending
    case 'success':   return COPY.addSuccess
    case 'duplicate': return COPY.addDuplicate
    case 'error':     return COPY.addError
    default:          return COPY.addIdle
  }
})
const ariaLabel = computed(() =>
  props.state.phase === 'error' ? A11Y.retryAdd(props.name) : A11Y.addTrack(props.name),
)
const locked = computed(() =>
  props.state.phase === 'pending' ||
  props.state.phase === 'success' ||
  props.state.phase === 'duplicate',
)
</script>

<template>
  <div class="add-wrap" :class="`is-${state.phase}`">
    <button
      type="button"
      class="add-btn"
      :aria-label="ariaLabel"
      :aria-busy="state.phase === 'pending' ? 'true' : undefined"
      :disabled="locked"
      @click="state.phase === 'error' ? emit('retry') : emit('add')"
    >
      <!-- idle: plus -->
      <svg v-if="state.phase === 'idle'" class="ico" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
        <path d="M8 3v10M3 8h10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </svg>
      <!-- pending: 三点 -->
      <span v-else-if="state.phase === 'pending'" class="dots" aria-hidden="true"><i></i><i></i><i></i></span>
      <!-- success: ✓ -->
      <svg v-else-if="state.phase === 'success'" class="ico" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
        <path d="M3.2 8.4 6.5 11.6 12.8 4.8" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <!-- duplicate: ⊘ -->
      <svg v-else-if="state.phase === 'duplicate'" class="ico" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
        <circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.5" />
        <path d="M4.2 11.8 11.8 4.2" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
      </svg>
      <!-- error: ↻ -->
      <svg v-else class="ico" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
        <path d="M13 8a5 5 0 1 1-1.6-3.7M13 2.6v3.2H9.8" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <span class="label">{{ label }}</span>
    </button>

    <!-- 409 幂等：中性补充，不用红不用绿、不弹 Toast -->
    <span v-if="state.phase === 'duplicate'" class="note is-duplicate">{{ state.reason }}</span>
    <!-- 失败：行内显示白名单映射后的具体原因，绝不渲染后端 message 原文 -->
    <span v-else-if="state.phase === 'error'" class="note is-error" role="alert">{{ state.reason }}</span>
  </div>
</template>

<style scoped>
.add-wrap {
  /* idle 透明底：无软底、无琥珀色块；四态才叠 --surface + 对应 soft */
  --add-tint: transparent;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
  gap: 2px;
  min-width: 0;
}
.add-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  /* 定宽定高：统一 104×44 / padding 0 16px，不通栏、不铺满整块状态软底 */
  min-height: var(--add-btn-h);
  min-width: var(--add-btn-min-w);
  padding: 0 var(--add-btn-pad);
  border: 1px solid var(--border-default);
  border-style: solid;
  border-radius: var(--r-sm);
  /* ★ idle：透明底 + 1px --border-default + --text-primary，避免 N 行色块墙 */
  background: transparent;
  color: var(--text-primary);
  font-size: var(--fs-label);
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  transition: border-color var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out);
}
/* 状态软底用伪元素叠在 --surface 之上：三态对比度按合成后的实际底色判断 */
.add-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: var(--add-tint);
  pointer-events: none;
}
.add-btn > * { position: relative; }
/* 仅 idle 的 hover 让边/字转 --accent */
.is-idle .add-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
.ico { width: 15px; height: 15px; }
.label { line-height: 1; }

/* pending / success / duplicate / error 四态保留软底（--surface + soft 合成） */
.is-pending { --add-tint: var(--signal-soft); }
.is-pending .add-btn { background: var(--surface); border-color: var(--signal-dim); color: var(--signal); }
/* success：实线边框 + ✓ +「已添加」 */
.is-success { --add-tint: var(--state-success-soft); }
.is-success .add-btn { background: var(--surface); border-color: var(--state-success); border-style: solid; color: var(--state-success); }
/* duplicate：1px dashed 虚线边框 + ⊘ +「已存在」，靠文案 + 图标 + 边框形态区分 */
.is-duplicate { --add-tint: var(--state-duplicate-soft); }
.is-duplicate .add-btn { background: var(--surface); border-color: var(--state-duplicate); border-style: dashed; color: var(--state-duplicate); }
.is-error { --add-tint: var(--state-error-soft); }
.is-error .add-btn { background: var(--surface); border-color: var(--state-error); color: var(--state-error); }
.add-btn:disabled { opacity: 1; } /* 终态保持：不用灰掉来削弱色 + 图标 + 文案编码 */

.dots { display: inline-flex; align-items: center; gap: 3px; height: 15px; }
.dots i {
  width: 3px; height: 3px;
  border-radius: 50%;
  background: currentColor;
  animation: dot-blink 900ms var(--ease-in-out) infinite alternate;
}
.dots i:nth-child(2) { animation-delay: -300ms; }
.dots i:nth-child(3) { animation-delay: -600ms; }
@keyframes dot-blink { from { opacity: .25; transform: scale(.7); } to { opacity: 1; transform: scale(1); } }

.note {
  max-width: 100%;
  font-size: var(--fs-micro);
  line-height: 1.35;
  text-align: right;
}
.note.is-duplicate { color: var(--text-tertiary); }
.note.is-error { color: var(--state-error); }
</style>
