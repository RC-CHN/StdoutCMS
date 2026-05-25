<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { listPostsAdmin, deletePost as apiDeletePost } from '../../api/posts'
import type { PostPayload } from '../../api/posts'

const router = useRouter()

const PAGE_SIZE = 5
const page = ref(0)
const posts = ref<PostPayload[]>([])
const total = ref(0)
const loading = ref(false)
const error = ref('')

const totalPages = computed(() => Math.ceil(total.value / PAGE_SIZE))

async function fetchPosts() {
  loading.value = true
  error.value = ''
  try {
    const apiPage = page.value + 1
    const data = await listPostsAdmin(apiPage, PAGE_SIZE)
    posts.value = data.posts
    total.value = data.total
  } catch (e: any) {
    error.value = e.message || 'failed to load posts'
  } finally {
    loading.value = false
  }
}

function nextPage() { if (page.value < totalPages.value - 1) { page.value++; fetchPosts() } }
function prevPage() { if (page.value > 0) { page.value--; fetchPosts() } }

function wordCount(content: string) {
  return content.replace(/\s/g, '').length
}

async function deletePost(slug: string) {
  try {
    await apiDeletePost(slug)
    fetchPosts()
  } catch (e: any) {
    window.alert('DELETE failed: ' + (e.message || 'unknown'))
  }
}

onMounted(fetchPosts)
</script>

<template>
  <div class="admin-prompt">
    root@k8s-node ~/admin $ <span class="muted">ls -la ./posts</span>
  </div>

  <div v-if="loading" class="status-line">loading posts...</div>
  <div v-else-if="error" class="status-line" style="color: #ff4444;">ERROR: {{ error }}</div>

  <table class="post-table" v-else>
    <thead>
      <tr>
        <th>PERMISSIONS</th>
        <th>OWNER</th>
        <th>SIZE</th>
        <th>DATE</th>
        <th>FILENAME</th>
        <th>ACTIONS</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="post in posts" :key="post.slug">
        <td>-rw-r--r--</td>
        <td>{{ post.author }}</td>
        <td>{{ wordCount(post.content) }}B</td>
        <td>{{ post.createdAt ? post.createdAt.slice(0, 10) : '—' }}</td>
        <td>
          <RouterLink :to="`/article/${post.slug}`" class="file-link">
            {{ post.slug }}.md
          </RouterLink>
        </td>
        <td>
          <button class="btn btn-sm" @click="router.push(`/admin/edit/${post.slug}`)">EDIT</button>
          <button class="btn btn-sm btn-danger" @click="deletePost(post.slug)">DEL</button>
        </td>
      </tr>
    </tbody>
  </table>

  <!-- 分页 -->
  <nav class="page-nav" v-if="totalPages > 1">
    <button class="btn" :disabled="page === 0" @click="prevPage">&lt;&lt; prev</button>
    <span class="page-info">{{ page + 1 }} / {{ totalPages }}</span>
    <button class="btn" :disabled="page >= totalPages - 1" @click="nextPage">next &gt;&gt;</button>
  </nav>

  <div class="admin-actions">
    <RouterLink to="/admin/edit" class="btn">+ NEW_POST.MD</RouterLink>
    <span class="status-line">{{ total }} file(s) found</span>
  </div>

  <div class="admin-prompt" style="margin-top: 2rem;">
    root@k8s-node ~/admin $<span class="cursor"></span>
  </div>
</template>

<style scoped>
.admin-prompt {
  font-weight: bold;
  margin-bottom: 1.5rem;
  font-size: 0.95rem;
}

.muted {
  color: var(--muted);
}

.post-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 2rem;
  font-size: 0.9rem;
}

.post-table th,
.post-table td {
  border: 1px solid var(--border);
  padding: 8px 12px;
  text-align: left;
}

.post-table th {
  background: var(--fg);
  color: var(--bg);
  font-weight: bold;
  text-transform: uppercase;
  font-size: 0.8rem;
}

.post-table tr:hover td {
  background: rgba(128, 128, 128, 0.05);
}

.file-link {
  color: var(--fg);
  text-decoration: underline;
  text-decoration-style: dashed;
}

.file-link:hover {
  background: var(--fg);
  color: var(--bg);
  text-decoration: none;
}

.btn-sm {
  padding: 4px 12px;
  font-size: 0.8rem;
  margin: 0 5px 0 0;
  display: inline-block;
  text-align: center;
  min-width: 50px;
  line-height: 1.4;
  vertical-align: middle;
  box-sizing: border-box;
}

.btn-danger:hover {
  background: #ff4444;
  color: #fff;
}

/* 分页 */
.page-nav {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin: 1.5rem 0;
  padding-bottom: 1.5rem;
  border-bottom: 2px dashed var(--border);
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

.admin-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 2px dashed var(--border);
  padding-top: 1.5rem;
}

.status-line {
  color: var(--muted);
  font-size: 0.85rem;
}

@media (max-width: 600px) {
  .post-table th,
  .post-table td {
    padding: 6px 8px;
    font-size: 0.8rem;
  }
  .post-table th:nth-child(1),
  .post-table td:nth-child(1),
  .post-table th:nth-child(2),
  .post-table td:nth-child(2) {
    display: none;
  }
}
</style>
