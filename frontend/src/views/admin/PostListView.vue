<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { listPostsAdmin, deletePost as apiDeletePost } from '../../api/posts'
import type { PostPayload } from '../../api/posts'
import PageState from '../../components/PageState.vue'
import AdminNav from '../../components/AdminNav.vue'
import TerminalFeedback from '../../components/TerminalFeedback.vue'

const router = useRouter()

const PAGE_SIZE = 5
const page = ref(0)
const posts = ref<PostPayload[]>([])
const total = ref(0)
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

const totalPages = computed(() => Math.ceil(total.value / PAGE_SIZE))

async function fetchPosts() {
  loading.value = true
  error.value = ''
  try {
    const apiPage = page.value + 1
    const data = await listPostsAdmin(apiPage, PAGE_SIZE)
    posts.value = data.posts || []
    if (!posts.value.length && page.value > 0) { page.value--; await fetchPosts(); return }
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
  if (!confirm(`Delete "${slug}.md"?`)) return
  try {
    await apiDeletePost(slug)
    showFeedback(`${slug}.md deleted`, 'ok')
    fetchPosts()
  } catch (e: any) {
    showFeedback(e.message || 'delete failed', 'err')
  }
}

onMounted(fetchPosts)
</script>

<template>
  <AdminNav />

  <div class="admin-prompt">
    root@k8s-node ~/admin $ <span class="muted">ls -la ./posts</span>
  </div>

  <PageState v-if="loading" mode="loading" message="Loading posts…" />
  <PageState v-else-if="error" mode="error" :message="error" retry @retry="fetchPosts" />
  <PageState v-else-if="!posts.length" mode="empty" message="No posts yet. Create your first article below." />

  <table class="post-table" v-else>
    <thead>
      <tr>
        <th>STATUS</th>
        <th>OWNER</th>
        <th>SIZE</th>
        <th>DATE</th>
        <th>FILENAME</th>
        <th>ACTIONS</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="post in posts" :key="post.slug">
        <td data-label="STATUS">{{ post.published ? 'PUBLISHED' : 'DRAFT' }}</td>
        <td data-label="AUTHOR">{{ post.author }}</td>
        <td data-label="WORDS">{{ wordCount(post.content) }}</td>
        <td data-label="DATE">{{ post.createdAt ? post.createdAt.slice(0, 10) : '—' }}</td>
        <td class="post-name" data-label="ARTICLE">
          <RouterLink :to="(post.published ? '/article/' : '/admin/edit/') + post.slug" class="file-link">
            {{ post.title }}
          </RouterLink>
        </td>
        <td class="row-actions" data-label="ACTIONS">
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
.admin-prompt {
  font-weight: bold;
  margin-bottom: 1.5rem;
  font-size: 0.95rem;
}

.muted {
  color: var(--muted);
}

.status-line {
  font-family: var(--font-main);
  font-size: 0.82rem;
  color: var(--muted);
  margin-bottom: 1rem;
}

.cursor {
  display: inline-block;
  width: 8px;
  height: 1em;
  background: var(--fg);
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.post-table {
  width: 100%;
  border: 2px solid var(--border);
  border-collapse: collapse;
  font-family: var(--font-main);
  font-size: 0.85rem;
  box-shadow: var(--shadow);
  margin-bottom: 1.5rem;
}

.post-table th {
  text-align: left;
  padding: 6px 10px;
  border-bottom: 2px solid var(--border);
  background: var(--fg);
  color: var(--bg);
  font-size: 0.75rem;
}

.post-table td {
  padding: 6px 10px;
  border-bottom: 1px solid var(--border);
}

.post-table tr:hover td {
  background: var(--fg);
  color: var(--bg);
}

.file-link {
  color: inherit;
  text-decoration: none;
  font-weight: bold;
}

.file-link:hover {
  text-decoration: underline;
}

.btn-danger {
  color: #ff4444;
  border-color: #ff4444;
}

.btn-danger:hover {
  background: #ff4444;
  color: #fff;
}

.page-nav {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 2px dashed var(--border);
}

.page-info {
  font-size: 0.85rem;
  color: var(--muted);
  min-width: 60px;
  text-align: center;
}

.admin-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

button:disabled {
  opacity: 0.35;
  cursor: default;
}

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
