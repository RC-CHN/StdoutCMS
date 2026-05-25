<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listProjects } from '../api/projects'
import type { ProjectPayload } from '../api/projects'

const projects = ref<ProjectPayload[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const data = await listProjects()
    projects.value = data.projects
  } catch {
    // keep default empty
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <h1 class="projects-title">> projects/__init__.py</h1>

  <div v-if="loading" class="status-line">loading projects...</div>

  <div class="project-card" v-for="p in projects" :key="p.name">
    <h2>{{ p.name }}</h2>
    <p>{{ p.description }}</p>
    <div class="project-meta">
      <span>LANG: {{ p.lang }}</span>
      <span>STATUS: {{ p.status }}</span>
    </div>
    <a :href="p.url" class="btn" target="_blank">VIEW_REPO.EXE</a>
  </div>

  <div class="eof-marker">EOF</div>
</template>

<style scoped>
.projects-title {
  font-size: 1.8rem;
  margin-bottom: 2rem;
  color: var(--fg);
}

.project-card {
  border: 2px solid var(--border);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  box-shadow: var(--shadow);
  background: var(--bg);
}

.project-card h2 {
  margin-top: 0;
  margin-bottom: 0.5rem;
  font-size: 1.2rem;
  color: var(--fg);
}

.project-card h2::before { content: "> "; color: var(--muted); }

.project-card p {
  margin-bottom: 1rem;
  color: var(--muted);
}

.project-meta {
  font-size: 0.85rem;
  color: var(--muted);
  margin-bottom: 1rem;
  display: flex;
  gap: 15px;
  border-bottom: 1px dotted var(--muted);
  padding-bottom: 8px;
}

.eof-marker {
  text-align: center;
  margin: 3rem 0 1.5rem 0;
  font-weight: bold;
  color: var(--muted);
  letter-spacing: 5px;
}
.eof-marker::before { content: "--- [ "; }
.eof-marker::after { content: " ] ---"; }

.page-nav {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  border-top: 2px dashed var(--border);
  padding-top: 2rem;
}

.page-info {
  font-size: 0.85rem;
  color: var(--muted);
  min-width: 60px;
  text-align: center;
}

button:disabled {
  opacity: 0.35;
  cursor: default;
}
button:disabled:hover {
  background: var(--bg);
  color: var(--fg);
}
</style>
