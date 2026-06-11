<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getPost } from '../api/posts'
import type { PostPayload } from '../api/posts'
import ArticleRenderer from '../components/ArticleRenderer.vue'
import ShareQR from '../components/ShareQR.vue'

const route = useRoute()
const slug = route.params.slug as string
const post = ref<PostPayload | null>(null)
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  loading.value = true
  try {
    post.value = await getPost(slug)
  } catch (e: any) {
    error.value = e.message || 'article not found'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="prompt-line muted">cat {{ slug }}.md...</div>
  <div v-else-if="error" class="prompt-line" style="color: #ff4444;">ERROR: {{ error }}</div>
  <template v-else-if="post">
    <article>
      <div class="article-header">
        <div class="article-meta">
          [AUTHOR: {{ post.author }}] [DATE: {{ post.createdAt?.slice(0, 10) || '—' }}] [WORDS: {{ post.wordCount }}]
        </div>
        <h1 class="article-title">{{ post.title }}</h1>
      </div>

      <ArticleRenderer :content="post.content" />

      <div class="eof-marker">EOF</div>

      <div class="share-row">
        <ShareQR
          :slug="post.slug"
          :title="post.title"
          :tags="post.tags"
          :read-time="post.readTime"
          :word-count="post.wordCount"
          :created-at="post.createdAt"
        />
      </div>
    </article>
  </template>
  <template v-else>
    <div class="prompt-line">
      guest@k8s-node ~/articles $ <span style="color: var(--muted);">cat {{ slug }}.md</span>
    </div>
    <p style="color: var(--muted); margin-top: 2rem;">ERROR: File not found.</p>
  </template>
</template>

<style scoped>
.article-header {
  margin-bottom: 3rem;
}

.article-meta {
  font-size: 0.85rem;
  color: var(--muted);
  margin-bottom: 1rem;
  padding: 10px;
  border: 1px dashed var(--muted);
  display: inline-block;
  background: rgba(128, 128, 128, 0.05);
}

.article-title {
  font-size: 2.2rem;
  line-height: 1.2;
  margin-bottom: 1.5rem;
}

.eof-marker {
  text-align: center;
  margin: 4rem 0 1.5rem 0;
  font-weight: bold;
  color: var(--muted);
  letter-spacing: 5px;
}
.eof-marker::before { content: "--- [ "; }
.eof-marker::after { content: " ] ---"; }

.post-nav {
  display: flex;
  justify-content: space-between;
  border-top: 2px dashed var(--border);
  padding-top: 2rem;
}

.btn.disabled {
  opacity: 0.35;
  cursor: default;
}
.btn.disabled:hover {
  background: var(--bg);
  color: var(--fg);
}

.prompt-line {
  font-weight: bold;
}

@media (max-width: 600px) {
  .article-title { font-size: 1.8rem; }
  .post-nav { flex-direction: column; gap: 1rem; }
  .post-nav .btn { width: 100%; text-align: center; }
}
</style>
