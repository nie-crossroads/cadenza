<script setup lang="ts">
import { computed } from 'vue'
import type { AddState, Track, WaveStatus } from '../api/types'
import { A11Y, COPY, playerStatusText } from '../copy/messages'
import { coverGradient, formatTime } from '../utils/format'
import { state as playerState, toggleAudition, retryPreview } from '../composables/usePlayer'
import { states, requestAdd } from '../composables/useAdd'
import WaveBars from './WaveBars.vue'
import EqLamp from './EqLamp.vue'
import AddButton from './AddButton.vue'

const props = defineProps<{ track: Track }>()

/** DOM 顺序 = 视觉顺序 = Tab 顺序：cover → audition → meta → duration → source → add
 *  桌面 Tab 顺序即「试听 → 添加」。手机仅重排 grid areas，不反转 DOM。 */
const isCurrent = computed(() => playerState.track?.sourceId === props.track.sourceId)
const playing = computed(() => isCurrent.value && playerState.status === 'playing')
const showPause = computed(() => isCurrent.value && (playing.value || playerState.status === 'buffering'))

const waveStatus = computed<WaveStatus>(() => {
  if (!isCurrent.value) return 'idle'
  switch (playerState.status) {
    case 'playing':   return 'playing'
    case 'buffering': return 'buffering'
    case 'error':     return 'error'
    case 'ended':
    case 'paused':    return 'paused'
    default:          return 'idle'
  }
})

const auditionLabel = computed(() => {
  if (!isCurrent.value) return A11Y.audition(props.track.name)
  if (playerState.status === 'error') return A11Y.retryAudition(props.track.name)
  if (showPause.value) return A11Y.pauseAudition(props.track.name)
  return A11Y.resumeAudition(props.track.name)
})
const auditionText = computed(() =>
  isCurrent.value ? playerStatusText(playerState.status) : COPY.statusIdle,
)
const eqLabel = computed(() => A11Y.eqLamp(auditionText.value))
const coverStyle = computed(() => ({ background: coverGradient(props.track.name) }))
/** 手机：歌手 + 时长合并为「同一文本节点」，避免重复读屏 */
const mobileAuthorLine = computed(() => `${props.track.author} · ${formatTime(props.track.duration)}`)
const IDLE_ADD: AddState = { phase: 'idle', reason: '' }
const addState = computed<AddState>(() => states[props.track.sourceId] ?? IDLE_ADD)

function onAudition(): void {
  if (isCurrent.value && playerState.status === 'error') {
    retryPreview()
    return
  }
  toggleAudition(props.track)
}
function onAdd(): void { void requestAdd(props.track) }
function onRetry(): void { void requestAdd(props.track) }
</script>

<template>
  <li
    class="row"
    :class="{ 'is-current': isCurrent, 'is-playing': playing }"
    :data-source-id="track.sourceId"
    :aria-current="isCurrent ? 'true' : undefined"
  >
    <!-- 1. cover：离线抽象封面（CSS 渐变），无外链图片 -->
    <span class="cover" :style="coverStyle" aria-hidden="true">
      <span class="cover-rings decor"></span>
    </span>

    <!-- 2. audition：桌面波形 / 手机四件套（竖条 + 暂停图标 + EQ 灯 + sr-only） -->
    <div class="audition">
      <button
        type="button"
        class="aud-btn"
        :aria-label="auditionLabel"
        :aria-pressed="isCurrent ? 'true' : 'false'"
        @click="onAudition"
      >
        <WaveBars class="wave-desktop" :status="waveStatus" :bars="5" />
        <!-- 手机隐藏行内波形柱，改由图标表达播放/暂停 -->
        <svg v-if="showPause" class="aud-icon" viewBox="0 0 20 20" aria-hidden="true" focusable="false">
          <path d="M6.5 4.5v11M13.5 4.5v11" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" />
        </svg>
        <svg v-else class="aud-icon" viewBox="0 0 20 20" aria-hidden="true" focusable="false">
          <path d="M6.5 4.2 15.4 10 6.5 15.8z" fill="currentColor" />
        </svg>
        <span class="aud-text">{{ auditionText }}</span>
      </button>
      <!-- EQ 状态灯：状态层，绝不 aria-hidden，仅正在试听行渲染 -->
      <EqLamp v-if="isCurrent" class="row-eq" :active="playing" :label="eqLabel" />
      <!-- 当前试听行读屏提示固定用「正在试听」 -->
      <span v-if="isCurrent" class="sr-only">{{ A11Y.playingRow(track.name) }}</span>
    </div>

    <!-- 3. meta：歌名 + 歌手（手机并入时长） -->
    <div class="meta">
      <span class="name">{{ track.name }}</span>
      <span class="author-line">
        <span class="author-desktop">{{ track.author }}</span>
        <span class="author-mobile">{{ mobileAuthorLine }}</span>
        <span class="source-inline">{{ track.source }}</span>
      </span>
    </div>

    <!-- 4. duration -->
    <div class="duration tabular">{{ formatTime(track.duration) }}</div>

    <!-- 5. source：手机第二行左侧徽标（左对齐、弱化） -->
    <div class="source"><span class="source-tag">{{ track.source }}</span></div>

    <!-- 6. add：手机第二行右侧定宽按钮 -->
    <div class="add">
      <AddButton :state="addState" :name="track.name" @add="onAdd" @retry="onRetry" />
    </div>
  </li>
</template>

<style scoped>
.row {
  position: relative;
  display: grid;
  grid-template-areas: var(--row-areas);
  grid-template-columns: var(--row-cols);
  grid-template-rows: var(--row-rows);
  align-items: center;
  column-gap: 16px;
  min-height: var(--row-h);
  padding: 8px var(--gutter);
  border-bottom: 1px solid var(--border-subtle); /* 分隔线行，不是卡片墙 */
  background: transparent;
  transition: background var(--dur-fast) var(--ease-out);
}
.row:hover { background: var(--surface-hover); }
/* ★ 正在试听行底色用 --surface-active（实色，不用信号软底）；
   左侧 2px signal 竖条用 inset box-shadow，不新增 DOM、不撑破布局。 */
.row.is-current {
  background: var(--surface-active);
  box-shadow: inset 2px 0 0 var(--signal);
}

.cover {
  grid-area: cover;
  position: relative;
  width: var(--cover-size);
  height: var(--cover-size);
  border-radius: var(--r-xs);
  overflow: hidden;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, .07);
}
.cover-rings {
  position: absolute;
  inset: 0;
  background:
    repeating-radial-gradient(circle at 68% 34%, rgba(255, 255, 255, .10) 0 1px, transparent 1px 5px),
    linear-gradient(120deg, rgba(255, 255, 255, .10), transparent 60%);
}

.audition { grid-area: audition; display: flex; align-items: center; gap: 8px; }
.aud-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
  padding: 0 8px;
  border-radius: var(--r-sm);
  color: var(--text-secondary);
}
.aud-btn:hover { background: var(--surface-active); color: var(--signal); }
.row.is-current .aud-btn { color: var(--signal); }
.wave-desktop { display: var(--wave-desktop-display); }
.aud-icon {
  display: var(--aud-icon-display);
  width: 18px; height: 18px;
}
.row-eq { display: var(--row-eq-display); }
.aud-text {
  display: var(--aud-text-display);
  font-size: var(--fs-micro);
  color: var(--text-tertiary);
  white-space: nowrap;
}

.meta { grid-area: meta; display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.name {
  font-size: var(--fs-label);
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.author-line {
  font-size: var(--fs-meta);
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.author-desktop { display: var(--author-desktop-display); }
.author-mobile { display: var(--author-mobile-display); }
.source-inline {
  display: var(--source-inline-display);
  margin-left: 8px;
  font-size: var(--fs-micro);
  color: var(--text-tertiary); /* 信息性文字用实色 token，不用 --text-disabled */
}

.duration {
  grid-area: duration;
  display: var(--dur-cell-display);
  justify-content: flex-end;
  font-size: var(--fs-meta);
  color: var(--text-secondary);
}
.source {
  grid-area: source;
  display: var(--source-cell-display);
  justify-content: var(--source-align);
  align-items: center;
  min-width: 0;
}
.source-tag {
  max-width: 100%;
  padding: 2px 8px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--r-xs);
  font-size: var(--fs-micro);
  color: var(--text-tertiary); /* 弱化靠字号与 tertiary，不用 disabled 色 */
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.add {
  grid-area: add;
  display: flex;
  justify-content: flex-end;
  justify-self: var(--add-justify-self);
  align-self: var(--add-align-self);
}
.add :deep(.add-wrap) { align-items: flex-end; }

/* ★ 唯一允许的组件级断点媒体查询：只重定义 grid-template-areas。
   第一行 cover/meta/meta/audition，第二行 cover/source/source/add；
   来源徽标左置、添加按钮右对齐定宽，不通栏、不铺软底。 */
@media (max-width: 640px) {
  .row {
    grid-template-areas:
      "cover meta meta audition"
      "cover source source add";
    column-gap: 8px;
    row-gap: 4px;
  }
}
</style>
