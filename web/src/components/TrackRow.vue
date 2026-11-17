<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Track } from '../api/types'
import type { RowModel } from '../utils/group'
import { sourceColor } from '../utils/group'
import { coverGradient, formatTime } from '../utils/format'
import { A11Y, COPY, qualityLabel } from '../copy/messages'
import { usePlayer, toggleAudition } from '../composables/usePlayer'
import { getAddState, requestAdd } from '../composables/useAdd'
import { useSession } from '../composables/useSession'

const props = defineProps<{ row: RowModel }>()
const open = ref(props.row.variants.length > 1)
const { state: player } = usePlayer()
const authed = computed(() => useSession().state.status === 'authenticated')

const merged = computed(() => props.row.variants.length > 1)
const currentId = computed(() => player.track?.sourceId ?? '')

function playing(track: Track): boolean {
  return currentId.value === track.sourceId && player.status !== 'empty'
}
function liveText(track: Track): string {
  if (currentId.value !== track.sourceId) return ''
  if (player.status === 'buffering') return COPY.statusBuffering
  if (player.status === 'playing') return COPY.auditioning
  if (player.status === 'error') return COPY.statusError
  return ''
}
function mark(name: string): string {
  return Array.from(name.trim())[0] || '♪'
}
function tone(code: string): string {
  if (code === 'hi') return 'q-hi'
  if (code === 'hr') return 'q-hr'
  return ''
}
async function onAdd(track: Track): Promise<void> {
  await requestAdd(track)
}
</script>

<template>
  <div class="block">
    <article class="row" :class="{ 'is-playing': playing(row.primary) && !merged, 'is-multi': merged }">
      <div class="cover" :style="{ background: coverGradient(row.primary.name) }" aria-hidden="true">
        {{ mark(row.primary.name) }}
        <span v-if="merged" class="badge">{{ row.variants.length }}</span>
      </div>

      <div class="main">
        <div class="title" :title="row.primary.name">
          <button
            v-if="!merged"
            type="button"
            class="name"
            :aria-label="playing(row.primary) ? A11Y.pauseAudition(row.primary.name) : A11Y.audition(row.primary.name)"
            @click="toggleAudition(row.primary)"
          >
            {{ row.primary.name }}
          </button>
          <span v-else class="name">{{ row.primary.name }}</span>
          <span v-if="merged" class="tag">{{ COPY.mergedTag }}</span>
          <span v-else class="q desk-q" :class="tone(row.primary.quality)">{{ qualityLabel(row.primary.quality) }}</span>
          <span v-if="!merged && liveText(row.primary)" class="live desk-live">{{ liveText(row.primary) }}</span>
        </div>
        <div class="sub">
          <span v-if="!merged" class="q phone-q" :class="tone(row.primary.quality)">{{ qualityLabel(row.primary.quality) }}</span>
          <span v-if="!merged && liveText(row.primary)" class="live phone-live">{{ liveText(row.primary) }}</span>
          <span class="author">{{ row.primary.author }}</span>
          <span v-if="merged" class="quiet">{{ row.variants.length }} 个来源</span>
          <span v-else class="phone-dur tabular">{{ formatTime(row.primary.duration) }}</span>
        </div>
      </div>

      <div class="dur tabular">{{ formatTime(row.primary.duration) }}</div>

      <div class="src">
        <i class="dot" :class="{ mix: merged }" :style="merged ? undefined : { background: sourceColor(row.primary.source) }" />
        <span>{{ merged ? `${row.variants.length} 源` : row.primary.source }}</span>
      </div>

      <div class="act">
        <button v-if="merged" type="button" class="add pick" @click="open = !open">
          {{ open ? COPY.groupFold : COPY.pickSource }}
        </button>
        <button
          v-else-if="authed"
          type="button"
          class="add reveal"
          :class="`is-${getAddState(row.primary.sourceId).phase}`"
          :disabled="getAddState(row.primary.sourceId).phase === 'pending' || getAddState(row.primary.sourceId).phase === 'success' || getAddState(row.primary.sourceId).phase === 'duplicate'"
          :aria-label="getAddState(row.primary.sourceId).phase === 'error' ? A11Y.retryAdd(row.primary.name) : A11Y.addTrack(row.primary.name)"
          @click="onAdd(row.primary)"
        >
          <i v-if="getAddState(row.primary.sourceId).phase === 'pending'" class="spin" />
          <template v-else-if="getAddState(row.primary.sourceId).phase === 'success'">{{ COPY.addSuccess }}</template>
          <template v-else-if="getAddState(row.primary.sourceId).phase === 'duplicate'">{{ COPY.addDuplicate }}</template>
          <template v-else-if="getAddState(row.primary.sourceId).phase === 'error'">{{ COPY.addError }}</template>
          <template v-else>{{ COPY.addIdle }}</template>
        </button>
      </div>

      <div class="acts">
        <button
          v-if="authed && !merged"
          type="button"
          class="round add-round reveal"
          :class="`is-${getAddState(row.primary.sourceId).phase}`"
          :aria-label="A11Y.addTrack(row.primary.name)"
          :disabled="getAddState(row.primary.sourceId).phase === 'pending' || getAddState(row.primary.sourceId).phase === 'success' || getAddState(row.primary.sourceId).phase === 'duplicate'"
          @click="onAdd(row.primary)"
        >
          <i v-if="getAddState(row.primary.sourceId).phase === 'pending'" class="spin" />
          <svg v-else-if="getAddState(row.primary.sourceId).phase === 'success'" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.2 4.2L19 7" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" /></svg>
          <svg v-else viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" fill="none" stroke="currentColor" stroke-width="2.4" /></svg>
        </button>
        <button v-if="merged" type="button" class="round toggle" :aria-label="open ? COPY.groupFold : COPY.pickSource" @click="open = !open">
          <span class="n">{{ row.variants.length }}</span>
        </button>
      </div>
    </article>

    <div v-if="merged" class="fold" :class="{ open }">
      <div class="subs">
        <div v-for="track in row.variants" :key="track.sourceId" class="variant" :class="{ 'is-on': playing(track) }">
          <button
            type="button"
            class="pf"
            :aria-label="playing(track) ? A11Y.pauseAudition(track.name) : A11Y.audition(track.name)"
            @click="toggleAudition(track)"
          >
            <i class="dot" :style="{ background: sourceColor(track.source) }" />
            {{ track.source }}
            <span class="q" :class="tone(track.quality)">{{ qualityLabel(track.quality) }}</span>
            <span v-if="liveText(track)" class="live">{{ liveText(track) }}</span>
          </button>
          <span class="tabular quiet">{{ formatTime(track.duration) }}</span>
          <div class="tail">
            <button
              v-if="authed"
              type="button"
              class="mini addlet reveal"
              :class="`is-${getAddState(track.sourceId).phase}`"
              :disabled="getAddState(track.sourceId).phase === 'pending' || getAddState(track.sourceId).phase === 'success' || getAddState(track.sourceId).phase === 'duplicate'"
              @click="onAdd(track)"
            >
              {{ getAddState(track.sourceId).phase === 'success' ? COPY.addSuccess : getAddState(track.sourceId).phase === 'duplicate' ? COPY.addDuplicate : getAddState(track.sourceId).phase === 'error' ? COPY.addError : COPY.addIdle }}
            </button>
          </div>
        </div>
      </div>
    </div>
    <p v-if="authed && !merged && getAddState(row.primary.sourceId).phase === 'error'" class="reason">
      {{ getAddState(row.primary.sourceId).reason }}
    </p>
  </div>
</template>

<style scoped>
.block { margin-bottom: 2px; }
.row {
  display: grid;
  grid-template-columns: 46px minmax(0, 1fr) 52px 88px auto;
  align-items: center;
  gap: 14px;
  padding: 8px 10px;
  border-radius: var(--r-m);
  border: 1px solid transparent;
}
.row:hover,
.row.is-playing,
.row.is-multi:hover { background: var(--surface); border-color: var(--line); }
.row.is-playing .name { color: var(--acc); }
.row.is-multi {
  background: linear-gradient(90deg, rgba(155, 140, 255, .07), transparent 46%);
  border-color: rgba(155, 140, 255, .2);
}
.play, .round {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: var(--t3);
  border: 1px solid var(--line-2);
  background: var(--bg-1);
}
.play svg, .round svg { width: 12px; height: 12px; }
.row.is-playing .play,
.round.is-on { background: var(--acc); border-color: var(--acc); color: var(--on-acc); }
.cover {
  width: 46px;
  height: 46px;
  border-radius: var(--r-s);
  display: grid;
  place-items: center;
  position: relative;
  font-family: var(--serif);
  color: rgba(255, 255, 255, .78);
  font-size: 16px;
}
.badge {
  position: absolute;
  right: -4px;
  bottom: -4px;
  min-width: 16px;
  height: 16px;
  padding: 0 3px;
  border-radius: 8px;
  background: var(--acc-2);
  color: #0B0716;
  font-family: var(--mono);
  font-size: 9px;
  display: grid;
  place-items: center;
  border: 2px solid var(--bg);
}
.title { display: flex; align-items: center; gap: 8px; min-width: 0; }
.name {
  font-size: 13.5px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  padding: 0;
  text-align: left;
}
button.name { cursor: pointer; }
button.name:hover { color: var(--acc); }
.tag {
  flex: none;
  font-size: 10px;
  color: var(--acc-2);
  border: 1px solid rgba(155, 140, 255, .3);
  background: rgba(155, 140, 255, .08);
  padding: 1px 6px;
  border-radius: 3px;
}
.q {
  flex: none;
  font-family: var(--mono);
  font-size: 9.5px;
  letter-spacing: .04em;
  line-height: 1.5;
  padding: 1px 5px;
  border-radius: 3px;
  border: 1px solid var(--line-2);
  color: var(--t3);
}
.q-hi {
  color: var(--acc);
  border-color: color-mix(in srgb, var(--acc) 35%, transparent);
  background: color-mix(in srgb, var(--acc) 8%, transparent);
}
.q-hr {
  color: var(--acc-2);
  border-color: color-mix(in srgb, var(--acc-2) 35%, transparent);
  background: color-mix(in srgb, var(--acc-2) 8%, transparent);
}
.phone-q, .phone-dur, .phone-live { display: none; }
.sub {
  display: flex;
  gap: 8px;
  margin-top: 2px;
  font-size: 11.5px;
  color: var(--t3);
  min-width: 0;
}
.author, .quiet { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dur { font-size: 11px; color: var(--t4); text-align: right; }
.src { display: inline-flex; align-items: center; gap: 6px; font-size: 11px; color: var(--t3); min-width: 0; }
.src span { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dot { width: 8px; height: 8px; border-radius: 50%; background: var(--acc); flex: none; }
.dot.mix { background: linear-gradient(135deg, #E5344B, #2E9BFF 50%, #31C27C); }
.add {
  height: 30px;
  padding: 0 12px;
  border-radius: var(--r-s);
  border: 1px solid var(--line-2);
  background: var(--bg-1);
  font-size: 12px;
  color: var(--t2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  min-width: 72px;
}
.add:hover:not(:disabled) { color: var(--on-acc); background: var(--acc); border-color: var(--acc); }
.add.pick { color: var(--acc-2); border-color: rgba(155, 140, 255, .32); }
.add.is-success { color: var(--ok); border-color: rgba(95, 214, 164, .3); background: var(--ok-soft); }
.add.is-duplicate { color: var(--warn); border-color: rgba(232, 180, 95, .35); background: var(--warn-soft); }
.add.is-pending { color: var(--t3); cursor: progress; }
.spin {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  border: 1.5px solid var(--line-3);
  border-top-color: var(--acc);
  animation: sp .7s linear infinite;
}
@keyframes sp { to { transform: rotate(360deg); } }
.act, .tail, .acts { display: flex; align-items: center; justify-content: flex-end; gap: 8px; }
.acts { display: none; }
.live {
  flex: none;
  height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--acc) 40%, transparent);
  background: var(--acc-soft);
  color: var(--acc);
  font-size: 11px;
  display: inline-flex;
  align-items: center;
}
.live.phone-live { display: none; }
.reveal {
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
}
.row:hover > .act .reveal,
.row:hover > .acts .reveal,
.row:focus-within > .act .reveal,
.row:focus-within > .acts .reveal,
.variant:hover .reveal,
.variant:focus-within .reveal {
  opacity: 1;
  visibility: visible;
  pointer-events: auto;
}
.fold {
  display: grid;
  grid-template-rows: 0fr;
  margin-left: 60px;
  transition: grid-template-rows .28s var(--ease);
}
.fold.open { grid-template-rows: 1fr; margin-bottom: 8px; }
.subs {
  min-height: 0;
  overflow: hidden;
  border-left: 1px solid rgba(155, 140, 255, .18);
  padding-left: 12px;
}
.variant {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 52px auto;
  align-items: center;
  gap: 10px;
  padding: 6px 8px;
  border-radius: var(--r-s);
  border: 1px solid transparent;
  font-size: 12px;
  color: var(--t2);
}
.variant:hover,
.variant.is-on { background: var(--surface); border-color: var(--line); }
.pf {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  padding: 0;
  color: inherit;
  font: inherit;
  text-align: left;
  cursor: pointer;
}
.pf:hover,
.variant.is-on .pf { color: var(--acc); }
.mini {
  height: 26px;
  border-radius: 5px;
  border: 1px solid var(--line-2);
  background: var(--bg-1);
  font-size: 11px;
  color: var(--t2);
}
.mini.is-success { color: var(--ok); background: var(--ok-soft); border-color: rgba(95, 214, 164, .28); }
.mini.is-duplicate { color: var(--warn); }
.reason { margin: 0 10px 6px 90px; font-size: 11px; color: var(--err); }
@media (hover: none), (max-width: 767px) {
  .reveal { opacity: 1; visibility: visible; pointer-events: auto; }
}
@media (max-width: 1023px) {
  .row { grid-template-columns: 42px minmax(0, 1fr) 88px auto; }
  .dur { display: none; }
  .cover { width: 42px; height: 42px; }
}
@media (max-width: 767px) {
  .row {
    grid-template-columns: 44px minmax(0, 1fr) auto;
    gap: 10px;
  }
  .play, .wave, .dur, .src, .act, .desk-q { display: none; }
  .phone-q, .live.phone-live { display: inline-flex; }
  .live.desk-live { display: none; }
  .phone-dur {
    display: inline;
    margin-left: auto;
    font-size: 9.5px;
    color: var(--t4);
  }
  .sub { font-size: 10.5px; gap: 6px; align-items: center; }
  .acts { display: flex; gap: 6px; }
  .cover { width: 44px; height: 44px; }
  .round {
    width: 32px;
    height: 32px;
    position: relative;
  }
  .round::before {
    content: "";
    position: absolute;
    inset: -6px;
  }
  .add-round { color: var(--acc); background: color-mix(in srgb, var(--acc) 12%, transparent); border-color: color-mix(in srgb, var(--acc) 28%, transparent); }
  .add-round.is-success { color: var(--ok); background: var(--ok-soft); }
  .fold { margin-left: 12px; }
  .subs { padding-left: 8px; }
  .variant {
    grid-template-columns: minmax(0, 1fr) auto;
  }
  .tail { grid-column: 1 / -1; }
}
</style>
