<script setup lang="ts">
import type { EditorUploads } from '../composables/useEditorUploads'
const props = defineProps<{ uploads: EditorUploads }>()
const { pending, failures } = props.uploads
</script>

<template>
  <div v-if="pending.length" class="upload-status" role="status">
    Uploading {{ pending.length }} file(s)… Wait before saving.
  </div>
  <div v-for="job in failures" :key="job.id" class="upload-status upload-error" role="alert">
    <p><strong>{{ job.file.name }}</strong>: {{ job.error }}</p>
    <div class="state-actions">
      <button @click="uploads.retry(job.id)">RETRY UPLOAD</button>
      <button @click="uploads.remove(job.id)">REMOVE</button>
    </div>
  </div>
</template>

<style scoped>
.upload-status { padding: 0.8rem; margin-block: 0.75rem; border: 1px dashed var(--border); overflow-wrap: anywhere; }
.upload-error { border-color: var(--danger); }
.state-actions { margin-top: 0.5rem; }
</style>
