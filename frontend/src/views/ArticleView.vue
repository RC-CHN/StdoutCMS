<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { getPostBySlug, posts } from '../mocks/posts'
import ArticleRenderer from '../components/ArticleRenderer.vue'

const route = useRoute()
const slug = route.params.slug as string
const post = computed(() => getPostBySlug(slug))

const currentIndex = computed(() => posts.findIndex((p) => p.slug === slug))
const prevPost = computed(() => (currentIndex.value > 0 ? posts[currentIndex.value - 1] : null))
const nextPost = computed(() => (currentIndex.value < posts.length - 1 ? posts[currentIndex.value + 1] : null))
</script>

<template>
  <template v-if="post">
    <article>
      <div class="article-header">
        <div class="article-meta">
          [AUTHOR: {{ post.author }}] [DATE: {{ post.date }}] [WORDS: {{ post.wordCount }}]
        </div>
        <h1 class="article-title">{{ post.title }}</h1>
      </div>

      <ArticleRenderer :content="post.content" />

      <div class="eof-marker">EOF</div>
    </article>

    <nav class="post-nav">
      <RouterLink v-if="prevPost" :to="`/article/${prevPost.slug}`" class="btn">
        &lt;&lt; {{ prevPost.slug }}.md
      </RouterLink>
      <span v-else class="btn disabled">[ TOP_OF_DIR ]</span>

      <RouterLink v-if="nextPost" :to="`/article/${nextPost.slug}`" class="btn">
        {{ nextPost.slug }}.md &gt;&gt;
      </RouterLink>
      <span v-else class="btn disabled">[ END_OF_DIR ]</span>
    </nav>
  </template>
  <template v-else>
    <div class="prompt-line">
      guest@k8s-node ~/articles $ <span style="color: var(--muted);">cat {{ slug }}.md</span>
    </div>
    <p style="color: var(--muted); margin-top: 2rem;">ERROR: File not found. Use the listing above to see available files.</p>
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
