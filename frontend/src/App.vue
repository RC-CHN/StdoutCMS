<script setup lang="ts">
import { onMounted, onUnmounted, computed, ref, watch, defineAsyncComponent, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import TerminalHeader from './components/TerminalHeader.vue'
import TerminalFooter from './components/TerminalFooter.vue'
import FileListing from './components/FileListing.vue'
import { useSiteMeta } from './composables/useSiteMeta'
import { listPostSummaries, type PostSummary } from './api/posts'
import { useBreakpoint } from './composables/useBreakpoint'

const ChatWidget = defineAsyncComponent(() => import('./components/ChatWidget.vue'))

const route = useRoute()
const { isMobile } = useBreakpoint()

const meta = useSiteMeta()
const sidebarPosts = ref<PostSummary[]>()
let loadingSidebar = false

/* 侧边栏 toggle */
const sidebarOpen = ref(false)
function toggleSidebar() { sidebarOpen.value = !sidebarOpen.value }
function closeSidebar() { sidebarOpen.value = false }

/* 侧边栏目录上下文 */
const promptPath = computed(() => {
  if (route.meta.requiresAuth) return 'root@stdoutcms:~/admin'
  if (route.name === 'article') return 'guest@stdoutcms:~/articles'
  if (route.name === 'about') return 'guest@stdoutcms:~'
  if (route.name === 'projects') return 'guest@stdoutcms:~/projects'
  return 'guest@stdoutcms:~'
})

const listingMode = computed<'root' | 'articles'>(() =>
  route.name === 'article' ? 'articles' : 'root'
)

const activeArticleSlug = computed(() =>
  route.name === 'article' ? (route.params.slug as string) : undefined
)

onMounted(() => {
  watch([activeArticleSlug, isMobile, sidebarOpen], async ([slug, mobile, open]) => {
    if (!slug || (mobile && !open) || sidebarPosts.value || loadingSidebar) return
    loadingSidebar = true
    try {
      sidebarPosts.value = (await listPostSummaries(1, 50)).posts
    } catch {
      // Reopening the directory or visiting another article retries the request.
    } finally {
      loadingSidebar = false
    }
  }, { immediate: true })
})

watch(() => route.path, closeSidebar)
function onEscape(event: KeyboardEvent) {
  if (event.key === 'Escape' && sidebarOpen.value) {
    closeSidebar()
    document.querySelector<HTMLButtonElement>('.menu-btn')?.focus()
  }
}
watch(sidebarOpen, async (open) => {
  if (open) {
    await nextTick()
    document.querySelector<HTMLAnchorElement>('.sidebar a')?.focus()
  }
})
onMounted(() => window.addEventListener('keydown', onEscape))
onUnmounted(() => window.removeEventListener('keydown', onEscape))
</script>

<template>
  <div class="app-shell">
    <a class="skip-link" href="#main-content">Skip to content</a>
    <TerminalHeader :sidebar-open="sidebarOpen" @menu-click="toggleSidebar" />

    <div class="app-body">
      <div class="sidebar-overlay" :class="{ open: sidebarOpen }" @click="closeSidebar" />

      <aside id="site-sidebar" class="sidebar" :class="{ open: sidebarOpen }" :inert="isMobile && !sidebarOpen" aria-label="Site navigation">
        <div class="sidebar-inner">
          <div class="prompt-line" aria-hidden="true">
            <div class="prompt-path">{{ promptPath }}</div>
            <div class="prompt-cmd">{{ route.meta.requiresAuth ? '# ls -lh' : '$ ls -lh' }}</div>
          </div>

          <FileListing
            :directory="listingMode"
            :active-slug="activeArticleSlug"
            :posts="sidebarPosts"
            @navigate="closeSidebar"
          />
        </div>
      </aside>

      <main id="main-content" class="main-content" tabindex="-1">
        <router-view v-slot="{ Component }">
          <component :is="Component" :key="route.path" />
        </router-view>
      </main>
    </div>

    <TerminalFooter :meta="meta" />
    <ChatWidget v-if="meta?.ai && !isMobile" />
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 1rem;
}

.app-body {
  display: flex;
  flex: 1;
}

/* ---- 侧边栏 ---- */
.sidebar {
  width: 220px;
  min-width: 220px;
  border-right: 2px solid var(--border);
  background: var(--bg);
  transition: transform 0.2s;
}

.sidebar-inner {
  padding: 1.5rem 1rem;
  position: sticky;
  top: var(--header-height);
  max-height: calc(100dvh - var(--header-height));
  overflow-y: auto;
}

/* ---- 主内容 ---- */
.main-content {
  flex: 1;
  padding: 1.5rem 1rem;
  min-width: 0;
  background: var(--bg);
}

/* ---- prompt ---- */
.prompt-line {
  font-weight: bold;
  margin-bottom: 1rem;
  font-size: 0.82rem;
  line-height: 1.5;
}
.prompt-path { color: var(--fg); }
.prompt-cmd  { color: var(--muted); }

/* ---- 侧边栏遮罩 ---- */
.sidebar-overlay {
  display: none;
}

/* ---- 响应式 ---- */
@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 100;
    transform: translateX(-100%);
    box-shadow: none;
  }

  .sidebar.open { transform: translateX(0); box-shadow: var(--shadow); }
  .sidebar-inner { position: static; max-height: 100dvh; }
  .app-shell { padding: 0; }

  .sidebar-overlay {
    display: block;
    position: fixed;
    inset: 0;
    z-index: 99;
    background: rgba(0, 0, 0, 0.4);
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s;
  }

  .sidebar-overlay.open {
    opacity: 1;
    pointer-events: auto;
  }

  .main-content {
    padding: 1rem 0.75rem;
  }
}
</style>
