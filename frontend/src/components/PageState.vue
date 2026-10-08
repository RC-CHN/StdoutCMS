<script setup lang="ts">
defineProps<{ mode: 'loading' | 'error' | 'empty'; message: string; retry?: boolean }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="page-state" :class="`state-${mode}`" :aria-busy="mode === 'loading'" :role="mode === 'error' ? 'alert' : 'status'">
    <span class="state-label">{{ mode === 'loading' ? '[ WAIT ]' : mode === 'error' ? '[ FAIL ]' : '[ EMPTY ]' }}</span>
    <p>{{ message }}</p>
    <button v-if="retry" @click="$emit('retry')">RETRY</button>
  </div>
</template>

<style scoped>
.page-state { padding: 2rem 1rem; margin-bottom: 1rem; border: 1px dashed var(--border); display: flex; flex-direction: column; align-items: flex-start; gap: 0.75rem; overflow-wrap: anywhere; }
.state-label { font-weight: bold; color: var(--muted); }
.state-error .state-label { color: var(--danger); }
.state-loading { min-height: 180px; }
</style>
