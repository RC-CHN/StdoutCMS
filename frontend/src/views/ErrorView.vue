<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const code = computed(() => Number(route.query.code) || Number(route.params.code) || 500)
const rawMessage = computed(() => (route.query.message as string) || '')

interface ErrorDef {
  signal: string
  label: string
  hint: string
}

const errorDefs: Record<number, ErrorDef> = {
  400: { signal: 'SIGBADREQ', label: 'BAD REQUEST', hint: 'The server could not parse your request. Try a different command.' },
  401: { signal: 'SIGAUTH', label: 'UNAUTHORIZED', hint: 'Authentication required. Use `sudo` or provide valid credentials.' },
  403: { signal: 'SIGPERM', label: 'PERMISSION DENIED', hint: 'You do not have permission to access this resource. Contact root.' },
  404: { signal: 'SIGABRT', label: 'FILE NOT FOUND', hint: 'The requested file does not exist in this filesystem. Try `ls -lh` to browse available content.' },
  405: { signal: 'SIGMETHOD', label: 'METHOD NOT ALLOWED', hint: 'That operation is not permitted on this resource.' },
  408: { signal: 'SIGTIMEOUT', label: 'REQUEST TIMEOUT', hint: 'The server waited too long. Check your connection and retry.' },
  429: { signal: 'SIGRATE', label: 'RATE LIMITED', hint: 'Too many requests. Slow down and retry later.' },
  500: { signal: 'SIGSEGV', label: 'INTERNAL ERROR', hint: 'Something went wrong on the server. The sysadmin has been paged. (Probably.)' },
  502: { signal: 'SIGPIPE', label: 'BAD GATEWAY', hint: 'Upstream server is unreachable. Check the proxy chain.' },
  503: { signal: 'SIGSTOP', label: 'SERVICE UNAVAILABLE', hint: 'The service is temporarily down for maintenance. Stand by.' },
}

const def = computed(() => errorDefs[code.value] ?? errorDefs[500])
const message = computed(() => rawMessage.value || def.value.label)
</script>

<template>
  <div class="error-view">
    <pre class="error-art">
========================================
[ ERROR ]  CODE: {{ code }}  |  SIGNAL: {{ def.signal }}
========================================
    </pre>

    <div class="error-label">
      > {{ message }}
    </div>

    <p class="error-hint">{{ def.hint }}</p>

    <div class="error-dump" v-if="rawMessage">
      <p>Additional info:</p>
      <pre class="dump-block">{{ rawMessage }}</pre>
    </div>

    <nav class="error-nav">
      <button class="btn" @click="router.push('/')">cd /</button>
    </nav>

    <div class="eof-marker">CORE DUMP</div>
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
