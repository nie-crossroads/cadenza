/**
 * 离线演示适配器【仅 VITE_API_MODE=mock 时启用，Go 侧接口冻结后删除本文件】。
 * 数据全部虚构，无任何真实源站 / 存储 / 临时地址。音质只用于列表标签。
 * 只用于本机 dev 预览，不参与生产构建的功能路径。
 */
import type { Track } from './types'

/** cover 留空：封面由 coverGradient(歌名) 在前端离线抽象生成，不加载任何外部图片。 */
/** 虚构来源名，只用于离线演示分组。页面不请求这些名字对应的任何站点。 */
export const MOCK_TRACKS: Track[] = [
  { name: 'Merry Christmas Mr. Lawrence', author: '坂本龍一', cover: '', duration: 278, source: '谱架甲', sourceId: 'demo-a-001', quality: 'hi' },
  { name: 'aqua', author: '坂本龍一', cover: '', duration: 328, source: '谱架甲', sourceId: 'demo-a-002', quality: 'hi' },
  { name: 'energy flow', author: '坂本龍一', cover: '', duration: 275, source: '谱架甲', sourceId: 'demo-a-003', quality: 'hi' },
  { name: 'energy flow', author: '坂本龍一', cover: '', duration: 276, source: '谱架乙', sourceId: 'demo-a-003b', quality: 'hr' },
  { name: 'energy flow', author: '坂本龍一', cover: '', duration: 273, source: '谱架丙', sourceId: 'demo-a-003c-dup', quality: 'std' },
  { name: 'The Last Emperor', author: '坂本龍一', cover: '', duration: 304, source: '谱架甲', sourceId: 'demo-a-004', quality: 'std' },
  { name: '戦場のメリークリスマス', author: '坂本龍一', cover: '', duration: 372, source: '谱架乙', sourceId: 'demo-a-005', quality: 'hr' },
  { name: 'Solitude', author: '坂本龍一', cover: '', duration: 202, source: '谱架乙', sourceId: 'demo-a-006-dup', quality: 'std' },
  { name: 'Merry Christmas Mr. Lawrence', author: '坂本龍一', cover: '', duration: 281, source: '谱架乙', sourceId: 'demo-a-007', quality: 'std' },
  { name: 'Rain', author: '坂本龍一', cover: '', duration: 214, source: '谱架丙', sourceId: 'demo-a-008-fail', quality: 'std' },
]

let mockAuthed = false

function abortError(): DOMException {
  return new DOMException('Aborted', 'AbortError')
}

function wait(ms: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) { reject(abortError()); return }
    const onAbort = (): void => {
      window.clearTimeout(timer)
      reject(abortError())
    }
    const timer = window.setTimeout(() => {
      signal?.removeEventListener('abort', onAbort)
      resolve()
    }, ms)
    signal?.addEventListener('abort', onAbort, { once: true })
  })
}

export function mockProbe(): boolean { return mockAuthed }

export function mockLogout(): void {
  mockAuthed = false
}

export async function mockLogin(_password: string): Promise<void> {
  await wait(420)
  mockAuthed = true
}

export async function mockSearch(q: string, signal: AbortSignal): Promise<Track[]> {
  await wait(560, signal)
  const key = q.trim().toLocaleLowerCase()
  return MOCK_TRACKS.filter(
    (t) => t.name.toLocaleLowerCase().includes(key) || t.author.toLocaleLowerCase().includes(key),
  )
}

export async function mockPlayable(sourceId: string): Promise<boolean> {
  await wait(280)
  return !sourceId.endsWith('-fail')
}

export async function mockAdd(sourceId: string): Promise<string> {
  await wait(900)
  if (sourceId.endsWith('-dup')) return 'ADD_DUPLICATE'
  if (sourceId.endsWith('-fail')) return 'ADD_DOWNLOAD_FAILED'
  return 'OK'
}
