<script setup lang="ts">
import { computed } from 'vue'
import { BRAND } from '../copy/messages'

/** variant=display 用于登录页（40px / 手机 28px），bar 用于主页（22px / 手机 18px）。 */
const props = withDefaults(defineProps<{ variant?: 'display' | 'bar' }>(), { variant: 'bar' })
const isDisplay = computed(() => props.variant === 'display')
</script>

<template>
  <header class="brand" :class="{ 'is-display': isDisplay }">
    <!-- 装饰层：五线谱淡线，两端 mask 淡出，opacity ≤ .06 -->
    <span class="staff decor" aria-hidden="true"></span>

    <component :is="isDisplay ? 'h1' : 'span'" class="wordmark">
      {{ BRAND.name }}
      <!-- Fermata 延长记号：弧线 + 圆点，字标右上角 -->
      <svg class="fermata decor" viewBox="0 0 24 16" aria-hidden="true" focusable="false">
        <path d="M2 13 C 4 3, 20 3, 22 13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <circle cx="12" cy="4.4" r="2.1" fill="currentColor" />
      </svg>
    </component>

    <span class="corner">
      <!-- Crossroads 十字准星：交叉线 + 空心圆 -->
      <svg class="crosshair decor" viewBox="0 0 12 12" aria-hidden="true" focusable="false">
        <circle cx="6" cy="6" r="3.6" fill="none" stroke="currentColor" stroke-width="1" />
        <path d="M6 0v3.2M6 8.8V12M0 6h3.2M8.8 6H12" fill="none" stroke="currentColor" stroke-width="1" />
      </svg>
      {{ BRAND.corner }}
    </span>
  </header>
</template>

<style scoped>
.brand {
  position: sticky;
  top: 0;
  z-index: 40;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  height: var(--brand-h);
  min-height: var(--brand-h);
  padding: 0 max(var(--gutter), calc((100% - var(--content-max)) / 2));
  border-bottom: 1px solid var(--border-subtle);
  background: var(--glass-bg);
  -webkit-backdrop-filter: var(--glass-blur);
  backdrop-filter: var(--glass-blur);
  overflow: hidden;
}
.staff {
  position: absolute;
  inset: 0;
  opacity: .06;
  background: repeating-linear-gradient(to bottom, var(--text-secondary) 0 1px, transparent 1px 12px);
  -webkit-mask-image: linear-gradient(to right, transparent, #000 30%, #000 70%, transparent);
  mask-image: linear-gradient(to right, transparent, #000 30%, #000 70%, transparent);
}
/* C4：登录页字标展示态的五线谱略提到 .09；主页仍 .06，周期差异保持现状 */
.brand.is-display .staff { opacity: .09; }
.wordmark {
  position: relative;
  display: inline-flex;
  align-items: flex-start;
  margin: 0;
  font-size: var(--fs-brand);
  font-weight: 600;
  letter-spacing: -.02em;
  line-height: 1.1;
  color: var(--text-primary);
}
.brand.is-display .wordmark { font-size: var(--fs-display); }
.fermata {
  width: 20px; height: 14px;
  margin-left: 2px;
  margin-top: -2px;
  color: var(--accent);
}
.brand.is-display .fermata { width: 26px; height: 18px; }
.corner {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: var(--fs-crossroads);
  font-weight: 500;
  letter-spacing: .18em;
  text-transform: uppercase;
  color: var(--text-tertiary);
  white-space: nowrap;
}
.crosshair { width: 12px; height: 12px; color: var(--text-tertiary); }
</style>
