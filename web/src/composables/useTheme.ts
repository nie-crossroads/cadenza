import { reactive } from 'vue'

const DARK_KEY = 'cadenza_dark'
const COLOR_KEY = 'cadenza_color'

/** 与博客前台同一组预设色，滑杆按色相连续取色。 */
export const PRESET_COLORS = [
  '#ff4500',
  '#ff8c00',
  '#ffd700',
  '#90ee90',
  '#00ced1',
  '#1e90ff',
  '#a71585',
  '#00d463',
  '#c06070',
  '#c81545',
] as const

export const DEFAULT_COLOR = '#5AC8E8'

interface ThemeState {
  dark: boolean
  color: string
  hue: number
}

const state = reactive<ThemeState>({
  dark: true,
  color: DEFAULT_COLOR,
  hue: 194,
})

function read(key: string): string | null {
  try {
    return localStorage.getItem(key)
  } catch {
    return null
  }
}

function write(key: string, value: string): void {
  try {
    localStorage.setItem(key, value)
  } catch {
    /* 隐私模式写不进时，这一次会话里仍然生效 */
  }
}

function parseHex(hex: string): { r: number; g: number; b: number; solid: string } {
  let h = hex.replace('#', '').trim()
  if (h.length === 3 || h.length === 4) h = [...h].map((c) => c + c).join('')
  const r = Number.parseInt(h.slice(0, 2), 16) || 0
  const g = Number.parseInt(h.slice(2, 4), 16) || 0
  const b = Number.parseInt(h.slice(4, 6), 16) || 0
  return { r, g, b, solid: `#${h.slice(0, 6).toLowerCase()}` }
}

function rgbToHue(r: number, g: number, b: number): number {
  r /= 255
  g /= 255
  b /= 255
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const d = max - min
  if (d === 0) return 0
  let h = 0
  if (max === r) h = ((g - b) / d) % 6
  else if (max === g) h = (b - r) / d + 2
  else h = (r - g) / d + 4
  h *= 60
  if (h < 0) h += 360
  return Math.round(h)
}

function hslToHex(h: number, s: number, l: number): string {
  const a = s * Math.min(l, 1 - l)
  const f = (n: number): string => {
    const k = (n + h / 30) % 12
    const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1)
    return Math.round(255 * color).toString(16).padStart(2, '0')
  }
  return `#${f(0)}${f(8)}${f(4)}`
}

function relLuma(r: number, g: number, b: number): number {
  const f = (c: number): number => {
    const s = c / 255
    return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4
  }
  return 0.2126 * f(r) + 0.7152 * f(g) + 0.0722 * f(b)
}

function sameColor(a: string, b: string): boolean {
  return a.replace('#', '').toLowerCase() === b.replace('#', '').toLowerCase()
}

function apply(): void {
  const root = document.documentElement
  const parsed = parseHex(state.color)
  state.hue = rgbToHue(parsed.r, parsed.g, parsed.b)
  root.classList.toggle('light', !state.dark)
  root.style.setProperty('--hue', String(state.hue))
  root.style.setProperty('--acc', parsed.solid)
  root.style.setProperty('--on-acc', relLuma(parsed.r, parsed.g, parsed.b) > 0.55 ? '#04161C' : '#F7FBFF')
  root.style.colorScheme = state.dark ? 'dark' : 'light'
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', state.dark ? '#07090E' : '#F6F8FB')
}

function setDark(value: boolean): void {
  state.dark = value
  write(DARK_KEY, value ? '1' : '0')
  apply()
}

function setColor(value: string): void {
  state.color = value
  write(COLOR_KEY, value)
  apply()
}

function setHue(value: number): void {
  setColor(hslToHex(value, 0.58, 0.58))
}

function toggle(): void {
  setDark(!state.dark)
}

const storedColor = read(COLOR_KEY)
state.dark = read(DARK_KEY) !== '0'
state.color = storedColor && /^#?[0-9a-fA-F]{6}$/.test(storedColor) ? storedColor : DEFAULT_COLOR
apply()

export function useTheme(): {
  state: ThemeState
  toggle: () => void
  setColor: (value: string) => void
  setHue: (value: number) => void
  sameColor: (a: string, b: string) => boolean
} {
  return { state, toggle, setColor, setHue, sameColor }
}
