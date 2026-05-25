<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { listPosts } from '../api/posts'
import type { PostPayload } from '../api/posts'
import PostCard from '../components/PostCard.vue'

const PAGE_SIZE = 3
const page = ref(0)
const totalPosts = ref(0)
const allPosts = ref<PostPayload[]>([])
const loading = ref(false)
const error = ref('')

async function fetchPosts() {
  loading.value = true
  error.value = ''
  try {
    const apiPage = page.value + 1
    const data = await listPosts(apiPage, PAGE_SIZE)
    allPosts.value = data.posts
    totalPosts.value = data.total
  } catch (e: any) {
    error.value = e.message || 'failed to load posts'
  } finally {
    loading.value = false
  }
}

const totalPages = computed(() => Math.ceil(totalPosts.value / PAGE_SIZE))

function nextPage() { if (page.value < totalPages.value - 1) { page.value++; fetchPosts() } }
function prevPage() { if (page.value > 0) { page.value--; fetchPosts() } }

onMounted(fetchPosts)
</script>

<template>
  <div v-if="loading" class="prompt-line muted">loading articles...</div>
  <div v-else-if="error" class="prompt-line" style="color: #ff4444;">ERROR: {{ error }}</div>
  <template v-else>
    <PostCard v-for="post in allPosts" :key="post.slug" :post="post" />

    <nav class="page-nav" v-if="totalPages > 1">
      <button class="btn" :disabled="page === 0" @click="prevPage">
        &lt;&lt; prev
      </button>
      <span class="page-info">{{ page + 1 }} / {{ totalPages }}</span>
      <button class="btn" :disabled="page >= totalPages - 1" @click="nextPage">
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
