<script setup lang="ts">
import { ref, onMounted, watch, nextTick, defineAsyncComponent } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getPost } from '../api/posts'
import type { PostPayload } from '../api/posts'
import ArticleRenderer from '../components/ArticleRenderer.vue'
import MobileArticleRenderer from '../components/MobileArticleRenderer.vue'
import PageState from '../components/PageState.vue'
import { fetchMeta } from '../api/meta'
import { useBreakpoint } from '../composables/useBreakpoint'

const ShareQR = defineAsyncComponent(() => import('../components/ShareQR.vue'))
const MobileChatWidget = defineAsyncComponent(() => import('../components/MobileChatWidget.vue'))
const route = useRoute()
const router = useRouter()
const slug = route.params.slug as string
const post = ref<PostPayload | null>(null)
const loading = ref(true)
const error = ref('')
const aiEnabled = ref(false)
const { isMobile } = useBreakpoint()
const contentEl = ref<HTMLElement | null>(null)
const headings = ref<{ id: string; text: string; nested: boolean }[]>([])
const showQR = ref(false)
const shareStatus = ref('')
const back = typeof router.options.history.state.back === 'string' ? router.options.history.state.back : ''
const listPath = back && /^\/(?:\?[^#]*)?$/.test(back) ? back : '/'

async function loadPost() {
  loading.value = true
  error.value = ''
  try { post.value = await getPost(slug) }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Could not load this article.' }
  finally { loading.value = false }
}
watch([post, isMobile, loading], async () => {
  await nextTick()
  headings.value = Array.from(contentEl.value?.querySelectorAll('h2[id], h3[id]') || []).map(el => ({
    id: el.id, text: el.textContent || '', nested: el.tagName === 'H3',
  }))
}, { flush: 'post' })
async function shareArticle() {
  try {
    if (navigator.share) await navigator.share({ title: post.value?.title, url: location.href })
    else {
      await navigator.clipboard.writeText(location.href)
      shareStatus.value = 'Link copied.'
    }
  } catch (cause) {
    if (cause instanceof Error && cause.name === 'AbortError') return
    shareStatus.value = 'Could not share. Copy the address from your browser, or use the QR card.'
  }
}
function backToList() {
  if (back === listPath) router.back()
  else router.push(listPath)
}
onMounted(() => {
  void loadPost()
  fetchMeta().then(meta => { aiEnabled.value = meta.ai }).catch(() => {})
})
</script>

<template>
  <div class="reading-page">
    <PageState v-if="loading" mode="loading" message="Loading article…" />
    <PageState v-else-if="error" mode="error" :message="error" retry @retry="loadPost" />
    <article v-else-if="post">
      <button class="back-link" @click="backToList">← ARTICLES</button>
      <header class="article-header">
        <div class="article-meta">
          <span>{{ post.author }}</span>
          <time :datetime="post.createdAt">{{ post.createdAt?.slice(0, 10) || '—' }}</time>
          <span v-if="post.readTime">{{ post.readTime }} READ</span>
        </div>
        <h1 class="article-title">{{ post.title }}</h1>
        <div v-if="post.tags?.length" class="article-tags">{{ post.tags.join(' / ') }}</div>
      </header>
      <details v-if="headings.length > 1" class="article-toc">
        <summary>CONTENTS <span>{{ headings.length }} sections</span></summary>
        <ol>
          <li v-for="heading in headings" :key="heading.id" :class="{ nested: heading.nested }">
            <RouterLink replace :to="{ path: route.path, hash: '#' + heading.id }">{{ heading.text }}</RouterLink>
          </li>
        </ol>
      </details>
      <div ref="contentEl">
        <MobileArticleRenderer v-if="isMobile" :content="post.content" />
        <ArticleRenderer v-else :content="post.content" />
      </div>
      <div class="eof-marker">— EOF —</div>
      <div class="article-actions">
        <button @click="backToList">← ARTICLES</button>
        <button @click="shareArticle">SHARE / COPY LINK</button>
        <button @click="showQR = true">QR CARD</button>
      </div>
      <p class="share-status" role="status">{{ shareStatus }}</p>
      <ShareQR v-if="showQR" :slug="post.slug" :title="post.title" :tags="post.tags" :read-time="post.readTime"
        :word-count="post.wordCount" :created-at="post.createdAt" @close="showQR = false" />
      <MobileChatWidget v-if="aiEnabled && isMobile" />
    </article>
  </div>
</template>

<style scoped>
.back-link { border: 0; padding: 0.5rem 0; margin-bottom: 1rem; font-size: 0.75rem; color: var(--muted); }
.article-header { margin-bottom: 2rem; }
.article-meta { display: flex; flex-wrap: wrap; gap: 0.4rem 1rem; color: var(--muted); font-size: 0.75rem; margin-bottom: 0.75rem; }
.article-meta time { white-space: nowrap; }
.article-title { font-size: clamp(1.6rem, 3vw, 2.2rem); line-height: 1.4; overflow-wrap: anywhere; }
.article-tags { color: var(--muted); font-size: 0.75rem; margin-top: 0.75rem; }
.article-toc { border-block: 1px solid var(--border); padding: 0.8rem 0; margin-bottom: 2rem; font-size: 0.85rem; }
.article-toc summary { font-weight: bold; }
.article-toc summary span { color: var(--muted); font-size: 0.75rem; margin-left: 0.75rem; }
.article-toc ol { list-style: none; margin-top: 0.75rem; }
.article-toc a { display: block; padding: 0.4rem 0; color: var(--fg); text-decoration-style: dashed; overflow-wrap: anywhere; }
.article-toc .nested { padding-left: 1rem; }
.eof-marker { text-align: center; color: var(--muted); margin-block: 3rem 1.5rem; }
.article-actions { display: flex; flex-wrap: wrap; gap: 0.5rem; border-top: 1px solid var(--border); padding-top: 1rem; }
.article-actions button { font-size: 0.75rem; min-height: 40px; }
.share-status { font-size: 0.8rem; color: var(--muted); margin-top: 0.75rem; }
</style>
