<script setup lang="ts">
import type { EditorUploads } from '../composables/useEditorUploads'
const props = defineProps<{ uploads: EditorUploads }>()
const { pending, failures } = props.uploads
</script>

<template>
  <div v-if="pending.length" class="upload-status" role="status">
    <strong>[ WAIT ]</strong> Uploading {{ pending.length }} {{ pending.length === 1 ? 'file' : 'files' }}…
  </div>
  <div v-for="job in failures" :key="job.id" class="upload-status upload-error" role="alert">
    <p><strong>[ ERROR ]</strong> Could not upload {{ job.file.name }}. {{ job.error }}</p>
    <div class="state-actions">
      <button @click="uploads.retry(job.id)">RETRY UPLOAD</button>
      <button @click="uploads.remove(job.id)">REMOVE FROM DRAFT</button>
    </div>
  </div>
</template>

<style scoped>
.upload-status { padding: 0.8rem; margin-block: 0.75rem; border: 1px dashed var(--border); overflow-wrap: anywhere; }
.upload-error { border-color: var(--danger); }
.upload-error strong { color: var(--danger); }
.state-actions { margin-top: 0.5rem; }
</style>
