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
    <p class="error-code"><strong>[ ERROR ]</strong> HTTP {{ code }}</p>
    <h1 class="error-label">{{ def.label }}</h1>
    <p class="error-hint">{{ def.hint }}</p>
    <pre v-if="rawMessage" class="dump-block">{{ rawMessage }}</pre>

    <nav class="error-nav" aria-label="Error page actions">
      <RouterLink v-if="code === 401" class="btn" to="/login">SIGN IN</RouterLink>
      <RouterLink class="btn" to="/">BROWSE ARTICLES</RouterLink>
    </nav>
  </div>
</template>

<style scoped>
.error-code {
  display: flex;
  gap: 0.75rem;
  font-size: 0.85rem;
  color: var(--muted);
  border-bottom: 1px dashed var(--border);
  padding-bottom: 0.75rem;
  margin-bottom: 1.5rem;
}

.error-code strong { color: var(--danger); }

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

.dump-block {
  margin-bottom: 1.5rem;
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

</style>
