<script setup lang="ts">
import { computed, ref } from 'vue'
import { COPY, fill } from '../copy/messages'
import { useSearch } from '../composables/useSearch'
import { groupTracks, sourceColor } from '../utils/group'
import TrackRow from './TrackRow.vue'

const props = defineProps<{ source: string }>()
const emit = defineEmits<{ pick: [query: string] }>()
const { state, retrySearch } = useSearch()
const folded = ref<Record<string, boolean>>({})

const groups = computed(() => groupTracks(state.items, props.source))
const visibleCount = computed(() => groups.value.reduce((n, group) => n + group.rows.length, 0))

function toggle(id: string): void {
  folded.value[id] = !folded.value[id]
}
</script>

<template>
  <section class="results" :aria-label="COPY.resultsLabel">
    <div class="sr-only" role="status" aria-live="polite">{{ state.announcement }}</div>

    <template v-if="state.phase === 'loading'">
      <div class="skeleton" role="status" :aria-label="COPY.skeletonLabel">
        <div v-for="i in 6" :key="i" class="sk">
          <span class="box" />
          <span class="lines"><i /><i /></span>
        </div>
      </div>
    </template>

    <template v-else-if="state.phase === 'results'">
      <p class="count">{{ COPY.found(visibleCount) }}</p>
      <div v-for="(group, index) in groups" :key="group.id" class="grp">
        <div class="grp-h">
          <span class="no">{{ index + 1 }}</span>
          <i class="dot" :class="{ mix: group.kind === 'merged' }" :style="group.kind === 'merged' ? undefined : { background: sourceColor(group.label) }" />
          <span class="gname">{{ group.kind === 'merged' ? COPY.mergedGroup : group.label }}</span>
          <span class="gn">{{ group.rows.length }} 条</span>
          <span class="line" />
          <button type="button" class="fold" @click="toggle(group.id)">
            {{ folded[group.id] ? COPY.groupExpand : COPY.groupFold }}
          </button>
        </div>
        <div class="grp-body" :class="{ shut: folded[group.id] }">
          <div class="grp-rows">
            <TrackRow v-for="row in group.rows" :key="row.key" :row="row" />
          </div>
        </div>
      </div>
    </template>

    <div v-else-if="state.phase === 'noResults'" class="empty">
      <div class="rings">∅</div>
      <b>{{ fill(COPY.noResultsTitle, { kw: state.keyword }) }}</b>
      <p>{{ COPY.noResultsSub }}</p>
    </div>

    <div v-else-if="state.phase === 'error'" class="empty">
      <div class="rings">!</div>
      <b>{{ COPY.searchFailedTitle }}</b>
      <p>{{ state.errorHint || COPY.searchFailedSub }}</p>
      <button type="button" class="retry" @click="retrySearch">{{ COPY.retry }}</button>
    </div>

    <div v-else class="empty">
      <div class="rings">𝄞</div>
      <b>{{ COPY.emptyTitle }}</b>
      <p>{{ COPY.emptySub }}</p>
      <div v-if="state.recent.length" class="recent">
        <span>{{ COPY.recentLabel }}</span>
        <button v-for="q in state.recent" :key="q" type="button" @click="emit('pick', q)">{{ q }}</button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.results { padding-bottom: 12px; }
.count { margin: 18px 0 4px; font-size: 12px; color: var(--t3); }
.count b { color: var(--t1); font-weight: 600; }
.grp { margin-top: 14px; }
.grp-h { display: flex; align-items: center; gap: 9px; padding-bottom: 8px; }
.no { font-family: var(--serif); font-style: italic; font-size: 17px; color: var(--acc); min-width: 14px; }
.gname { font-size: 12px; color: var(--t2); font-weight: 600; }
.gn { font-family: var(--mono); font-size: 10px; color: var(--t4); }
.line { flex: 1; height: 1px; background: var(--line); }
.fold { font-size: 11px; color: var(--t4); }
.fold:hover { color: var(--t2); }
.grp-body {
  display: grid;
  grid-template-rows: 1fr;
  transition: grid-template-rows .28s var(--ease);
}
.grp-body.shut { grid-template-rows: 0fr; }
.grp-rows { min-height: 0; overflow: hidden; }
.dot { width: 8px; height: 8px; border-radius: 50%; flex: none; }
.dot.mix { background: linear-gradient(135deg, #E5344B, #2E9BFF 50%, #31C27C); }
.hold { min-height: 120px; }
.empty { text-align: center; padding: 48px 16px 24px; }
.rings {
  width: 56px; height: 56px; margin: 0 auto 10px; border-radius: 50%;
  border: 1px solid var(--line-2); display: grid; place-items: center;
  background: radial-gradient(circle, var(--acc-soft), transparent 70%);
  font-family: var(--serif); color: var(--acc); font-size: 22px;
}
.empty b { display: block; font-size: 14px; margin-bottom: 4px; }
.empty p { margin: 0 auto; max-width: 36ch; font-size: 12px; color: var(--t4); }
.recent { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; margin-top: 16px; align-items: center; }
.recent span { font-size: 11px; color: var(--t4); }
.recent button, .retry {
  height: 28px; padding: 0 10px; border-radius: var(--r-pill);
  border: 1px solid var(--line-2); color: var(--t2); font-size: 12px; background: var(--bg-1);
}
.retry { margin-top: 14px; background: var(--acc); color: var(--on-acc); border-color: var(--acc); }
.skeleton { display: grid; gap: 14px; padding-top: 18px; }
.sk { display: grid; grid-template-columns: 46px minmax(0, 1fr); gap: 12px; align-items: center; }
.box, .lines i {
  background: linear-gradient(90deg, var(--surface) 0%, var(--surface-3) 50%, var(--surface) 100%);
  background-size: 220% 100%;
  animation: sh 1.4s linear infinite;
  border-radius: 4px;
}
.box { width: 46px; height: 46px; border-radius: 6px; }
.lines { display: grid; gap: 8px; }
.lines i { display: block; height: 10px; }
.lines i:last-child { width: 40%; height: 8px; }
@keyframes sh { to { background-position: -220% 0; } }
</style>
