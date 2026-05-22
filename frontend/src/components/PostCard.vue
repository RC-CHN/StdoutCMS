<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { Post } from '../mocks/posts'

defineProps<{
  post: Post
}>()
</script>

<template>
  <article class="post-card">
    <div class="post-meta">
      <span>DATE: {{ post.date }}</span>
      <span>TAGS: [{{ post.tags.join(', ') }}]</span>
    </div>
    <RouterLink :to="`/article/${post.slug}`" class="post-title">{{ post.title }}</RouterLink>
    <div class="post-excerpt">{{ post.excerpt }}</div>
    <RouterLink :to="`/article/${post.slug}`" class="btn">READ_MORE.EXE</RouterLink>
  </article>
</template>

<style scoped>
.post-card {
  border: 2px solid var(--border);
  background: var(--bg);
  padding: 1.5rem;
  margin-bottom: 2rem;
  box-shadow: var(--shadow);
  transition: transform 0.2s, box-shadow 0.2s;
  position: relative;
}

.post-card:hover {
  transform: translate(-2px, -2px);
  box-shadow: 7px 7px 0px var(--border);
}

.post-card::before {
  content: "";
  position: absolute;
  top: 5px;
  right: 5px;
  width: 15px;
  height: 15px;
  border-top: 2px solid var(--border);
  border-right: 2px solid var(--border);
}

.post-meta {
  font-size: 0.85rem;
  color: var(--muted);
  margin-bottom: 0.8rem;
  display: flex;
  gap: 15px;
  border-bottom: 1px dotted var(--muted);
  padding-bottom: 5px;
}

.post-title {
  font-size: 1.4rem;
  color: var(--fg);
  text-decoration: none;
  display: block;
  margin-bottom: 0.5rem;
  font-weight: bold;
}

.post-title:hover {
  text-decoration: underline;
}

.post-excerpt {
  color: var(--muted);
  margin-bottom: 1rem;
}

@media (max-width: 600px) {
  .post-card { padding: 1rem; }
}
</style>
