/** 时间码 / 时长 / 封面配色：纯函数，无副作用、无持久化。 */

/** 秒 → mm:ss（>=1h 时 h:mm:ss）；非法值统一为 --:-- */
export function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return '--:--'
  const total = Math.floor(seconds)
  const s = total % 60
  const m = Math.floor(total / 60) % 60
  const h = Math.floor(total / 3600)
  const ss = String(s).padStart(2, '0')
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${ss}`
  return `${String(m).padStart(2, '0')}:${ss}`
}

/** aria-valuetext 用：把 mm:ss 说成可读时长 */
export function spokenTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return '未知时长'
  const total = Math.floor(seconds)
  const m = Math.floor(total / 60)
  const s = total % 60
  if (m === 0) return `${s} 秒`
  return s === 0 ? `${m} 分` : `${m} 分 ${s} 秒`
}

/** 0..1 比例 → 百分比整数（音量） */
export function toPercent(ratio: number): number {
  return Math.round(Math.min(1, Math.max(0, ratio)) * 100)
}

/** 字符串 → 稳定色相，用于离线生成抽象封面（无外链图片） */
export function seedHue(seed: string): number {
  let h = 0
  for (let i = 0; i < seed.length; i += 1) {
    h = (h * 31 + seed.charCodeAt(i)) % 360
  }
  return h
}

/** 由歌名生成封面渐变。浅色页用亮底，深色页用暗底，纯离线。 */
export function coverGradient(seed: string, light = false): string {
  const h = seedHue(seed)
  const h2 = (h + 48) % 360
  if (light) return `linear-gradient(135deg, hsl(${h} 42% 78%) 0%, hsl(${h2} 36% 66%) 100%)`
  return `linear-gradient(135deg, hsl(${h} 34% 26%) 0%, hsl(${h2} 28% 14%) 58%, #0E1014 100%)`
}

/** 限制值域 */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}
