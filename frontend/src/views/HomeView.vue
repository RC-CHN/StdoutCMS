<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { listPosts } from '../api/posts'
import type { PostPayload } from '../api/posts'
import { useRoute, useRouter } from 'vue-router'
import PageState from '../components/PageState.vue'
import PostCard from '../components/PostCard.vue'

const PAGE_SIZE = 3
const route = useRoute()
const router = useRouter()
const page = computed(() => {
  const value = Number(route.query.page)
  return Number.isSafeInteger(value) && value > 0 ? value : 1
})
const totalPosts = ref(0)
const allPosts = ref<PostPayload[]>([])
const loading = ref(true)
let requestId = 0
const error = ref('')

async function fetchPosts() {
  const id = ++requestId
  loading.value = true
  error.value = ''
  try {
    const data = await listPosts(page.value, PAGE_SIZE)
    if (id !== requestId) return
    const lastPage = Math.max(1, Math.ceil(data.total / PAGE_SIZE))
    if (page.value > lastPage) { await router.replace({ query: { page: String(lastPage) } }); return }
    allPosts.value = data.posts
    totalPosts.value = data.total
  } catch (e: any) {
    if (id !== requestId) return
    error.value = e.message || 'failed to load posts'
  } finally {
    if (id === requestId) loading.value = false
  }
}

const totalPages = computed(() => Math.ceil(totalPosts.value / PAGE_SIZE))

function goToPage(value: number) {
  router.push({ query: { ...route.query, page: String(value) } })
}
watch(page, fetchPosts, { immediate: true })
</script>

<template>
  <PageState v-if="loading" mode="loading" message="Loading articles…" />
  <PageState v-else-if="error" mode="error" :message="error" retry @retry="fetchPosts" />
  <PageState v-else-if="!allPosts.length" mode="empty" message="No articles published yet. Check back soon." />
  <template v-else>
    <PostCard v-for="post in allPosts" :key="post.slug" :post="post" />

    <nav class="page-nav" v-if="totalPages > 1">
      <button class="btn" :disabled="page === 1" @click="goToPage(page - 1)">
        &lt;&lt; prev
      </button>
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <button class="btn" :disabled="page >= totalPages" @click="goToPage(page + 1)">
        next &gt;&gt;
      </button>
    </nav>
  </template>

  <div class="prompt-line">
    guest@k8s-node ~/articles<span class="cursor"></span>
  </div>
</template>

<style scoped>
.page-nav {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin: 2rem 0;
  padding-top: 1.5rem;
  border-top: 2px dashed var(--border);
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

.prompt-line {
  font-weight: bold;
  margin-top: 1rem;
}
</style>
