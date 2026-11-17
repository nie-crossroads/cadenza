<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { COPY, qualityLabel } from '../copy/messages'
import {
  cancelPlayableCheck,
  setQualityChoice,
  settleConfirm,
  settleNotice,
  settleQuality,
  useConfirm,
} from '../composables/useConfirm'

const { state, quality, checking, notice } = useConfirm()

function onKey(event: KeyboardEvent): void {
  if (event.key !== 'Escape') return
  if (quality.open) {
    event.preventDefault()
    settleQuality(false)
    return
  }
  if (checking.open) {
    event.preventDefault()
    cancelPlayableCheck()
    return
  }
  if (notice.open) {
    event.preventDefault()
    settleNotice()
    return
  }
  if (!state.open) return
  event.preventDefault()
  settleConfirm(false)
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div v-if="quality.open" class="mask" @click.self="settleQuality(false)">
    <div class="card" role="dialog" aria-modal="true" :aria-label="COPY.addQualityTitle">
      <h2>{{ COPY.addQualityTitle }}</h2>
      <p>{{ COPY.addQualityHint(quality.name) }}</p>
      <div class="options">
        <label
          v-for="code in quality.options"
          :key="code"
          class="opt"
          :class="{ 'is-on': quality.selected === code }"
        >
          <span>{{ qualityLabel(code) }}</span>
          <input
            type="radio"
            name="add-quality"
            :value="code"
            :checked="quality.selected === code"
            @change="setQualityChoice(code)"
          />
        </label>
      </div>
      <div class="acts">
        <button type="button" class="ghost" @click="settleQuality(false)">{{ COPY.addConfirmCancel }}</button>
        <button type="button" class="ok" @click="settleQuality(true)">{{ COPY.addQualityNext }}</button>
      </div>
    </div>
  </div>
  <div v-else-if="checking.open" class="mask" @click.self="cancelPlayableCheck">
    <div class="card" role="dialog" aria-modal="true" :aria-label="COPY.addPlayableTitle">
      <h2>{{ COPY.addPlayableTitle }}</h2>
      <p>{{ COPY.addPlayableChecking }}</p>
      <div class="acts">
        <button type="button" class="ghost" @click="cancelPlayableCheck">{{ COPY.addConfirmCancel }}</button>
      </div>
    </div>
  </div>
  <div v-else-if="notice.open" class="mask" @click.self="settleNotice">
    <div class="card" role="dialog" aria-modal="true" :aria-label="COPY.addUnplayableTitle">
      <h2>{{ COPY.addUnplayableTitle }}</h2>
      <p>{{ COPY.addUnplayableBody }}</p>
      <div class="acts">
        <button type="button" class="ok" @click="settleNotice">{{ COPY.addUnplayableOk }}</button>
      </div>
    </div>
  </div>
  <div v-else-if="state.open" class="mask" @click.self="settleConfirm(false)">
    <div class="card" role="dialog" aria-modal="true" :aria-label="COPY.addConfirmTitle">
      <h2>{{ COPY.addConfirmTitle }}</h2>
      <p>{{ COPY.addConfirmBody(state.name, state.tone) }}</p>
      <div class="acts">
        <button type="button" class="ghost" @click="settleConfirm(false)">{{ COPY.addConfirmCancel }}</button>
        <button type="button" class="ok" @click="settleConfirm(true)">{{ COPY.addConfirmOk }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.mask {
  position: fixed;
  inset: 0;
  z-index: 80;
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
h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  text-align: center;
}
p {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--t3);
  text-align: center;
}
.options {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 14px;
}
.opt {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 40px;
  padding: 0 12px;
  border: 1px solid var(--line);
  border-radius: 10px;
  background: var(--bg-1);
  color: var(--t1);
}
.opt.is-on { border-color: var(--acc); }
.opt input { accent-color: var(--acc); }
.acts {
  display: flex;
  gap: 10px;
  margin-top: 18px;
}
.ghost, .ok {
  flex: 1;
  height: 36px;
  border-radius: 10px;
  font-size: 13px;
}
.ghost {
  color: var(--t2);
  background: var(--bg-1);
  border: 1px solid var(--line);
}
.ok {
  background: var(--acc);
  color: var(--on-acc);
  font-weight: 600;
}
</style>
