<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAbout } from '../api/about'
import type { AboutPayload } from '../api/about'
import PageState from '../components/PageState.vue'
import MobileArticleRenderer from '../components/MobileArticleRenderer.vue'
import { useBreakpoint } from '../composables/useBreakpoint'
import ArticleRenderer from '../components/ArticleRenderer.vue'

const about = ref<AboutPayload | null>(null)
const loading = ref(true)
const { isMobile } = useBreakpoint()
const error = ref('')

async function loadAbout() {
  loading.value = true
  error.value = ''
  try {
    about.value = await getAbout()
  } catch {
    error.value = 'Could not load the About page. Please try again.'
  } finally {
    loading.value = false
  }
}
onMounted(loadAbout)
</script>

<template>
  <div class="reading-page">
  <div class="prompt-line">
    guest@k8s-node ~ %  <span style="color: var(--muted);">cat about.md</span>
  </div>

  <PageState v-if="loading" mode="loading" message="Loading About…" />
  <PageState v-else-if="error" mode="error" :message="error" retry @retry="loadAbout" />

  <template v-else-if="about">
    <div class="about-header">
      <h1 class="about-title">{{ about.title }}</h1>
    </div>
    <article>
      <MobileArticleRenderer v-if="isMobile" :content="about.content" />
      <ArticleRenderer v-else :content="about.content" />
    </article>
    <div class="eof-marker">EOF</div>
  </template>
  </div>
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

.eof-marker { text-align: center; margin: 4rem 0 2rem 0; font-weight: bold; color: var(--muted); letter-spacing: 5px; }
.eof-marker::before { content: "--- [ "; }
.eof-marker::after { content: " ] ---"; }
</style>
