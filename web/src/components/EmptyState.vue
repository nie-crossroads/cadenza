<script setup lang="ts">
import { computed } from 'vue'
import { COPY, fill } from '../copy/messages'

const props = withDefaults(defineProps<{
  kind: 'empty' | 'noResults' | 'error'
  keyword?: string
  errorHint?: string
}>(), { keyword: '', errorHint: '' })

/** retry：搜索失败重试；focus-search：把焦点送回搜索框（不抢焦点、不改变状态） */
const emit = defineEmits<{ retry: []; 'focus-search': [] }>()

const title = computed(() => {
  if (props.kind === 'noResults') return fill(COPY.noResultsTitle, { kw: props.keyword })
  if (props.kind === 'error') return COPY.searchFailedTitle
  return COPY.emptyTitle
})
const sub = computed(() => {
  if (props.kind === 'noResults') return COPY.noResultsSub
  if (props.kind === 'error') return props.errorHint || COPY.searchFailedSub
  return COPY.emptySub
})
</script>

<template>
  <div class="empty" :class="`is-${kind}`">
    <!-- 装饰层：五线谱淡线 opacity .09（≤ .1，aria-hidden + pointer-events:none） -->
    <span class="staff decor" aria-hidden="true"></span>

    <!-- C4：空状态黑胶主视觉 桌面 240 / 平板 200 / 手机 160，
         沟槽 + 琥珀中心标签 + 中心孔 + 外圈高光都用实色/不透明纹理，整体 opacity .92 ≥ .5 -->
    <span class="vinyl-hero decor" aria-hidden="true">
      <span class="vh-grooves"></span>
      <span class="vh-label"></span>
      <span class="vh-hole"></span>
      <span class="vh-shine"></span>
    </span>

    <p class="title">{{ title }}</p>
    <p class="sub">{{ sub }}</p>

    <button
      v-if="kind === 'error'"
      type="button"
      class="retry"
      @click="emit('retry')"
    >{{ COPY.retry }}</button>
    <!-- accent 文字按钮：把焦点送回搜索框；不抢焦点、不改变状态 -->
    <button
      v-else
      type="button"
      class="focus-btn"
      @click="emit('focus-search')"
    >{{ COPY.emptyFocus }}</button>
    <!-- 不抢焦点：空态 / 无结果 / 错误都靠外层 role=status 播报 -->
  </div>
</template>

<style scoped>
.empty {
  position: relative;
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 0;
  padding: 20px var(--gutter);
  text-align: center;
  overflow: hidden;
}
.staff {
  position: absolute;
  inset: 0;
  /* C4：空状态五线谱提到 .09，仍 ≤ .1 */
  opacity: .09;
  background: repeating-linear-gradient(to bottom, var(--text-secondary) 0 1px, transparent 1px 14px);
  -webkit-mask-image: radial-gradient(ellipse at center, #000 0%, transparent 72%);
  mask-image: radial-gradient(ellipse at center, #000 0%, transparent 72%);
}

/* ---- 黑胶主视觉 ---- */
.vinyl-hero {
  position: relative;
  flex: none;
  width: var(--empty-disc);
  height: var(--empty-disc);
  margin-bottom: 12px;
  border-radius: 50%;
  background: radial-gradient(circle at 32% 28%, #343842 0%, #171A20 52%, #08090B 100%);
  /* 外圈高光 + 一圈实色描边，保证在深色底上轮廓清晰 */
  box-shadow:
    inset 0 0 0 1px rgba(255, 255, 255, .14),
    0 0 0 1px var(--border-default),
    0 18px 48px rgba(0, 0, 0, .45);
  opacity: .92;
}
.vh-grooves {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: repeating-radial-gradient(circle at 50% 50%, rgba(255, 255, 255, .11) 0 2px, transparent 2px 5px);
  -webkit-mask-image: radial-gradient(circle, transparent 0 21%, #000 25% 95%, transparent 100%);
  mask-image: radial-gradient(circle, transparent 0 21%, #000 25% 95%, transparent 100%);
}
.vh-label {
  position: absolute;
  inset: 33%;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: inset 0 0 0 1px rgba(23, 18, 10, .45);
}
.vh-hole {
  position: absolute;
  top: 50%; left: 50%;
  width: 10px; height: 10px;
  border-radius: 50%;
  background: var(--bg-base);
  box-shadow: 0 0 0 2px rgba(23, 18, 10, .5);
  transform: translate(-50%, -50%);
}
.vh-shine {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: conic-gradient(from 205deg, transparent 0 16%, rgba(255, 255, 255, .10) 24%, transparent 32% 100%);
}

.title {
  position: relative;
  /* C4：主文案 20px / 600 */
  font-size: var(--fs-h2);
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -.01em;
}
.sub {
  position: relative;
  /* C4：副文案 13px 实色 token */
  font-size: var(--fs-aux);
  color: var(--text-tertiary);
  max-width: 30em;
}
/* 无结果用中性色，不用 error 色 */
.empty.is-noResults .title { color: var(--text-secondary); }
.empty.is-error .title { color: var(--text-primary); }
.retry {
  position: relative;
  margin-top: 10px;
  min-height: 44px;
  padding: 0 20px;
  border: 1px solid var(--border-default);
  border-radius: var(--r-sm);
  color: var(--text-primary);
  font-size: var(--fs-label);
  font-weight: 500;
  background: var(--surface);
  transition: background var(--dur-fast) var(--ease-out);
}
.retry:hover { background: var(--surface-hover); }
/* accent 文字按钮：无底色、无边框，靠字色表达主操作 */
.focus-btn {
  position: relative;
  margin-top: 6px;
  min-height: 44px;
  padding: 0 12px;
  border-radius: var(--r-sm);
  color: var(--accent);
  font-size: var(--fs-label);
  font-weight: 600;
  background: transparent;
  transition: color var(--dur-fast) var(--ease-out), background var(--dur-fast) var(--ease-out);
}
.focus-btn:hover { color: var(--accent-hover); background: var(--accent-soft); }
</style>
