<script setup lang="ts">
import { ref, computed } from 'vue'
import { posts } from '../mocks/posts'
import PostCard from '../components/PostCard.vue'

const PAGE_SIZE = 3
const page = ref(0)
const totalPages = computed(() => Math.ceil(posts.length / PAGE_SIZE))
const paged = computed(() => {
  const start = page.value * PAGE_SIZE
  return posts.slice(start, start + PAGE_SIZE)
})

function nextPage() { if (page.value < totalPages.value - 1) page.value++ }
function prevPage() { if (page.value > 0) page.value-- }
</script>

<template>
  <PostCard v-for="post in paged" :key="post.slug" :post="post" />

  <nav class="page-nav" v-if="totalPages > 1">
    <button class="btn" :disabled="page === 0" @click="prevPage">
      &lt;&lt; prev
    </button>
    <span class="page-info">{{ page + 1 }} / {{ totalPages }}</span>
    <button class="btn" :disabled="page >= totalPages - 1" @click="nextPage">
      next &gt;&gt;
    </button>
  </nav>

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
