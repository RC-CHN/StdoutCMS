<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { listProjectsAdmin, deleteProject as apiDelete } from '../../api/projects'
import type { ProjectPayload } from '../../api/projects'
import PageState from '../../components/PageState.vue'
import AdminNav from '../../components/AdminNav.vue'
import TerminalFeedback from '../../components/TerminalFeedback.vue'

const router = useRouter()

const projects = ref<ProjectPayload[]>([])
const loading = ref(true)
const error = ref('')

const feedbackMsg = ref<string | null>(null)
const feedbackType = ref<'ok' | 'err'>('ok')
const feedbackTrigger = ref(0)

function showFeedback(msg: string, type: 'ok' | 'err') {
  feedbackMsg.value = msg
  feedbackType.value = type
  feedbackTrigger.value++
}

async function fetch() {
  loading.value = true
  error.value = ''
  try {
    const data = await listProjectsAdmin()
    projects.value = data.projects || []
  } catch (e: any) {
    error.value = e.message || 'failed'
  } finally {
    loading.value = false
  }
}

async function deleteProject(id: number, name: string) {
  if (!confirm(`Delete project "${name}"?`)) return
  try {
    await apiDelete(id)
    showFeedback(`${name} deleted`, 'ok')
    fetch()
  } catch (e: any) {
    showFeedback(e.message || 'delete failed', 'err')
  }
}

onMounted(fetch)
</script>

<template>
  <AdminNav />

  <div class="admin-prompt">
    root@k8s-node ~/admin $ <span class="muted">ls -la ./projects</span>
  </div>

  <PageState v-if="loading" mode="loading" message="Loading projects…" />
  <PageState v-else-if="error" mode="error" :message="error" retry @retry="fetch" />
  <PageState v-else-if="!projects.length" mode="empty" message="No projects yet. Add your first project below." />

  <table class="post-table" v-else>
    <thead>
      <tr>
        <th>PERMISSIONS</th>
        <th>NAME</th>
        <th>LANG</th>
        <th>STATUS</th>
        <th>ACTIONS</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="p in projects" :key="p.id">
        <td class="permissions">-rw-r--r--</td>
        <td class="post-name" data-label="PROJECT">
          <a :href="p.url" class="file-link" target="_blank">{{ p.name }}</a>
        </td>
        <td data-label="LANGUAGE">{{ p.lang }}</td>
        <td data-label="STATUS">{{ p.status }}</td>
        <td class="row-actions" data-label="ACTIONS">
          <button class="btn btn-sm" @click="router.push(`/admin/projects/edit/${p.id}`)">EDIT</button>
          <button class="btn btn-sm btn-danger" @click="deleteProject(p.id, p.name)">DEL</button>
        </td>
      </tr>
    </tbody>
  </table>

  <div class="admin-actions">
    <RouterLink to="/admin/projects/edit" class="btn">+ NEW_PROJECT.SH</RouterLink>
    <span class="status-line">{{ projects.length }} file(s) found</span>
  </div>

  <TerminalFeedback
    :message="feedbackMsg"
    :type="feedbackType"
    :duration="feedbackType === 'err' ? 0 : 4000"
    :trigger="feedbackTrigger"
  />

  <div class="admin-prompt" style="margin-top: 2rem;">
    root@k8s-node ~/admin $<span class="cursor"></span>
  </div>
</template>

<style scoped>
.admin-prompt { font-weight: bold; margin-bottom: 1.5rem; font-size: 0.95rem; }
.muted { color: var(--muted); }
.post-table { width: 100%; border-collapse: collapse; margin-bottom: 2rem; font-size: 0.9rem; }
.post-table th, .post-table td { border: 1px solid var(--border); padding: 8px 12px; text-align: left; }
.post-table th { background: var(--fg); color: var(--bg); font-weight: bold; text-transform: uppercase; font-size: 0.8rem; }
.post-table tr:hover td { background: rgba(128, 128, 128, 0.05); }
.file-link { color: var(--fg); text-decoration: underline; text-decoration-style: dashed; }
.file-link:hover { background: var(--fg); color: var(--bg); text-decoration: none; }
.btn-sm { padding: 4px 12px; font-size: 0.8rem; margin: 0 5px 0 0; display: inline-block; text-align: center; min-width: 50px; line-height: 1.4; vertical-align: middle; box-sizing: border-box; }
.btn-danger:hover { background: #ff4444; color: #fff; }
.admin-actions { display: flex; justify-content: space-between; align-items: center; border-top: 2px dashed var(--border); padding-top: 1.5rem; }
.status-line { color: var(--muted); font-size: 0.85rem; }
@media (max-width: 600px) { .post-table th, .post-table td { padding: 6px 8px; font-size: 0.8rem; } }

.post-table { table-layout: fixed; background: var(--bg); }
.post-table td { overflow-wrap: anywhere; }
.row-actions { white-space: normal; }
.row-actions button { margin: 3px; }
.admin-prompt { overflow-wrap: anywhere; }
@media (max-width: 768px) {
  .post-table, .post-table tbody { display: block; width: 100%; border: 0; box-shadow: none; }
  .post-table thead { display: none; }
  .post-table tr { display: grid; grid-template-columns: 1fr 1fr; border: 1px solid var(--border); margin-bottom: 1rem; padding: 0.5rem; }
  .post-table td { display: flex; flex-direction: column; border: 0; padding: 0.5rem; min-width: 0; }
  .post-table td::before { content: attr(data-label); color: var(--muted); font-size: 0.65rem; }
  .post-table .post-name { grid-column: 1 / -1; grid-row: 1; font-size: 1rem; }
  .post-table .permissions { display: none; }
  .post-table .row-actions { grid-column: 1 / -1; flex-direction: row; gap: 0.5rem; }
  .row-actions::before { display: none; }
  .row-actions button { flex: 1; min-height: 40px; margin: 0; }
  .admin-actions { gap: 0.75rem; flex-wrap: wrap; }
}
</style>
