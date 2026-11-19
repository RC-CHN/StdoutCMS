<script setup lang="ts">
import type { MetaResponse } from '../api/meta'

defineProps<{ meta: MetaResponse | null }>()
</script>

<template>
  <footer class="site-footer">
    <ul v-if="meta" class="site-stats" aria-label="Site information">
      <li>build <b>{{ meta.version }}</b></li>
      <li>articles <b>{{ meta.posts }}</b></li>
      <li>projects <b>{{ meta.projects }}</b></li>
      <li>uptime <b>{{ meta.uptime }}</b></li>
    </ul>
    <small>© 2026 Ruochen Pan</small>
  </footer>
</template>

<style scoped>
.site-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem 2rem;
  padding: 0.65rem 1rem;
  border-top: 2px solid var(--border);
  background: var(--bg);
  color: var(--muted);
  font-size: 0.8rem;
}
.site-stats { display: flex; flex-wrap: wrap; gap: 0.25rem 1.25rem; list-style: none; }
.site-stats li { white-space: nowrap; }
.site-stats b { color: var(--fg); }
.site-footer small { font-size: inherit; }
</style>
