<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { listPostSummaries, type PostSummary } from '../api/posts'
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
const allPosts = ref<PostSummary[]>([])
const loading = ref(true)
let requestId = 0
const error = ref('')

async function fetchPosts() {
  const id = ++requestId
  loading.value = true
  error.value = ''
  try {
    const data = await listPostSummaries(page.value, PAGE_SIZE)
    if (id !== requestId) return
    const lastPage = Math.max(1, Math.ceil(data.total / PAGE_SIZE))
    if (page.value > lastPage) { await router.replace({ query: { page: String(lastPage) } }); return }
    allPosts.value = data.posts
    totalPosts.value = data.total
  } catch {
    if (id !== requestId) return
    error.value = 'Could not load articles.'
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
  <PageState v-else-if="!allPosts.length" mode="empty" message="No articles published yet." />
  <template v-else>
    <PostCard v-for="post in allPosts" :key="post.slug" :post="post" />

    <nav class="page-nav" v-if="totalPages > 1" aria-label="Article pages">
      <button class="btn" :disabled="page === 1" @click="goToPage(page - 1)">
        ← PREVIOUS
      </button>
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <button class="btn" :disabled="page >= totalPages" @click="goToPage(page + 1)">
        NEXT →
      </button>
    </nav>
  </template>
</template>

<style scoped>
.page-nav {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin: 1.5rem 0 2rem;
}

.page-info {
  font-size: 0.85rem;
  color: var(--muted);
  min-width: 60px;
  text-align: center;
}
</style>
