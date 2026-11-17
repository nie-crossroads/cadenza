<script setup lang="ts">
import { computed, ref } from 'vue'
import { A11Y, COPY, qualityLabel } from '../copy/messages'
import { coverGradient, formatTime, toPercent } from '../utils/format'
import { sourceColor } from '../utils/group'
import { usePlayer, playTrack, togglePlay, seek, setVolume, retryPreview } from '../composables/usePlayer'
import { getAddState, add as addTrack, retry } from '../composables/useAdd'
import { confirmAdd } from '../composables/useConfirm'
import { useSession } from '../composables/useSession'
import { useSearch } from '../composables/useSearch'
import { useTheme } from '../composables/useTheme'
import Waveform from './Waveform.vue'
import Scrubber from './Scrubber.vue'

const { state } = usePlayer()
const { state: search } = useSearch()
const { state: theme } = useTheme()
const authed = computed(() => useSession().state.status === 'authenticated')
const sheet = ref(false)
const hasTrack = computed(() => state.track !== null)
const playing = computed(() => state.status === 'playing' || state.status === 'buffering')
const ratio = computed(() => (state.duration > 0 ? state.currentTime / state.duration : 0))
const queueIndex = computed(() => search.items.findIndex((item) => item.sourceId === state.track?.sourceId))
const canPrev = computed(() => queueIndex.value > 0)
const canNext = computed(() => queueIndex.value >= 0 && queueIndex.value < search.items.length - 1)
const siblings = computed(() => {
  const track = state.track
  if (!track) return []
  const key = `${track.name.trim().toLocaleLowerCase()}\0${track.author.trim().toLocaleLowerCase()}`
  const found = search.items.filter((item) => `${item.name.trim().toLocaleLowerCase()}\0${item.author.trim().toLocaleLowerCase()}` === key)
  return found.length > 0 ? found : [track]
})

function step(dir: -1 | 1): void {
  const next = search.items[queueIndex.value + dir]
  if (next) playTrack(next)
}
function addLabel(): string {
  const track = state.track
  if (!track) return COPY.saveToBlog
  const phase = getAddState(track.sourceId).phase
  if (phase === 'success') return COPY.addSuccess
  if (phase === 'duplicate') return COPY.addDuplicate
  if (phase === 'pending') return COPY.addPending
  if (phase === 'error') return COPY.addError
  return COPY.saveToBlog
}

function onSeek(next: number): void {
  if (!state.rangeSupported) return
  seek(next * (state.duration || 0))
}
async function onAdd(): Promise<void> {
  const track = state.track
  if (!track) return
  const phase = getAddState(track.sourceId).phase
  if (phase === 'pending' || phase === 'success' || phase === 'duplicate') return
  const ok = await confirmAdd(track.name)
  if (!ok) return
  if (phase === 'error') retry(track)
  else void addTrack(track)
}
function mark(name: string): string {
  return Array.from(name.trim())[0] || '♪'
}
</script>

<template>
  <footer class="dock">
    <div class="desk">
      <div class="lead">
        <div class="cover" :class="{ pic: hasTrack, lite: hasTrack && !theme.dark }" :style="hasTrack ? { background: coverGradient(state.track!.name, !theme.dark) } : undefined" aria-hidden="true">
          {{ hasTrack ? mark(state.track!.name) : '𝄞' }}
        </div>
        <div class="meta">
          <div class="title">{{ hasTrack ? state.track!.name : COPY.previewEmpty }}</div>
          <div v-if="hasTrack" class="sub">
            <i class="dot" :style="{ background: sourceColor(state.track!.source) }" />
            <span>{{ state.track!.author }}</span>
            <span v-if="state.track!.source">· {{ state.track!.source }}</span>
            <span>· {{ qualityLabel(state.track!.quality) }}</span>
          </div>
          <p v-else-if="!state.rangeSupported" class="sub">{{ COPY.rangeNotice }}</p>
        </div>
        <div class="scrub">
          <Waveform
            v-if="hasTrack"
            :seed="state.track!.name"
            :progress="ratio"
            :active="state.status === 'playing'"
            @seek="onSeek"
          />
          <div v-else class="wave-empty" />
        </div>
      </div>
      <div class="transport">
        <button type="button" class="play" :disabled="!hasTrack" :aria-label="playing ? A11Y.pause : A11Y.play" @click="togglePlay">
          <svg v-if="playing" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 5h4v14H7zM13 5h4v14h-4z" fill="currentColor" /></svg>
          <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z" fill="currentColor" /></svg>
        </button>
      </div>
      <div class="side">
        <div class="vol-stack">
          <svg class="speaker" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M4 9.5v5h3.5L12 19V5L7.5 9.5H4z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" />
            <path v-if="!state.muted && state.volume > 0" d="M16 9.2a4 4 0 0 1 0 5.6M18.4 7a7 7 0 0 1 0 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
          </svg>
          <label class="vol">
            <span class="sr-only">{{ A11Y.volume }}</span>
            <input
              type="range"
              min="0"
              max="100"
              :value="toPercent(state.muted ? 0 : state.volume)"
              :aria-valuetext="A11Y.volumeValue(toPercent(state.volume))"
              @input="setVolume(Number(($event.target as HTMLInputElement).value) / 100)"
            />
          </label>
          <span class="time tabular">{{ formatTime(state.currentTime) }} / {{ formatTime(state.duration) }}</span>
        </div>
        <button
          v-if="authed && hasTrack"
          type="button"
          class="primary"
          :disabled="getAddState(state.track!.sourceId).phase === 'pending' || getAddState(state.track!.sourceId).phase === 'success' || getAddState(state.track!.sourceId).phase === 'duplicate'"
          @click="onAdd"
        >
          {{ getAddState(state.track!.sourceId).phase === 'success' ? COPY.addSuccess : COPY.addToBlog }}
        </button>
        <button v-if="state.status === 'error'" type="button" class="primary" @click="retryPreview">{{ COPY.retry }}</button>
      </div>
      <Scrubber v-if="hasTrack" class="edge" :ratio="ratio" :disabled="!state.rangeSupported" @seek="onSeek" />
    </div>

    <div v-if="hasTrack" class="mini" @click="sheet = true">
      <span class="mini-lead">
        <span class="cover sm" :class="{ pic: true, lite: !theme.dark }" :style="{ background: coverGradient(state.track!.name, !theme.dark) }" aria-hidden="true">{{ mark(state.track!.name) }}</span>
        <span class="mini-meta">
          <span class="title">{{ state.track!.name }}</span>
          <span class="sub">{{ state.track!.author }}<template v-if="state.track!.source"> · {{ state.track!.source }}</template></span>
        </span>
      </span>
      <button type="button" class="mini-play" :aria-label="playing ? A11Y.pause : A11Y.play" @click.stop="togglePlay">
        <svg v-if="playing" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 5h4v14H7zM13 5h4v14h-4z" fill="currentColor" /></svg>
        <svg v-else class="nudge" viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z" fill="currentColor" /></svg>
      </button>
      <span class="mini-side">
        <button type="button" class="mini-open" :aria-label="A11Y.openPlayer" @click.stop="sheet = true">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 15l6-6 6 6" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" /></svg>
        </button>
      </span>
      <Scrubber class="edge" :ratio="ratio" :disabled="!state.rangeSupported" @seek="onSeek" />
    </div>
    <div v-else class="mini mini-empty">{{ COPY.previewEmpty }}</div>

    <Transition name="sheet">
    <div v-if="sheet && hasTrack" class="sheet">
      <header class="sheet-h">
        <button type="button" class="back" :aria-label="COPY.closePlayer" @click="sheet = false">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M14.5 6.5L8 12l6.5 5.5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg>
        </button>
      </header>
      <div class="sheet-stage">
        <div class="art-slot">
          <div class="art" :class="{ lite: !theme.dark }" :style="{ background: coverGradient(state.track!.name, !theme.dark) }">
            <span>{{ mark(state.track!.name) }}</span>
          </div>
        </div>
        <div class="sheet-info">
          <div class="sheet-title">{{ state.track!.name }}</div>
          <div class="who">
            <span>{{ state.track!.author }}</span>
            <span class="src-chip">
              <i class="dot" :style="{ background: sourceColor(state.track!.source) }" />
              {{ state.track!.source }} · {{ qualityLabel(state.track!.quality) }}
            </span>
          </div>
        </div>
        <Scrubber class="sheet-scrub" :ratio="ratio" :disabled="!state.rangeSupported" @seek="onSeek" />
        <div class="times tabular"><span>{{ formatTime(state.currentTime) }}</span><span>{{ formatTime(state.duration) }}</span></div>
      </div>
      <div class="sheet-lower">
        <div class="transport-row">
          <button type="button" class="step" :disabled="!canPrev" :aria-label="A11Y.prev" @click="step(-1)">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6h2v12H6zM20 6v12l-9-6z" fill="currentColor" /></svg>
          </button>
          <button type="button" class="big" :aria-label="playing ? A11Y.pause : A11Y.play" @click="togglePlay">
            <svg v-if="playing" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 5h4v14H7zM13 5h4v14h-4z" fill="currentColor" /></svg>
            <svg v-else class="nudge" viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z" fill="currentColor" /></svg>
          </button>
          <button type="button" class="step" :disabled="!canNext" :aria-label="A11Y.next" @click="step(1)">
            <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M16 6h2v12h-2zM4 6v12l9-6z" fill="currentColor" /></svg>
          </button>
        </div>
        <section v-if="siblings.length > 1" class="sources">
          <div class="lbl">{{ COPY.sourcesHere(siblings.length) }}</div>
          <div class="src-row">
            <button
              v-for="item in siblings"
              :key="item.sourceId"
              type="button"
              class="src-pick"
              :class="{ on: item.sourceId === state.track!.sourceId }"
              :aria-label="A11Y.switchSource(item.source)"
              @click="playTrack(item)"
            >
              <span><i class="dot" :style="{ background: sourceColor(item.source) }" />{{ item.source }}</span>
              <em>{{ qualityLabel(item.quality) }}</em>
            </button>
          </div>
        </section>
        <button
          v-if="authed"
          type="button"
          class="cta"
          :disabled="getAddState(state.track!.sourceId).phase === 'pending' || getAddState(state.track!.sourceId).phase === 'success' || getAddState(state.track!.sourceId).phase === 'duplicate'"
          @click="onAdd"
        >
          <svg v-if="getAddState(state.track!.sourceId).phase === 'success'" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.2 4.2L19 7" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" /></svg>
          <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" /></svg>
          {{ addLabel() }}
        </button>
        <p class="note">{{ COPY.proxyNote }}</p>
      </div>
    </div>
    </Transition>
  </footer>
</template>

<style scoped>
.dock { flex: none; background: transparent; }
.desk {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 16px;
  width: var(--lane);
  margin: 18px auto 16px;
  padding: 12px 20px;
  min-height: 76px;
  border: 1px solid var(--acc);
  border-radius: 18px;
  box-shadow: 0 0 0 4px var(--acc-soft);
  background: linear-gradient(180deg, var(--surface-2), var(--bg-1));
  position: relative;
}
.lead { display: flex; align-items: center; gap: 14px; min-width: 0; }
.cover {
  width: 52px; height: 52px; border-radius: var(--r-s);
  display: grid; place-items: center; flex: none;
  font-family: var(--serif); font-size: 22px;
  color: var(--t1);
  background: color-mix(in srgb, var(--surface-3) 72%, var(--acc));
  border: 1px solid var(--line-2);
}
.cover.pic {
  color: rgba(255, 255, 255, .86);
  border-color: color-mix(in srgb, var(--t1) 22%, transparent);
}
.cover.pic.lite { color: rgba(16, 24, 36, .84); }
.cover.sm { width: 38px; height: 38px; font-size: 13px; }
.meta { min-width: 0; flex: 0 1 220px; }
.title { font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sub { display: flex; align-items: center; gap: 6px; margin-top: 2px; font-size: 11px; color: var(--t3); min-width: 0; }
.sub span { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dot { width: 8px; height: 8px; border-radius: 50%; flex: none; }
.play, .big {
  width: 38px; height: 38px; border-radius: 50%;
  display: grid; place-items: center;
  background: var(--acc); color: var(--on-acc);
}
.play:disabled { opacity: .4; }
.play svg, .big svg { width: 14px; height: 14px; }
.transport { justify-self: center; }
.scrub { flex: 1; min-width: 0; display: flex; align-items: center; }
.wave-empty { flex: 1; height: 22px; }
.side { justify-self: end; display: flex; align-items: center; gap: 12px; }
.vol-stack {
  display: grid;
  grid-template-columns: 16px 92px;
  column-gap: 8px;
  row-gap: 4px;
  align-items: center;
}
.speaker { width: 16px; height: 16px; color: var(--t2); grid-column: 1; grid-row: 1; }
.vol { grid-column: 2; grid-row: 1; display: flex; }
.vol input {
  -webkit-appearance: none;
  appearance: none;
  width: 92px;
  height: 12px;
  margin: 0;
  background: transparent;
  accent-color: var(--acc);
}
.vol input::-webkit-slider-runnable-track {
  height: 3px;
  border-radius: 999px;
  background: var(--line-2);
}
.vol input::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 10px;
  height: 10px;
  margin-top: -3.5px;
  border: 0;
  border-radius: 50%;
  background: var(--acc);
}
.vol input::-moz-range-track {
  height: 3px;
  border: 0;
  border-radius: 999px;
  background: var(--line-2);
}
.vol input::-moz-range-thumb {
  width: 10px;
  height: 10px;
  border: 0;
  border-radius: 50%;
  background: var(--acc);
}
.time {
  grid-column: 2;
  grid-row: 2;
  font-size: 10px;
  color: var(--t4);
  text-align: center;
  white-space: nowrap;
}
.primary {
  height: 34px; padding: 0 14px; border-radius: var(--r-s);
  background: var(--acc); color: var(--on-acc); font-size: 12px; font-weight: 600;
}
.primary:disabled { opacity: .55; }
.primary.wide { width: calc(100% - 48px); margin: 16px 24px 0; height: 44px; }
.edge {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: -9px;
  z-index: 2;
}
.mini, .mini-empty { display: none; }
.sheet { display: none; }
@media (max-width: 1023px) {
  .desk { padding: 10px 14px; }
  .meta { flex-basis: 140px; }
  .vol-stack { grid-template-columns: 16px 72px; }
  .vol input { width: 72px; }
}
@media (max-width: 767px) {
  .desk { display: none; }
  .mini, .mini-empty {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 36px minmax(0, 1fr);
    align-items: center;
    gap: 8px;
    position: relative;
    margin: 10px auto calc(10px + var(--safe-bottom));
    height: 58px;
    padding: 0 10px;
    border-radius: 14px;
    border: 1px solid var(--line-2);
    box-shadow: 0 -8px 30px -14px rgba(0, 0, 0, .9);
    background: linear-gradient(180deg, var(--surface-2), var(--bg-1));
    text-align: left;
    overflow: visible;
    width: var(--lane);
    line-height: 1;
  }
  .mini-empty { display: flex; align-items: center; color: var(--t4); font-size: 12px; }
  .mini-lead {
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .cover.sm { align-self: center; margin: 0; }
  .mini-meta { min-width: 0; display: flex; flex-direction: column; justify-content: center; }
  .mini .title { font-size: 12px; }
  .mini .sub { font-size: 10px; margin-top: 6px; }
  .mini-side { display: flex; justify-content: flex-end; }
  .mini-play, .mini-open {
    position: relative;
    padding: 0;
    border-radius: 50%;
    display: grid;
    place-items: center;
    justify-self: center;
    align-self: center;
  }
  .mini-play {
    width: 36px;
    height: 36px;
    background: var(--acc);
    color: var(--on-acc);
    border: 0;
    box-shadow: none;
  }
  .mini-open {
    width: 30px;
    height: 30px;
    color: var(--t3);
    border: 1px solid var(--line);
    background: transparent;
  }
  .mini-play svg, .mini-open svg {
    position: absolute;
    inset: 0;
    margin: auto;
    display: block;
  }
  .mini-play svg { width: 12px; height: 12px; }
  .mini-play svg.nudge { transform: translateX(1px); }
  .mini-open svg { width: 13px; height: 13px; }
  .sheet-enter-active,
  .sheet-leave-active { transition: transform .34s var(--ease); }
  .sheet-enter-from,
  .sheet-leave-to { transform: translateX(100%); }
  .sheet {
    display: flex;
    flex-direction: column;
    position: fixed;
    top: var(--header-h);
    right: 0;
    bottom: 0;
    left: 0;
    z-index: 4;
    overflow: hidden;
    background:
      radial-gradient(ellipse 70% 42% at 50% 0%, color-mix(in srgb, var(--acc) 10%, transparent), transparent 68%),
      var(--bg);
    padding-bottom: calc(16px + var(--safe-bottom));
  }
  .sheet-h {
    flex: none;
    height: 56px;
    display: flex;
    align-items: center;
    padding: 0 16px;
  }
  .back {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    border: 1px solid var(--line-2);
    background: var(--surface);
    color: var(--t1);
    display: grid;
    place-items: center;
    padding: 0;
  }
  .back svg { width: 18px; height: 18px; }
  .sheet-stage {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }
  .art-slot {
    flex: 1;
    min-height: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 6px 30px 0;
  }
  .art {
    width: min(100%, 320px);
    max-height: 100%;
    aspect-ratio: 1;
    border-radius: 18px;
    position: relative;
    overflow: hidden;
    display: grid;
    place-items: center;
    box-shadow: 0 26px 50px -22px rgba(0, 0, 0, .85);
  }
  .art::after {
    content: "";
    position: absolute;
    inset: 0;
    background: repeating-radial-gradient(circle at 50% 50%, transparent 0 9px, rgba(255, 255, 255, .045) 9px 10px);
  }
  .art span {
    position: relative;
    z-index: 1;
    font-family: var(--serif);
    font-size: 54px;
    color: rgba(255, 255, 255, .75);
  }
  .art.lite span { color: rgba(16, 24, 36, .84); }
  .sheet-info { flex: none; padding: 16px 24px 0; }
  .sheet-title { font-size: 17px; font-weight: 600; letter-spacing: -.01em; line-height: 1.3; }
  .who {
    margin-top: 4px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--t3);
    min-width: 0;
  }
  .who > span:first-child { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .src-chip {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    flex: none;
    font-size: 10px;
    color: var(--t3);
    border: 1px solid var(--line);
    border-radius: 3px;
    padding: 1px 5px;
  }
  .src-chip .dot { width: 6px; height: 6px; }
  .sheet-scrub { flex: none; margin: 14px 24px 0; }
  .times {
    flex: none;
    display: flex;
    justify-content: space-between;
    margin: 6px 24px 0;
    font-size: 10px;
    color: var(--t4);
  }
  .sheet-lower { flex: none; }
  .transport-row {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 26px;
    margin-top: 14px;
  }
  .step {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    color: var(--t3);
    padding: 0;
  }
  .step:disabled { opacity: .35; }
  .step svg { width: 17px; height: 17px; }
  .transport-row .big {
    position: relative;
    width: 58px;
    height: 58px;
    margin: 0;
    box-shadow: 0 10px 26px -8px color-mix(in srgb, var(--acc) 75%, transparent);
  }
  .transport-row .big svg {
    position: absolute;
    inset: 0;
    margin: auto;
    width: 20px;
    height: 20px;
  }
  .transport-row .big svg.nudge { transform: translateX(1.5px); }
  .sources { margin: 16px 24px 0; }
  .lbl {
    font-size: 10px;
    letter-spacing: .14em;
    color: var(--t4);
    margin-bottom: 7px;
  }
  .src-row { display: flex; gap: 6px; }
  .src-pick {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--line);
    border-radius: 8px;
    padding: 6px 8px;
    text-align: center;
    background: var(--bg-1);
    color: var(--t3);
  }
  .src-pick.on {
    border-color: color-mix(in srgb, var(--acc) 40%, transparent);
    background: var(--acc-soft);
    color: var(--t1);
  }
  .src-pick span {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    font-size: 10px;
  }
  .src-pick .dot { width: 6px; height: 6px; }
  .src-pick em {
    display: block;
    margin-top: 2px;
    font-style: normal;
    font-family: var(--mono);
    font-size: 9px;
    color: var(--t4);
  }
  .cta {
    width: calc(100% - 48px);
    height: 44px;
    margin: 14px 24px 0;
    border-radius: var(--r-m);
    background: var(--acc);
    color: var(--on-acc);
    font-size: 13px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
  }
  .cta:disabled { opacity: .55; }
  .cta svg { width: 13px; height: 13px; }
  .note {
    margin: 10px 24px 0;
    font-size: 10px;
    color: var(--t4);
    text-align: center;
    line-height: 1.6;
  }
}
</style>
