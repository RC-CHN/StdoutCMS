<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const code = computed(() => Number(route.query.code) || Number(route.params.code) || 500)
const rawMessage = computed(() => (route.query.message as string) || '')

interface ErrorDef {
  label: string
  hint: string
}

const errorDefs: Record<number, ErrorDef> = {
  400: { label: 'Bad request', hint: 'Check the address and try again.' },
  401: { label: 'Sign in required', hint: 'Sign in to open the admin panel.' },
  403: { label: 'Permission denied', hint: 'This page requires administrator access.' },
  404: { label: 'Page not found', hint: 'Check the address or browse the article list.' },
  405: { label: 'Method not allowed', hint: 'This page does not support the requested action.' },
  408: { label: 'Request timed out', hint: 'Check your connection and try again.' },
  429: { label: 'Too many requests', hint: 'Wait a moment and try again.' },
  500: { label: 'Server error', hint: 'The server could not complete the request. Try again later.' },
  502: { label: 'Bad gateway', hint: 'The gateway could not get a response from the server. Try again later.' },
  503: { label: 'Service unavailable', hint: 'The service is temporarily unavailable. Try again later.' },
}

const def = computed(() => errorDefs[code.value] ?? { label: 'Request failed', hint: 'Return to the article list to continue.' })
</script>

<template>
  <div class="error-view">
    <pre class="error-art">
========================================
[ ERROR ]  HTTP {{ code }}
========================================
    </pre>

    <h1 class="error-label">&gt; {{ def.label }}</h1>

    <p class="error-hint">{{ def.hint }}</p>

    <div class="error-dump" v-if="rawMessage">
      <p>Additional info:</p>
      <pre class="dump-block">{{ rawMessage }}</pre>
    </div>

    <nav class="error-nav">
      <RouterLink v-if="code === 401" class="btn" to="/login">SIGN IN</RouterLink>
      <RouterLink class="btn" to="/">cd /</RouterLink>
    </nav>

    <div class="eof-marker">EOF</div>
  </div>
</template>

<style scoped>
.error-view {
  animation: glitch-in 0.3s ease;
}

.error-art {
  font-size: 0.72rem;
  line-height: 1.4;
  color: var(--muted);
  margin-bottom: 2rem;
  white-space: pre;
  overflow-x: auto;
}

.error-label {
  font-size: 1.4rem;
  font-weight: bold;
  margin-bottom: 1rem;
  color: var(--fg);
}

.error-hint {
  color: var(--muted);
  margin-bottom: 2rem;
  max-width: 600px;
}

.error-dump {
  margin-bottom: 2rem;
}

.error-dump p {
  font-size: 0.85rem;
  color: var(--muted);
  margin-bottom: 0.5rem;
}

.dump-block {
  border: 1px dashed var(--border);
  padding: 1rem;
  font-size: 0.82rem;
  color: var(--muted);
  background: rgba(128, 128, 128, 0.05);
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
}

.error-nav {
  display: flex;
  gap: 1rem;
  margin-bottom: 3rem;
}

.eof-marker {
  text-align: center;
  font-weight: bold;
  color: var(--muted);
  letter-spacing: 4px;
  font-size: 0.9rem;
}
.eof-marker::before { content: "--- [ "; }
.eof-marker::after { content: " ] ---"; }

@keyframes glitch-in {
  0%   { opacity: 0; transform: translateX(-2px); }
  20%  { opacity: 0.5; transform: translateX(2px); }
  40%  { opacity: 1; transform: translateX(0); }
  60%  { opacity: 0.8; transform: translateX(-1px); }
  100% { opacity: 1; transform: translateX(0); }
}
</style>
