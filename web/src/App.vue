<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import ToastHost from './components/ToastHost.vue'
import ConfirmHost from './components/ConfirmHost.vue'
import { onSessionExpired } from './composables/useSession'
import { togglePlay } from './composables/usePlayer'
import { resetAddStates } from './composables/useAdd'
import { clearToasts } from './composables/useToast'

let unregisterExpired: (() => void) | null = null

onMounted(() => {
  // 会话失效只收起添加能力。不跳登录页，搜索和试听继续留在当前页。
  unregisterExpired = onSessionExpired(() => {
    resetAddStates()
    clearToasts()
  })
  window.addEventListener('keydown', onGlobalKeydown)
})

onUnmounted(() => {
  unregisterExpired?.()
  window.removeEventListener('keydown', onGlobalKeydown)
})

/** 空格在非表单焦点切换播放暂停并 preventDefault；
 *  焦点在按钮 / 输入框 / slider 上交给原生，避免双触发。 */
function onGlobalKeydown(e: KeyboardEvent): void {
  if (e.code !== 'Space' && e.key !== ' ') return
  const el = document.activeElement as HTMLElement | null
  if (!el) return
  const tag = el.tagName
  const isSlider = el.getAttribute('role') === 'slider'
  const interactive =
    tag === 'BUTTON' || tag === 'INPUT' || tag === 'TEXTAREA' ||
    tag === 'SELECT' || tag === 'A' || isSlider || el.isContentEditable
  if (interactive) {
    if (isSlider) e.preventDefault()
    return
  }
  e.preventDefault()
  togglePlay()
}
</script>

<template>
  <!-- A6 噪点：纯装饰层，aria-hidden + pointer-events:none，手机由 --noise-display 关闭 -->
  <RouterView />
  <ToastHost />
  <ConfirmHost />
</template>
