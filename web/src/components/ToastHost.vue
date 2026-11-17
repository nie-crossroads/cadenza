<script setup lang="ts">
import { computed } from 'vue'
import { A11Y } from '../copy/messages'
import { useToast } from '../composables/useToast'

/**
 * Toast：居中底部、播放条上方；桌面 320px / 手机满宽（位置变量集中在 tokens.css）。
 * error tone 6s 且把容器 aria-live 提升为 assertive；duplicate 不产生任何 Toast。
 */
const { state, dismissToast } = useToast()
const hasError = computed(() => state.items.some((i) => i.tone === 'error'))
</script>

<template>
  <div
    class="toast-host"
    role="status"
    :aria-live="hasError ? 'assertive' : 'polite'"
    aria-atomic="false"
  >
    <div
      v-for="item in state.items"
      :key="item.id"
      class="toast"
      :class="`is-${item.tone}`"
    >
      <svg
        v-if="item.tone === 'success'"
        class="ico"
        viewBox="0 0 16 16"
        aria-hidden="true"
        focusable="false"
      >
        <path d="M3.2 8.4 6.5 11.6 12.8 4.8" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <svg
        v-else-if="item.tone === 'error'"
        class="ico"
        viewBox="0 0 16 16"
        aria-hidden="true"
        focusable="false"
      >
        <circle cx="8" cy="8" r="6.2" fill="none" stroke="currentColor" stroke-width="1.6" />
        <path d="M8 4.6v4.4M8 11.2v.2" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </svg>
      <span class="body">
        <span class="text">{{ item.text }}</span>
        <span v-if="item.detail" class="detail">{{ item.detail }}</span>
      </span>
      <button
        type="button"
        class="close"
        :aria-label="A11Y.closeToast"
        @click="dismissToast(item.id)"
      >
        <svg viewBox="0 0 16 16" aria-hidden="true" focusable="false">
          <path d="M4 4l8 8M12 4l-8 8" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
        </svg>
      </button>
    </div>
  </div>
</template>

<style scoped>
.toast-host {
  position: fixed;
  z-index: 70;
  left: 50%;
  bottom: var(--toast-bottom);
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: var(--toast-width);
  pointer-events: none;
}
.toast {
  pointer-events: auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--line-2);
  border-radius: var(--r-m);
  background: var(--surface-2);
  color: var(--t1);
  box-shadow: 0 12px 28px -14px rgba(0, 0, 0, .8);
}
.toast.is-success { border-color: rgba(95, 214, 164, .35); background: rgba(95, 214, 164, .08); }
.toast.is-error { border-color: rgba(244, 115, 111, .35); background: rgba(244, 115, 111, .08); }
.ico { width: 16px; height: 16px; flex: none; }
.toast.is-success .ico { color: var(--ok); }
.toast.is-error .ico { color: var(--err); }
.body { display: flex; flex-direction: column; gap: 2px; min-width: 0; flex: 1; }
.text { font-size: 12px; font-weight: 600; }
.detail { font-size: 11px; color: var(--t3); }
.close {
  width: 28px; height: 28px; color: var(--t4); display: grid; place-items: center;
}
.close svg { width: 12px; height: 12px; }
</style>
