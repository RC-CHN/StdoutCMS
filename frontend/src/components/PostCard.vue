<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { PostSummary } from '../api/posts'

defineProps<{
  post: PostSummary
}>()
</script>

<template>
  <article class="post-card">
    <h2 class="post-heading">
      <RouterLink :to="`/article/${post.slug}`" class="post-title">{{ post.title }}</RouterLink>
    </h2>
    <p v-if="post.excerpt" class="post-excerpt">{{ post.excerpt }}</p>
    <div class="post-meta">
      <time v-if="post.createdAt" :datetime="post.createdAt">{{ post.createdAt.slice(0, 10) }}</time>
      <span v-if="post.tags?.length">{{ post.tags.join(' / ') }}</span>
      <RouterLink :to="`/article/${post.slug}`" class="read-link command" :aria-label="`Read ${post.title}`">./read.elf</RouterLink>
    </div>
  </article>
</template>

<style scoped>
.post-card {
  padding: 0.75rem 0 1rem;
  border-bottom: 1px solid var(--border);
  overflow-wrap: anywhere;
}

.post-card + .post-card { margin-top: 0.75rem; }

.post-heading {
  font-size: clamp(1.3rem, 2vw, 1.6rem);
  line-height: 1.35;
}

.post-title {
  color: var(--fg);
  text-decoration: none;
}

.post-excerpt {
  margin-top: 0.45rem;
  font-family: var(--font-reading);
  color: var(--fg);
}

.post-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.25rem 1rem;
  margin-top: 0.6rem;
  font-size: 0.82rem;
  color: var(--muted);
}

.post-meta time, .read-link { white-space: nowrap; }

.read-link {
  display: inline-flex;
  align-items: center;
  min-height: 2rem;
  margin-inline-start: auto;
  color: var(--fg);
  text-underline-offset: 0.2em;
}

.post-title:is(:hover, :focus-visible),
.read-link:is(:hover, :focus-visible) {
  color: var(--accent);
  text-decoration: underline;
  text-underline-offset: 0.2em;
}

@media (max-width: 600px) {
  .post-title {
    display: inline-block;
    max-width: 100%;
    padding-block: 0.5rem;
  }
  .read-link { min-height: 44px; }
}
</style>
