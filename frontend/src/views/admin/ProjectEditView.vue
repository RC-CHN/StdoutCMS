<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listProjectsAdmin, createProject, updateProject } from '../../api/projects'
import type { ProjectPayload } from '../../api/projects'
import AdminNav from '../../components/AdminNav.vue'

const route = useRoute()
const router = useRouter()
const idParam = route.params.id as string | undefined
const isEdit = idParam !== undefined

const project = ref<ProjectPayload>({
  id: 0,
  name: '',
  description: '',
  lang: '',
  status: 'active',
  url: '',
  sortOrder: 0,
})
const status = ref('')
const saving = ref(false)

onMounted(async () => {
  if (isEdit) {
    try {
      const data = await listProjectsAdmin()
      const id = parseInt(idParam!)
      const found = data.projects.find(p => p.id === id)
      if (found) {
        project.value = { ...found }
        status.value = `loaded: ${project.value.name}`
      } else {
        status.value = 'ERROR: project not found'
      }
    } catch (e: any) {
      status.value = 'ERROR: ' + e.message
    }
  }
})

async function handleSave() {
  saving.value = true
  try {
    if (isEdit) {
      await updateProject(project.value.id, project.value)
      status.value = `SAVED: ${project.value.name}`
    } else {
      await createProject(project.value)
      status.value = `CREATED: ${project.value.name}`
      router.push('/admin/projects')
    }
  } catch (e: any) {
    status.value = 'ERROR: ' + (e.message || 'save failed')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AdminNav />

  <div class="editor-header">
    <div class="editor-prompt">
      root@k8s-node ~/admin $ <span class="muted">{{ isEdit ? `vim project_${idParam}.sh` : 'vim new_project.sh' }}</span>
    </div>
    <div class="editor-status" :class="{ 'status-error': status.startsWith('ERROR') }">
      {{ status || 'READY' }}
    </div>
  </div>

  <div class="meta-panel">
    <div class="meta-row">
      <label>NAME:</label>
      <input v-model="project.name" type="text" placeholder="project name" />
    </div>
    <div class="meta-row">
      <label>DESC:</label>
      <input v-model="project.description" type="text" placeholder="short description" />
    </div>
    <div class="meta-row">
      <label>LANG:</label>
      <input v-model="project.lang" type="text" placeholder="Go / TypeScript / Python" />
    </div>
    <div class="meta-row">
      <label>STATUS:</label>
      <select v-model="project.status">
        <option>active</option>
        <option>archived</option>
        <option>wip</option>
      </select>
    </div>
    <div class="meta-row">
      <label>URL:</label>
      <input v-model="project.url" type="text" placeholder="https://github.com/..." />
    </div>
    <div class="meta-row">
      <label>SORT:</label>
      <input v-model.number="project.sortOrder" type="number" placeholder="0" />
    </div>
  </div>

  <div class="editor-actions">
    <button class="btn" @click="handleSave" :disabled="saving">
      {{ saving ? '[ SAVING... ]' : '[ SAVE ]' }}
    </button>
    <RouterLink to="/admin/projects" class="btn">cd ..</RouterLink>
  </div>
</template>

<style scoped>
.editor-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; gap: 0.5rem; }
.editor-prompt { font-weight: bold; font-size: 0.95rem; }
.muted { color: var(--muted); }
.editor-status { font-size: 0.8rem; font-weight: bold; padding: 2px 8px; border: 1px solid var(--border); color: var(--muted); }
.status-error { background: #ff4444; color: #fff !important; }
.meta-panel { border: 2px solid var(--border); padding: 1.5rem; margin-bottom: 1.5rem; background: var(--bg); box-shadow: var(--shadow); }
.meta-row { display: flex; align-items: center; gap: 10px; margin-bottom: 1rem; }
.meta-row:last-child { margin-bottom: 0; }
.meta-row label { font-weight: bold; font-size: 0.85rem; min-width: 60px; color: var(--muted); }
.meta-row input, .meta-row select { flex: 1; background: var(--bg); color: var(--fg); border: 1px solid var(--border); padding: 6px 10px; font-family: var(--font-main); font-size: 0.9rem; outline: none; }
.meta-row input:focus, .meta-row select:focus { border-width: 2px; }
.editor-actions { display: flex; gap: 10px; align-items: center; border-top: 2px dashed var(--border); padding-top: 1.5rem; }
@media (max-width: 600px) { .meta-row { flex-direction: column; align-items: flex-start; } .meta-row label { min-width: auto; } .meta-row input, .meta-row select { width: 100%; } }
</style>
