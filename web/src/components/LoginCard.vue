<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { ApiError, login as apiLogin } from '../api'
import { A11Y, BRAND, COPY, mapError } from '../copy/messages'
import { handleApiFailure, markAuthenticated } from '../composables/useSession'

const props = withDefaults(defineProps<{ expired?: boolean }>(), { expired: false })
const emit = defineEmits<{ success: [] }>()

type LoginStatus = 'idle' | 'typing' | 'pending' | 'error' | 'success'

const password = ref('')
const status = ref<LoginStatus>('idle')
const errorMessage = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

async function focusAndSelect(): Promise<void> {
  await nextTick()
  const el = inputRef.value
  if (!el) return
  el.focus()
  el.select()
}

function onPasswordInput(e: Event): void {
  password.value = (e.target as HTMLInputElement).value
  if (status.value === 'error' || status.value === 'idle') {
    status.value = 'typing'
    errorMessage.value = ''
  }
}

function onEsc(): void {
  password.value = ''
  status.value = 'idle'
  errorMessage.value = ''
  inputRef.value?.focus()
}

async function onSubmit(): Promise<void> {
  if (status.value === 'pending') return
  status.value = 'pending'
  errorMessage.value = ''
  try {
    await apiLogin(password.value)
    status.value = 'success'
    markAuthenticated()
    emit('success')
  } catch (err) {
    const code = err instanceof ApiError ? err.code : 'LOGIN_UNAVAILABLE'
    if (handleApiFailure(code)) return
    const fallback = code === 'LOGIN_INVALID' ? COPY.loginInvalid : COPY.loginUnavailable
    errorMessage.value = mapError(code, fallback)
    status.value = 'error'
    await focusAndSelect()
  }
}

onMounted(() => {
  const isDesktop = window.matchMedia('(min-width: 768px)').matches
  if (isDesktop || props.expired) inputRef.value?.focus()
})
</script>

<template>
  <form class="card" @submit.prevent="onSubmit">
    <div class="brand">
      <strong>{{ BRAND.name }}</strong>
      <span>{{ BRAND.corner }}</span>
    </div>
    <p class="sub">{{ COPY.loginSub }}</p>
    <p v-if="expired" class="err" role="alert">{{ COPY.loginExpired }}</p>
    <div class="label">{{ COPY.loginAccess }}</div>
    <div class="field" :class="{ err: status === 'error' }">
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <rect x="4" y="10" width="16" height="10" rx="2" fill="none" stroke="currentColor" stroke-width="1.6" />
        <path d="M8 10V7a4 4 0 0 1 8 0v3" fill="none" stroke="currentColor" stroke-width="1.6" />
      </svg>
      <input
        id="password-input"
        ref="inputRef"
        :type="'password'"
        :value="password"
        :aria-label="A11Y.passwordLabel"
        :aria-invalid="status === 'error' ? 'true' : undefined"
        autocomplete="current-password"
        @input="onPasswordInput"
        @keydown.esc.prevent="onEsc"
      />
    </div>
    <p v-if="status === 'error'" id="login-error" class="err" role="alert">{{ errorMessage }}</p>
    <button type="submit" class="enter" :disabled="status === 'pending'">
      {{ status === 'pending' ? COPY.loginPending : COPY.loginEnter }}
    </button>
    <p class="foot">{{ COPY.loginHint }}</p>
  </form>
</template>

<style scoped>
.card {
  position: relative;
  width: min(368px, 100%);
  background: color-mix(in srgb, var(--surface) 92%, transparent);
  border: 1px solid var(--line-2);
  border-radius: var(--r-xl);
  padding: 30px 28px 24px;
  box-shadow: 0 30px 70px -30px rgba(0, 0, 0, .9);
}
.brand { display: flex; align-items: baseline; justify-content: space-between; }
.brand strong { font-family: var(--serif); font-style: italic; font-weight: 400; font-size: 26px; }
.brand span { font-size: 10px; letter-spacing: .22em; text-transform: uppercase; color: var(--t4); }
.sub { margin: 8px 0 22px; font-size: 12px; color: var(--t3); }
.label { font-size: 11px; letter-spacing: .12em; text-transform: uppercase; color: var(--t4); margin-bottom: 6px; }
.field {
  display: flex;
  align-items: center;
  gap: 9px;
  height: 42px;
  padding: 0 12px;
  background: var(--bg);
  border: 1px solid var(--line-2);
  border-radius: var(--r-s);
}
.field.err { border-color: rgba(244, 115, 111, .55); box-shadow: 0 0 0 3px rgba(244, 115, 111, .09); }
.field svg { width: 14px; height: 14px; color: var(--t4); }
.field input {
  flex: 1;
  background: none;
  border: 0;
  outline: none;
  box-shadow: none;
  color: var(--t1);
  font-family: var(--mono);
  font-size: 16px;
  letter-spacing: .08em;
}
.field input:focus,
.field input:focus-visible {
  outline: none;
  box-shadow: none;
}
.err { margin: 8px 0 0; font-size: 12px; color: var(--err); }
.enter {
  width: 100%;
  height: 42px;
  margin-top: 16px;
  border-radius: var(--r-s);
  background: var(--acc);
  color: var(--on-acc);
  font-size: 14px;
  font-weight: 600;
}
.enter:hover:not(:disabled) { background: color-mix(in srgb, var(--acc) 82%, white); }
.enter:disabled { opacity: .7; }
.foot {
  margin: 16px 0 0;
  padding-top: 14px;
  border-top: 1px solid var(--line);
  font-size: 11px;
  color: var(--t4);
  line-height: 1.7;
}
@media (max-width: 767px) {
  .card {
    width: 100%;
    min-height: 100dvh;
    border: 0;
    border-radius: 0;
    box-shadow: none;
    display: flex;
    flex-direction: column;
    justify-content: center;
  }
}
</style>
