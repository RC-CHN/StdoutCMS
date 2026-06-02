<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPost } from '../api/posts'
import type { PostPayload } from '../api/posts'
import ArticleRenderer from '../components/ArticleRenderer.vue'

const post = ref<PostPayload | null>(null)
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  loading.value = true
  try {
    post.value = await getPost('about')
  } catch {
    error.value = 'about.md not found'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="prompt-line">
    guest@k8s-node ~ %  <span style="color: var(--muted);">cat about.md</span>
  </div>

  <div v-if="loading" class="status-line">cat about.md...</div>
  <div v-else-if="error" class="status-line" style="color: #ff4444;">ERROR: {{ error }}</div>

  <template v-else-if="post">
    <div class="about-header">
      <h1 class="about-title">{{ post.title }}</h1>
    </div>
    <article class="article-content">
      <ArticleRenderer :content="post.content" />
    </article>
    <div class="eof-marker">EOF</div>
  </template>
</template>

<style scoped>
.prompt-line { margin-bottom: 2rem; font-weight: bold; }
.status-line { color: var(--muted); font-size: 0.85rem; }

/* ---- ASCII box header ---- */
.about-header {
  position: relative;
  border: 4px double var(--border);
  padding: 1.8rem 2rem 1.5rem;
  margin-bottom: 2.5rem;
  text-align: center;
  box-shadow: 6px 6px 0 var(--border);
  background: var(--bg);
}

.about-header::before {
  content: "[ ABOUT.TXT ]";
  position: absolute;
  top: -0.65rem;
  left: 1.5rem;
  background: var(--bg);
  padding: 0 8px;
  font-size: 0.7rem;
  font-weight: bold;
  color: var(--muted);
  letter-spacing: 1px;
}

.about-title {
  font-size: 1.8rem;
  font-weight: bold;
  margin: 0;
  letter-spacing: 2px;
  color: var(--fg);
}

.article-content { font-size: 1.05rem; }
.article-content :deep(h2) { margin-top: 2.5rem; margin-bottom: 1rem; color: var(--fg); }
.article-content :deep(h2::before) { content: ">> "; color: var(--muted); }
.article-content :deep(p) { margin-bottom: 1.5rem; }
.article-content :deep(ul) { margin-bottom: 1.5rem; padding-left: 2rem; list-style-type: square; }
.article-content :deep(li) { margin-bottom: 0.5rem; }
.article-content :deep(code) { background: var(--fg); color: var(--bg); padding: 2px 6px; font-weight: bold; }
.article-content :deep(blockquote) { border-left: 4px solid var(--border); padding: 1rem 1.5rem; margin: 2rem 0; color: var(--muted); background: rgba(128,128,128,0.05); font-style: italic; }
.article-content :deep(a) { color: var(--fg); text-decoration: underline; text-decoration-style: dashed; }
.article-content :deep(a:hover) { background-color: var(--fg); color: var(--bg); text-decoration: none; }
.eof-marker { text-align: center; margin: 4rem 0 2rem 0; font-weight: bold; color: var(--muted); letter-spacing: 5px; }
.eof-marker::before { content: "--- [ "; }
.eof-marker::after { content: " ] ---"; }
</style>
