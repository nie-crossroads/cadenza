<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { A11Y } from '../copy/messages'
import LoginCard from '../components/LoginCard.vue'

const route = useRoute()
const router = useRouter()
const expired = computed(() => route.query.reason === 'expired')

function onSuccess(): void {
  const raw = route.query.redirect
  const target = typeof raw === 'string' && raw.startsWith('/') && !raw.startsWith('//')
    ? raw
    : '/'
  void router.push(target)
}
</script>

<template>
  <main class="stage" tabindex="-1" :aria-label="A11Y.loginMain">
    <LoginCard :expired="expired" @success="onSuccess" />
  </main>
</template>

<style scoped>
.stage {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  padding: 56px 24px;
  position: relative;
  background:
    radial-gradient(ellipse 60% 70% at 50% 0%, color-mix(in srgb, var(--acc) 12%, transparent), transparent 70%),
    var(--bg);
}
.stage::before {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: repeating-linear-gradient(to bottom, transparent 0 17px, var(--staff) 17px 18px);
  mask-image: linear-gradient(to bottom, transparent, #000 30%, #000 70%, transparent);
}
@media (max-width: 767px) {
  .stage { padding: 0; align-items: stretch; }
}
</style>
