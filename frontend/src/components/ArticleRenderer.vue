<script setup lang="ts">
import { computed } from 'vue'
import { parseMarkdown } from '../utils/md'

const props = defineProps<{
  content: string
}>()

const html = computed(() => parseMarkdown(props.content))
</script>

<template>
  <div class="article-content" v-html="html" />
</template>

<style scoped>
.article-content {
  font-size: 1.05rem;
}

.article-content :deep(h2) {
  margin-top: 2.5rem;
  margin-bottom: 1rem;
  color: var(--fg);
}
.article-content :deep(h2::before) {
  content: ">> ";
  color: var(--muted);
}

.article-content :deep(h3) {
  margin-top: 2.5rem;
  margin-bottom: 1rem;
  color: var(--fg);
}
.article-content :deep(h3::before) {
  content: "# ";
  color: var(--muted);
}

.article-content :deep(p) {
  margin-bottom: 1.5rem;
}

.article-content :deep(a) {
  color: var(--fg);
  text-decoration: underline;
  text-decoration-style: dashed;
  text-underline-offset: 4px;
}
.article-content :deep(a:hover) {
  background-color: var(--fg);
  color: var(--bg);
  text-decoration: none;
}

.article-content :deep(ul) {
  margin-bottom: 1.5rem;
  padding-left: 2rem;
  list-style-type: square;
}
.article-content :deep(li) {
  margin-bottom: 0.5rem;
}

.article-content :deep(blockquote) {
  border-left: 4px solid var(--border);
  padding: 1rem 1rem 1rem 1.5rem;
  margin: 2rem 0;
  color: var(--muted);
  background: rgba(128, 128, 128, 0.05);
  font-style: italic;
}

.article-content :deep(code) {
  background: var(--fg);
  color: var(--bg);
  padding: 2px 6px;
  font-size: 0.9em;
  font-weight: bold;
}

.article-content :deep(pre) {
  background: var(--bg);
  border: 2px solid var(--border);
  padding: 1rem;
  overflow-x: auto;
  margin: 2rem 0;
  box-shadow: var(--shadow);
  position: relative;
}

.article-content :deep(.code-lang) {
  border-bottom: 2px dashed var(--border);
  padding-bottom: 5px;
  margin-bottom: 10px;
  font-size: 0.8rem;
  color: var(--muted);
  text-transform: uppercase;
}

.article-content :deep(pre code) {
  background: transparent;
  color: var(--fg);
  padding: 0;
  font-weight: normal;
}

.article-content :deep(strong) {
  color: var(--fg);
}

.article-content :deep(.md-image) {
  display: block;
  margin: 2rem auto;
  max-width: 80%;
  border: 2px solid var(--border);
  box-shadow: var(--shadow);
  background: var(--bg);
}

.article-content :deep(.md-image-bar) {
  border-bottom: 2px dashed var(--border);
  padding: 5px 10px;
  font-size: 0.8rem;
  color: var(--muted);
  font-weight: bold;
  text-transform: uppercase;
}

.article-content :deep(.md-image img) {
  display: block;
  max-width: 100%;
  height: auto;
  filter: grayscale(1);
  transition: filter 0.3s;
}

.article-content :deep(.md-image img:hover) {
  filter: grayscale(0);
}
</style>
