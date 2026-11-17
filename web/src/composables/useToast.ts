import { reactive } from 'vue'

export interface ToastItem {
  id: number
  text: string
  tone: 'success' | 'neutral' | 'error'
  /** 失败原因副标题（12px）；只取白名单映射后的中文，绝不透传后端原文 */
  detail: string
}

/** 内存队列：普通 4000ms / error 6000ms 自动出队；Toast 不抢焦点。 */
const state = reactive<{ items: ToastItem[] }>({ items: [] })
let seq = 0
const timers = new Map<number, number>()

export function pushToast(
  text: string,
  tone: ToastItem['tone'] = 'neutral',
  detail = '',
): void {
  const id = ++seq
  state.items.push({ id, text, tone, detail })
  while (state.items.length > 2) {
    const oldest = state.items[0]
    if (oldest) dismissToast(oldest.id)
    else break
  }
  const ttl = tone === 'error' ? 6000 : 4000
  timers.set(id, window.setTimeout(() => dismissToast(id), ttl))
}

export function dismissToast(id: number): void {
  const timer = timers.get(id)
  if (timer !== undefined) {
    window.clearTimeout(timer)
    timers.delete(id)
  }
  const index = state.items.findIndex((item) => item.id === id)
  if (index >= 0) state.items.splice(index, 1)
}

export function clearToasts(): void {
  for (const timer of timers.values()) window.clearTimeout(timer)
  timers.clear()
  state.items.splice(0, state.items.length)
}

export function useToast() {
  return { state, pushToast, dismissToast, clearToasts }
}
