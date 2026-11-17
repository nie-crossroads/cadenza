import type { Track } from '../api/types'

export interface RowModel {
  key: string
  /** 展示用的主版本；合并行取第一条 */
  primary: Track
  /** 同一歌名 + 作者的各来源版本。长度为 1 时是普通行 */
  variants: Track[]
}

export interface TrackGroup {
  id: string
  label: string
  kind: 'source' | 'merged'
  count: number
  rows: RowModel[]
}

function norm(track: Track): string {
  return `${track.name.trim().toLocaleLowerCase()}\0${track.author.trim().toLocaleLowerCase()}`
}

/** 按来源分组；歌名和作者相同且来源不同的曲目收成「同曲合并」。 */
export function groupTracks(items: Track[], sourceFilter: string): TrackGroup[] {
  const buckets = new Map<string, Track[]>()
  for (const track of items) {
    const key = norm(track)
    const list = buckets.get(key)
    if (list) list.push(track)
    else buckets.set(key, [track])
  }

  const merged: RowModel[] = []
  const singles: Track[] = []
  for (const [key, list] of buckets) {
    const sources = new Set(list.map((t) => t.source))
    if (sources.size > 1) merged.push({ key, primary: list[0], variants: list })
    else singles.push(...list)
  }

  const bySource = new Map<string, Track[]>()
  for (const track of singles) {
    if (sourceFilter !== 'all' && track.source !== sourceFilter) continue
    const list = bySource.get(track.source)
    if (list) list.push(track)
    else bySource.set(track.source, [track])
  }

  const groups: TrackGroup[] = []
  for (const [source, tracks] of bySource) {
    groups.push({
      id: `src:${source}`,
      label: source || '未标明来源',
      kind: 'source',
      count: tracks.length,
      rows: tracks.map((track) => ({ key: track.sourceId, primary: track, variants: [track] })),
    })
  }

  const mergedRows = merged
    .map((row) => {
      const variants = sourceFilter === 'all'
        ? row.variants
        : row.variants.filter((t) => t.source === sourceFilter)
      if (variants.length === 0) return null
      if (variants.length === 1) {
        const only = variants[0]
        const id = `src:${only.source}`
        let group = groups.find((g) => g.id === id)
        if (!group) {
          group = { id, label: only.source || '未标明来源', kind: 'source', count: 0, rows: [] }
          groups.push(group)
        }
        group.rows.push({ key: only.sourceId, primary: only, variants })
        group.count = group.rows.length
        return null
      }
      return { ...row, variants, primary: variants[0] }
    })
    .filter((row): row is RowModel => row !== null)

  if (mergedRows.length > 0) {
    const mergedGroup: TrackGroup = {
      id: 'merged',
      label: '',
      kind: 'merged',
      count: mergedRows.length,
      rows: mergedRows,
    }
    groups.splice(Math.min(1, groups.length), 0, mergedGroup)
  }

  return groups
}

const DOTS = ['#E5344B', '#2E9BFF', '#31C27C', '#5AC8E8', '#9B8CFF', '#E8B45F']
const NAMED: Record<string, string> = {
  网易: '#E5344B',
  ＱＱ: '#31C27C',
  酷狗: '#2E9BFF',
  酷我: '#E8B45F',
  千千: '#9B8CFF',
  一听: '#5AC8E8',
  咪咕: '#F07A3A',
}

export function sourceColor(source: string): string {
  const named = NAMED[source]
  if (named) return named
  let h = 0
  for (let i = 0; i < source.length; i += 1) h = (h + source.charCodeAt(i) * (i + 1)) % DOTS.length
  return DOTS[h]
}
