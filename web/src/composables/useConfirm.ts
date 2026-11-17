import { reactive } from 'vue'

interface ConfirmState {
  open: boolean
  name: string
  resolve: ((ok: boolean) => void) | null
}

const state = reactive<ConfirmState>({
  open: false,
  name: '',
  resolve: null,
})

export function useConfirm() {
  return { state }
}

/** 添加前的确认。取消或被新的确认替换时返回 false。 */
export function confirmAdd(name: string): Promise<boolean> {
  if (state.resolve) state.resolve(false)
  return new Promise((resolve) => {
    state.name = name
    state.open = true
    state.resolve = resolve
  })
}

export function settleConfirm(ok: boolean): void {
  const resolve = state.resolve
  state.open = false
  state.name = ''
  state.resolve = null
  resolve?.(ok)
}
