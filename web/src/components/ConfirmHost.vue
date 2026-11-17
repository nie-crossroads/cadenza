<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { COPY } from '../copy/messages'
import { settleConfirm, useConfirm } from '../composables/useConfirm'

const { state } = useConfirm()

function onKey(event: KeyboardEvent): void {
  if (!state.open || event.key !== 'Escape') return
  event.preventDefault()
  settleConfirm(false)
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <div v-if="state.open" class="mask" @click.self="settleConfirm(false)">
    <div class="card" role="dialog" aria-modal="true" :aria-label="COPY.addConfirmTitle">
      <h2>{{ COPY.addConfirmTitle }}</h2>
      <p>{{ COPY.addConfirmBody(state.name) }}</p>
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
}
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
