<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { parseMarkdown } from '../utils/md'

const props = defineProps<{
  content: string
}>()

const html = computed(() => parseMarkdown(props.content))

// lightbox
const lbOpen = ref(false)
const lbSrc = ref('')
const lbAlt = ref('')

function openLightbox(img: HTMLImageElement) {
  lbSrc.value = img.src
  lbAlt.value = img.alt || 'image'
  lbOpen.value = true
}

function closeLightbox() {
  lbOpen.value = false
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closeLightbox()
}

function onClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  const img = target.closest('.md-image img') as HTMLImageElement | null
  if (img) openLightbox(img)
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <div class="article-content" v-html="html" @click="onClick" />

  <teleport to="body">
    <div v-if="lbOpen" class="lb-backdrop" @click.self="closeLightbox">
      <div class="lb-frame">
        <div class="lb-bar">
          <span>> viewing: {{ lbAlt }}</span>
          <button class="lb-close" @click="closeLightbox">[X]</button>
        </div>
        <img :src="lbSrc" :alt="lbAlt" class="lb-image" />
      </div>
    </div>
  </teleport>
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

.article-content :deep(.md-image img) {
  cursor: pointer;
}

/* ---- lightbox ---- */

.lb-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: lb-fade-in 0.15s ease;
}

@keyframes lb-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.lb-frame {
  max-width: 90vw;
  max-height: 90vh;
  border: 2px solid var(--fg);
  box-shadow: 8px 8px 0px var(--fg);
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

.lb-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  border-bottom: 2px dashed var(--border);
  font-size: 0.8rem;
  font-weight: bold;
  color: var(--muted);
}

.lb-close {
  background: none;
  border: none;
  color: var(--fg);
  font-family: var(--font-main);
  font-weight: bold;
  cursor: pointer;
  padding: 2px 6px;
}

.lb-close:hover {
  background: var(--fg);
  color: var(--bg);
}

.lb-image {
  display: block;
  max-width: 90vw;
  max-height: calc(90vh - 40px);
  object-fit: contain;
  filter: none;
}
</style>
