<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { A11Y, BRAND, COPY } from '../copy/messages'
import { logout as apiLogout } from '../api'
import { useSearch, setQuery, submitNow, clearSearch, resetSearch } from '../composables/useSearch'
import { markAnonymous, useSession } from '../composables/useSession'
import { stopPlayback } from '../composables/usePlayer'
import { resetAddStates } from '../composables/useAdd'
import { clearToasts } from '../composables/useToast'
import { sourceColor } from '../utils/group'
import SearchPill from '../components/SearchPill.vue'
import ResultList from '../components/ResultList.vue'
import PlayerBar from '../components/PlayerBar.vue'
import ThemeControls from '../components/ThemeControls.vue'

const route = useRoute()
const router = useRouter()
const { state } = useSearch()
const authed = computed(() => useSession().state.status === 'authenticated')
const searchRef = ref<InstanceType<typeof SearchPill> | null>(null)
const source = ref('all')
const picked = ref<string[]>(COPY.sourceNames.slice(0, 3))
const pickerOpen = ref(false)
const draft = ref<string[]>([])
const platsRef = ref<HTMLElement | null>(null)
const phone = ref(false)
const rail = reactive({ overflow: false, thumb: 100, x: 0 })
const allDraft = computed(() => draft.value.length === COPY.sourceNames.length)
const allPicked = computed(() => picked.value.length === COPY.sourceNames.length)
const showRail = computed(() => phone.value && allPicked.value && rail.overflow)

function openPicker(): void {
  draft.value = [...picked.value]
  pickerOpen.value = true
}

function toggleDraft(name: string): void {
  draft.value = draft.value.includes(name)
    ? draft.value.filter((item) => item !== name)
    : [...draft.value, name]
}

function toggleAllDraft(): void {
  draft.value = allDraft.value ? [] : [...COPY.sourceNames]
}

function applySources(): void {
  const chosen = new Set(draft.value)
  picked.value = COPY.sourceNames.filter((name) => chosen.has(name))
  if (source.value !== 'all' && !picked.value.includes(source.value)) source.value = 'all'
  pickerOpen.value = false
}

function measurePlats(): void {
  const el = platsRef.value
  if (!el) {
    rail.overflow = false
    return
  }
  const total = el.scrollWidth
  const view = el.clientWidth
  const overflow = total > view + 1
  const thumb = total > 0 ? (view / total) * 100 : 100
  const max = Math.max(0, 100 - thumb)
  rail.overflow = overflow
  rail.thumb = thumb
  rail.x = total <= view ? 0 : (el.scrollLeft / (total - view)) * max
}

function seekPlats(railEl: HTMLElement, clientX: number): void {
  const el = platsRef.value
  if (!el) return
  const rect = railEl.getBoundingClientRect()
  const view = el.clientWidth
  const total = el.scrollWidth
  const thumbW = rect.width * (view / Math.max(total, 1))
  const travel = Math.max(0, rect.width - thumbW)
  const x = Math.min(travel, Math.max(0, clientX - rect.left - thumbW / 2))
  el.scrollLeft = travel === 0 ? 0 : (x / travel) * (total - view)
}

function onRailDown(event: PointerEvent): void {
  const railEl = event.currentTarget as HTMLElement
  railEl.setPointerCapture(event.pointerId)
  seekPlats(railEl, event.clientX)
}

function onRailMove(event: PointerEvent): void {
  const railEl = event.currentTarget as HTMLElement
  if (!railEl.hasPointerCapture(event.pointerId)) return
  seekPlats(railEl, event.clientX)
}

function searchQuery(q: string): void {
  setQuery(q)
  submitNow()
  const next = q.trim()
  if (next && route.query.q !== next) {
    void router.replace({ name: 'detail', query: { q: next } })
  }
}

function pickRecent(q: string): void {
  searchQuery(q)
}

function onSubmit(): void {
  searchQuery(state.query)
}

function syncRouteQuery(): void {
  const raw = route.query.q
  const q = typeof raw === 'string' ? raw : ''
  if (!q.trim()) return
  if (state.query === q && state.phase !== 'empty') return
  setQuery(q)
  submitNow()
}

async function onLogout(): Promise<void> {
  await apiLogout()
  stopPlayback()
  resetAddStates()
  resetSearch()
  clearToasts()
  markAnonymous()
  await router.push({ name: 'cover' })
}

function onKey(e: KeyboardEvent): void {
  if (pickerOpen.value && e.key === 'Escape') {
    e.preventDefault()
    pickerOpen.value = false
    return
  }
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    searchRef.value?.focus()
  }
}

let phoneMq: MediaQueryList | null = null
function onPhoneChange(): void {
  phone.value = phoneMq?.matches ?? false
  measurePlats()
}

onMounted(() => {
  phoneMq = window.matchMedia('(max-width: 767px)')
  phone.value = phoneMq.matches
  phoneMq.addEventListener('change', onPhoneChange)
  window.addEventListener('keydown', onKey)
  syncRouteQuery()
})
onUnmounted(() => {
  phoneMq?.removeEventListener('change', onPhoneChange)
  window.removeEventListener('keydown', onKey)
})
watch(() => route.query.q, () => syncRouteQuery())
watch(platsRef, () => { void nextTick(measurePlats) })
watch(picked, () => { void nextTick(measurePlats) })
</script>

<template>
  <div class="shell">
    <header class="top">
      <div class="bar">
        <button type="button" class="brand" @click="router.push({ name: 'cover' })">
          <strong>Cadenza</strong>
          <span>{{ BRAND.corner }}</span>
        </button>
        <div class="top-r">
          <template v-if="authed">
            <span class="tag">{{ COPY.loggedIn }}</span>
            <button type="button" class="out" @click="onLogout">{{ COPY.logout }}</button>
          </template>
          <ThemeControls />
        </div>
      </div>
    </header>

    <div class="body">
      <SearchPill
        ref="searchRef"
        variant="hero"
        :model-value="state.query"
        :loading="state.phase === 'loading'"
        @update:model-value="setQuery"
        @submit="onSubmit"
        @clear="clearSearch"
      />

      <div v-if="state.phase === 'results' && state.items.length" class="plats">
        <span class="plabel">{{ COPY.sourceLabel }}</span>
        <div class="chip-col">
          <div ref="platsRef" class="chip-scroll" :class="{ rail: showRail }" @scroll="measurePlats">
            <button type="button" class="pf" :class="{ 'is-on': source === 'all' }" @click="source = 'all'">
              <i class="dot mix" />{{ COPY.allSources }}
            </button>
            <button
              v-for="name in picked"
              :key="name"
              type="button"
              class="pf"
              :class="{ 'is-on': source === name }"
              @click="source = name"
            >
              <i class="dot" :style="{ background: sourceColor(name) }" />
              {{ name }}
            </button>
            <button type="button" class="pf pick" @click="openPicker">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" /></svg>
              {{ COPY.pickSources }}
            </button>
          </div>
          <div
            v-if="showRail"
            class="source-rail"
            @pointerdown="onRailDown"
            @pointermove="onRailMove"
          >
            <i class="thumb" :style="{ width: `${rail.thumb}%`, left: `${rail.x}%` }" />
          </div>
        </div>
      </div>

      <div class="scroll" :aria-label="A11Y.homeMain">
        <ResultList :source="source" @pick="pickRecent" />
      </div>
    </div>

    <PlayerBar />

    <div v-if="pickerOpen" class="mask" @click.self="pickerOpen = false">
      <div class="card" role="dialog" aria-modal="true" :aria-label="COPY.pickSources">
        <h2>{{ COPY.pickSources }}</h2>
        <div class="opts">
          <label class="opt">
            <span>{{ COPY.allSources }}</span>
            <input class="box" type="checkbox" :checked="allDraft" @change="toggleAllDraft" />
          </label>
          <label v-for="name in COPY.sourceNames" :key="name" class="opt">
            <span class="lead">
              <i class="dot" :style="{ background: sourceColor(name) }" />
              <span>{{ name }}</span>
            </span>
            <input class="box" type="checkbox" :checked="draft.includes(name)" @change="toggleDraft(name)" />
          </label>
        </div>
        <div class="dlg-acts">
          <button type="button" class="ghost" @click="pickerOpen = false">{{ COPY.sourcePickCancel }}</button>
          <button type="button" class="ok" @click="applySources">{{ COPY.sourcePickOk }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.shell {
  height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--bg);
}
.top {
  height: var(--header-h);
  flex: none;
  position: relative;
  z-index: 5;
  border-bottom: 1px solid var(--line);
  background: color-mix(in srgb, var(--bg-1) 92%, transparent);
}
.bar {
  width: var(--lane);
  height: 100%;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}
.brand { display: flex; align-items: baseline; gap: 10px; text-align: left; padding: 0; }
.brand strong {
  font-family: var(--serif);
  font-style: italic;
  font-weight: 400;
  font-size: 21px;
}
.brand span {
  font-size: 10px;
  letter-spacing: .24em;
  text-transform: uppercase;
  color: var(--t4);
}
.top-r { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--t3); }
.tag {
  height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--acc) 35%, transparent);
  background: var(--acc-soft);
  color: var(--acc);
  font-size: 12px;
  display: inline-flex;
  align-items: center;
}
.out {
  height: 40px;
  padding: 0 14px;
  border-radius: 12px;
  background: var(--surface);
  color: var(--acc);
  box-shadow: 0 4px 10px -6px rgb(0 0 0 / .45);
  font-size: 13px;
}
.out:hover { background: var(--surface-3); }
.body {
  flex: 1;
  min-height: 0;
  width: var(--lane);
  margin: 0 auto;
  padding: 22px 0 0;
  display: flex;
  flex-direction: column;
}
.plats {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 12px;
  flex: none;
}
.chip-col { flex: 1; min-width: 0; }
.chip-scroll {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow-x: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--acc) transparent;
}
.chip-scroll::-webkit-scrollbar { height: 10px; }
.chip-scroll::-webkit-scrollbar-track { background: transparent; }
.chip-scroll::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--acc) 72%, var(--surface));
  border: 2px solid transparent;
  border-radius: 999px;
  background-clip: padding-box;
}
.chip-scroll::-webkit-scrollbar-thumb:hover {
  background: var(--acc);
  background-clip: padding-box;
  border: 2px solid transparent;
}
.chip-scroll.rail { scrollbar-width: none; }
.chip-scroll.rail::-webkit-scrollbar { display: none; }
.source-rail {
  position: relative;
  height: 10px;
  margin-top: 6px;
  touch-action: none;
  cursor: pointer;
}
.source-rail .thumb {
  position: absolute;
  top: 0;
  height: 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--acc) 72%, var(--surface));
  border: 2px solid transparent;
  background-clip: padding-box;
  box-sizing: border-box;
  pointer-events: none;
}
.plabel {
  display: inline-flex;
  align-items: center;
  height: 30px;
  flex: none;
  font-size: 11px;
  line-height: 1;
  letter-spacing: .14em;
  text-transform: uppercase;
  color: var(--t4);
}
.pf {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 30px;
  padding: 0 12px;
  border: 1px solid var(--line);
  border-radius: var(--r-pill);
  font-size: 12px;
  color: var(--t3);
  background: var(--bg-1);
  flex: none;
}
.pf.is-on { color: var(--t1); border-color: color-mix(in srgb, var(--acc) 45%, transparent); background: var(--acc-soft); }
.pick svg { width: 10px; height: 10px; }
.pick:hover { border-color: var(--acc); color: var(--t1); }
.dot { width: 8px; height: 8px; border-radius: 50%; flex: none; }
.mask {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: grid;
  place-items: center;
  padding: 24px;
  background: color-mix(in srgb, var(--bg) 62%, transparent);
}
.card {
  width: min(360px, 100%);
  padding: 22px 22px 18px;
  border-radius: 16px;
  border: 1px solid var(--line-2);
  background: var(--surface);
  box-shadow: 0 24px 48px -24px rgb(0 0 0 / .65);
}
.card h2 { margin: 0; font-size: 16px; font-weight: 600; text-align: center; }
.opts { display: flex; flex-direction: column; gap: 4px; margin-top: 16px; }
.opt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 36px;
  padding: 0 10px;
  border-radius: 10px;
  font-size: 13px;
  color: var(--t1);
  cursor: pointer;
}
.opt:hover { background: var(--surface-3); }
.lead { display: inline-flex; align-items: center; gap: 8px; min-width: 0; }
.box {
  appearance: none;
  width: 18px;
  height: 18px;
  margin: 0;
  flex: none;
  border-radius: 50%;
  border: 1.5px solid var(--line-2);
  background: transparent center / 11px no-repeat;
}
.box:checked {
  border-color: var(--acc);
  background-color: var(--acc);
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath d='M5 12.5l4.2 4.2L19 7' fill='none' stroke='%2304161C' stroke-width='3' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
}
.dlg-acts { display: flex; gap: 10px; margin-top: 18px; }
.ghost, .ok { flex: 1; height: 36px; border-radius: 10px; font-size: 13px; }
.ghost { color: var(--t2); background: var(--bg-1); border: 1px solid var(--line); }
.ok { background: var(--acc); color: var(--on-acc); font-weight: 600; }
.dot.mix { background: linear-gradient(135deg, #E5344B, #2E9BFF 50%, #31C27C); }
.scroll {
  flex: 1;
  min-height: 0;
  overflow: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--acc) transparent;
}
.scroll::-webkit-scrollbar { width: 10px; }
.scroll::-webkit-scrollbar-track {
  background: transparent;
  margin: 6px 0;
}
.scroll::-webkit-scrollbar-thumb {
  background: color-mix(in srgb, var(--acc) 72%, var(--surface));
  border: 2px solid transparent;
  border-radius: 999px;
  background-clip: padding-box;
}
.scroll::-webkit-scrollbar-thumb:hover {
  background: var(--acc);
  background-clip: padding-box;
  border: 2px solid transparent;
}
@media (max-width: 767px) {
  .brand strong { font-size: 19px; }
  .body { padding-top: 12px; }
  .plabel, .pf { height: 27px; }
  .pf { font-size: 11px; }
}
</style>
