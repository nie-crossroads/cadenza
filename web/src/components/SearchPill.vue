<script setup lang="ts">
import { ref } from 'vue'
import { A11Y, COPY } from '../copy/messages'

withDefaults(
  defineProps<{ modelValue: string; loading?: boolean; variant?: 'bar' | 'hero' }>(),
  { loading: false, variant: 'bar' },
)
const emit = defineEmits<{
  'update:modelValue': [value: string]
  submit: []
  clear: []
}>()

const inputRef = ref<HTMLInputElement | null>(null)

function onInput(e: Event): void {
  emit('update:modelValue', (e.target as HTMLInputElement).value)
}
function onClear(): void {
  emit('clear')
  inputRef.value?.focus()
}
function focus(): void { inputRef.value?.focus() }
defineExpose({ focus })
</script>

<template>
  <form class="search" :class="variant" role="search" @submit.prevent="emit('submit')">
    <div class="field">
      <svg v-if="variant !== 'hero'" viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="11" cy="11" r="6.5" fill="none" stroke="currentColor" stroke-width="1.8" />
        <path d="M16 16l4.5 4.5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </svg>
      <input
        ref="inputRef"
        type="search"
        enterkeyhint="search"
        :value="modelValue"
        :placeholder="COPY.searchPlaceholder"
        :aria-label="A11Y.searchLabel"
        :aria-busy="loading ? 'true' : undefined"
        @input="onInput"
        @keydown.esc.prevent="onClear"
      />
      <kbd v-if="variant !== 'hero'">⌘K</kbd>
      <button v-if="variant === 'hero'" type="submit" class="go">{{ COPY.searchGo }}</button>
    </div>
    <button v-if="variant !== 'hero'" type="submit" class="go">{{ COPY.searchGo }}</button>
  </form>
</template>

<style scoped>
.search { display: flex; gap: 10px; align-items: stretch; }
.field {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  height: 50px;
  padding: 0 14px;
  background: var(--surface);
  border: 1px solid var(--line-2);
  border-radius: var(--r-m);
}
.field svg { width: 17px; height: 17px; color: var(--t3); flex: none; }
input {
  flex: 1;
  min-width: 0;
  background: none;
  border: 0;
  outline: none;
  box-shadow: none;
  font-size: 15px;
  color: var(--t1);
}
input:focus,
input:focus-visible {
  outline: none;
  box-shadow: none;
}
input::placeholder { color: var(--t4); }
input::-webkit-search-cancel-button { appearance: none; }
kbd {
  font-family: var(--mono);
  font-size: 10px;
  color: var(--t4);
  border: 1px solid var(--line-2);
  border-radius: 4px;
  padding: 2px 5px;
}
.go {
  height: 50px;
  padding: 0 22px;
  border-radius: var(--r-m);
  background: var(--acc);
  color: var(--on-acc);
  font-weight: 600;
  font-size: 14px;
}
.go:hover { background: color-mix(in srgb, var(--acc) 82%, white); }
.hero .field {
  position: relative;
  height: 58px;
  padding: 0;
  border-radius: 999px;
  transition: border-color .28s var(--ease), box-shadow .28s var(--ease);
}
.hero .field:focus-within {
  border-color: var(--acc);
  box-shadow: 0 0 0 4px var(--acc-soft);
}
.hero input,
.hero input:placeholder-shown {
  height: 100%;
  padding: 0 148px;
  text-align: center;
}
.hero .go {
  position: absolute;
  right: 6px;
  top: 6px;
  height: calc(100% - 12px);
  min-width: 104px;
  padding: 0 28px;
  border-radius: 999px;
}
@media (max-width: 1023px) {
  .bar .field, .bar .go { height: 46px; }
}
@media (max-width: 767px) {
  .bar .go, .bar kbd { display: none; }
  .bar .field { height: 40px; }
  input { font-size: 16px; }
  .hero .field { height: 52px; }
  .hero input,
  .hero input:placeholder-shown { padding-left: 112px; padding-right: 112px; text-align: center; }
  .hero .go { min-width: 84px; padding: 0 20px; }
}
</style>
