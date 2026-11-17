<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { BRAND, COPY } from '../copy/messages'
import SearchPill from '../components/SearchPill.vue'

const router = useRouter()
const query = ref('')

function onSubmit(): void {
  const q = query.value.trim()
  if (!q) return
  void router.push({ name: 'detail', query: { q } })
}
</script>

<template>
  <main class="cover">
    <div class="glow" aria-hidden="true" />
    <div class="mast">
      <div class="corner">{{ BRAND.corner }}</div>
      <h1 class="wordmark">Ca<span>d</span>enza</h1>
      <p class="lead">{{ COPY.coverLead }}</p>
    </div>
    <div class="spacer" aria-hidden="true" />
    <SearchPill
      class="find"
      variant="hero"
      :model-value="query"
      @update:model-value="query = $event"
      @submit="onSubmit"
    />
    <div class="spacer" aria-hidden="true" />
    <svg class="flourish" viewBox="0 0 1280 168" fill="none" aria-hidden="true">
        <g class="lines" stroke="currentColor" stroke-width="1.35">
          <line x1="0" y1="40" x2="1280" y2="40" />
          <line x1="0" y1="66" x2="1280" y2="66" />
          <line x1="0" y1="92" x2="1280" y2="92" />
          <line x1="0" y1="118" x2="1280" y2="118" />
          <line x1="0" y1="144" x2="1280" y2="144" />
        </g>
        <path
          class="melody"
          d="M56 150 C 214 150, 300 36, 428 78 C 542 116, 574 158, 668 132 C 754 108, 790 32, 882 50 C 982 70, 1002 132, 1082 114 C 1140 102, 1168 76, 1224 84"
          stroke="currentColor"
          stroke-width="1.8"
          stroke-linecap="round"
        />
        <g class="heads" fill="currentColor">
          <ellipse cx="428" cy="78" rx="9" ry="6.4" transform="rotate(-18 428 78)" />
          <ellipse cx="668" cy="132" rx="9" ry="6.4" transform="rotate(-12 668 132)" />
          <ellipse cx="882" cy="50" rx="9" ry="6.4" transform="rotate(-20 882 50)" />
          <ellipse cx="1082" cy="114" rx="9" ry="6.4" transform="rotate(-10 1082 114)" />
        </g>
        <g class="stems" stroke="currentColor" stroke-width="1.4">
          <line x1="436" y1="74" x2="436" y2="36" />
          <line x1="676" y1="128" x2="676" y2="90" />
          <line x1="890" y1="46" x2="890" y2="8" />
          <line x1="1090" y1="110" x2="1090" y2="72" />
        </g>
      </svg>
  </main>
</template>

<style scoped>
.cover {
  min-height: 100dvh;
  position: relative;
  display: grid;
  grid-template-rows: auto 1fr auto 1fr auto;
  align-items: start;
  padding: 88px 56px 48px;
  overflow: hidden;
  background: var(--bg);
}
.glow {
  position: absolute;
  left: -12%;
  top: -18%;
  width: 720px;
  height: 520px;
  border-radius: 50%;
  background: radial-gradient(circle, color-mix(in srgb, var(--acc) 16%, transparent), transparent 68%);
  pointer-events: none;
}
.mast {
  position: relative;
  z-index: 1;
  width: min(1180px, 100%);
  margin-top: 8px;
  margin-left: 96px;
}
.corner {
  font-size: 12px;
  letter-spacing: .28em;
  text-transform: uppercase;
  color: var(--t4);
}
.wordmark {
  font-family: var(--serif);
  font-style: italic;
  font-weight: 400;
  font-size: clamp(88px, 12vw, 148px);
  line-height: .9;
  letter-spacing: -.03em;
  margin: 14px 0 0;
}
.wordmark span { color: var(--acc); }
.lead {
  margin: 18px 0 0;
  max-width: 34em;
  font-size: 22px;
  line-height: 1.5;
  color: var(--t2);
}
.spacer { min-height: 28px; }
.find {
  justify-self: center;
  width: min(720px, 100%);
  z-index: 1;
}
.flourish {
  justify-self: center;
  display: block;
  width: min(1320px, calc(100vw - 80px));
  height: auto;
  color: var(--acc);
  z-index: 1;
}
.lines { opacity: .42; }
.melody, .stems { opacity: .78; }
.heads { opacity: .95; }
@media (max-width: 1023px) and (min-width: 768px) {
  .cover { padding: 56px 32px 36px; }
  .mast {
    margin-left: auto;
    margin-right: auto;
    width: min(680px, 100%);
    text-align: center;
  }
  .lead { margin-left: auto; margin-right: auto; }
  .flourish { width: min(920px, calc(100vw - 64px)); }
}
@media (max-width: 767px) {
  .cover { padding: 32px 20px 24px; justify-items: center; }
  .mast {
    margin-left: auto;
    margin-right: auto;
    width: 100%;
    text-align: center;
  }
  .wordmark { font-size: clamp(56px, 16vw, 80px); }
  .lead { font-size: 16px; margin-left: auto; margin-right: auto; }
  .find, .flourish { width: 100%; }
}
@media (max-height: 760px) and (min-width: 768px) {
  .cover { padding-top: 28px; padding-bottom: 20px; }
  .wordmark { font-size: clamp(64px, 9vw, 96px); }
  .lead { font-size: 17px; margin-top: 10px; }
  .spacer { min-height: 16px; }
}
</style>
