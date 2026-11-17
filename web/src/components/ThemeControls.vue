<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { COPY } from '../copy/messages'
import { PRESET_COLORS, useTheme } from '../composables/useTheme'

const { state, toggle, setColor, setHue, sameColor } = useTheme()
const open = ref(false)
const wrap = ref<HTMLElement | null>(null)

function onDoc(event: Event): void {
  if (!open.value) return
  const el = wrap.value
  if (el && !el.contains(event.target as Node)) open.value = false
}

function togglePanel(event: MouseEvent): void {
  event.stopPropagation()
  open.value = !open.value
}

onMounted(() => document.addEventListener('pointerdown', onDoc))
onUnmounted(() => document.removeEventListener('pointerdown', onDoc))
</script>

<template>
  <div class="actions">
    <button
      class="icon-btn"
      type="button"
      :aria-label="state.dark ? COPY.themeToLight : COPY.themeToDark"
      :title="state.dark ? COPY.themeToLight : COPY.themeToDark"
      @click="toggle"
    >
      <svg v-if="state.dark" viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="12" cy="12" r="4" fill="none" stroke="currentColor" stroke-width="1.8" />
        <path d="M12 3v2.2M12 18.8V21M3 12h2.2M18.8 12H21M5.6 5.6l1.6 1.6M16.8 16.8l1.6 1.6M18.4 5.6l-1.6 1.6M7.2 16.8l-1.6 1.6" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </svg>
      <svg v-else viewBox="0 0 24 24" aria-hidden="true">
        <path d="M15 4.5A7.5 7.5 0 1 0 19.5 15 6.2 6.2 0 0 1 15 4.5z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" />
      </svg>
    </button>
    <div ref="wrap" class="setting-wrap">
      <button
        class="icon-btn"
        type="button"
        :aria-label="COPY.themeSettings"
        :aria-expanded="open"
        :title="COPY.themeSettings"
        @click="togglePanel"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 0 0 2.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 0 0 1.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 0 0-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 0 0-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 0 0-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 0 0-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 0 0 1.066-2.573c-.94-1.543.826-3.31 2.37-2.37 1 .608 2.296.07 2.572-1.065z" />
          <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" stroke-width="1.8" />
        </svg>
      </button>
      <div v-show="open" class="float-panel" @click.stop>
        <p class="hue-row">{{ COPY.themeColor }}</p>
        <div class="presets">
          <button
            v-for="value in PRESET_COLORS"
            :key="value"
            type="button"
            class="swatch"
            :class="{ on: sameColor(state.color, value) }"
            :style="{ background: value }"
            :aria-label="COPY.themeColor"
            :aria-pressed="sameColor(state.color, value)"
            @click="setColor(value)"
          />
        </div>
        <input
          class="hue"
          type="range"
          min="0"
          max="360"
          :aria-label="COPY.themeColor"
          :value="state.hue"
          @input="setHue(Number(($event.target as HTMLInputElement).value))"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.actions { display: flex; align-items: center; gap: 8px; }
.icon-btn {
  width: 40px;
  height: 40px;
  border: none;
  border-radius: 12px;
  background: var(--surface);
  color: var(--acc);
  box-shadow: 0 4px 10px -6px rgb(0 0 0 / .45);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
}
.icon-btn:hover { background: var(--surface-3); }
.icon-btn svg { width: 20px; height: 20px; }
.setting-wrap { position: relative; }
.float-panel {
  position: absolute;
  right: 0;
  top: calc(100% + 8px);
  width: 216px;
  padding: 14px;
  z-index: 40;
  background: var(--surface);
  border: 1px solid var(--acc);
  border-radius: 16px;
  box-shadow: 0 12px 32px -16px rgb(0 0 0 / .55);
}
.hue-row {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 700;
  color: var(--t1);
}
.presets {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  margin-bottom: 12px;
  justify-items: center;
}
.swatch {
  width: 22px;
  height: 22px;
  border: 2px solid transparent;
  border-radius: 50%;
  padding: 0;
}
.swatch.on {
  border-color: var(--t1);
  box-shadow: 0 0 0 2px var(--surface);
}
.hue {
  -webkit-appearance: none;
  appearance: none;
  width: 100%;
  height: 24px;
  background-image: var(--color-selection-bar);
  border-radius: 4px;
}
.hue::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 8px;
  height: 16px;
  border-radius: 2px;
  background: rgb(255 255 255 / .9);
  cursor: pointer;
}
</style>
