<script setup lang="ts">
import { onMounted, onUnmounted, computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import TerminalHeader from './components/TerminalHeader.vue'
import TerminalFooter from './components/TerminalFooter.vue'
import FileListing from './components/FileListing.vue'
import ChatWidget from './components/ChatWidget.vue'
import { fetchMeta } from './api/meta'
import { listPosts } from './api/posts'
import type { PostPayload } from './api/posts'
import { useBreakpoint } from './composables/useBreakpoint'

const route = useRoute()
const { isMobile } = useBreakpoint()

const THEME_KEY = 'blog_theme_pref'
const aiEnabled = ref(false)
const sidebarPosts = ref<PostPayload[]>([])

function initTheme() {
  const saved = localStorage.getItem(THEME_KEY)
  if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.body.classList.add('dark-mode')
  }
}

onMounted(() => {
  initTheme()
  fetchMeta()
    .then(m => { aiEnabled.value = m.ai })
    .catch(() => { aiEnabled.value = false })
  // 拉取文章列表填充侧边栏（取较多条方便浏览）
  listPosts(1, 50)
    .then(data => { sidebarPosts.value = data.posts })
    .catch(() => { sidebarPosts.value = [] })
})

/* 侧边栏 toggle */
const sidebarOpen = ref(false)
function toggleSidebar() { sidebarOpen.value = !sidebarOpen.value }
function closeSidebar() { sidebarOpen.value = false }

/* 路径指示 */
const promptPath = computed(() => {
  const base = 'guest@k8s-node ~'
  if (route.name === 'article') return base + '/articles'
  if (route.name === 'about')   return base + '/about'
  if (route.name === 'projects') return base + '/projects'
  return base
})

/* 侧边栏目录上下文 */
const listingMode = computed<'root' | 'articles'>(() =>
  route.name === 'article' ? 'articles' : 'root'
)

const activeArticleSlug = computed(() =>
  route.name === 'article' ? (route.params.slug as string) : undefined
)

/* fake-log page transition */
const navLog = ref<string[] | null>(null)
let navTimer: number | undefined

// play once per section — first visit gets the boot log, later
// navigations within/back to it just fade in without the flicker
const seenSections = new Set<string>()

watch(() => route.name, (name) => {
  const key = String(name ?? 'home')
  if (seenSections.has(key)) return
  seenSections.add(key)
  if (navTimer) clearTimeout(navTimer)
  navLog.value = [`$ cd ${route.fullPath}`, 'fetching page...', '[ OK ]']
  navTimer = window.setTimeout(() => { navLog.value = null }, 550)
})

onUnmounted(() => { if (navTimer) clearTimeout(navTimer) })
</script>

<template>
  <div class="app-shell">
    <TerminalHeader @menu-click="toggleSidebar" />

    <div class="app-body">
      <div class="sidebar-overlay" :class="{ open: sidebarOpen }" @click="closeSidebar" />

      <aside class="sidebar" :class="{ open: sidebarOpen }">
        <div class="sidebar-inner">
          <div class="prompt-line">
            <div class="prompt-path">{{ promptPath }}</div>
            <div class="prompt-cmd">$ ls -lh</div>
          </div>

          <FileListing
            :directory="listingMode"
            :active-slug="activeArticleSlug"
            :posts="sidebarPosts"
            @navigate="closeSidebar"
          />
        </div>
      </aside>

      <main class="main-content" @click="closeSidebar">
        <div v-if="navLog" class="nav-log">
          <div
            v-for="(line, i) in navLog"
            :key="i"
            class="nav-log-line"
            :class="{ ok: line === '[ OK ]' }"
            :style="{ animationDelay: `${i * 0.12}s` }"
          >{{ line }}</div>
          <span class="cursor" />
        </div>
        <div v-else :key="route.fullPath" class="page-wrapper">
          <router-view />
        </div>
      </main>
    </div>

    <TerminalFooter />
    <ChatWidget v-if="aiEnabled && !isMobile" />
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
  max-height: calc(100vh - 100px);
  overflow-y: auto;
}

/* ---- 主内容 ---- */
.main-content {
  flex: 1;
  padding: 1.5rem 1rem;
  max-width: none;
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

/* ---- 页面切换动画 ---- */
.page-wrapper {
  animation: page-in 0.2s ease;
}

/* fake log transition overlay */
.nav-log {
  padding: 0.5rem 0;
  font-size: 0.9rem;
  font-weight: bold;
}

.nav-log-line {
  opacity: 0;
  animation: log-line 0.01s step-end forwards;
}

.nav-log-line.ok {
  color: var(--accent);
}

@keyframes log-line {
  to { opacity: 1; }
}

@keyframes page-in {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

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
    box-shadow: var(--shadow);
  }

  .sidebar.open { transform: translateX(0); }

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
