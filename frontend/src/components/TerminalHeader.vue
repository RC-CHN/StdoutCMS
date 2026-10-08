<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { logout } from '../api/auth'

defineProps<{ sidebarOpen: boolean }>()
defineEmits<{ 'menu-click': [] }>()
const route = useRoute()
const router = useRouter()
const { isLoggedIn, clearToken } = useAuth()
const dark = ref(document.body.classList.contains('dark-mode'))
function toggleTheme() {
  dark.value = document.body.classList.toggle('dark-mode')
  try { localStorage.setItem('blog_theme_pref', dark.value ? 'dark' : 'light') } catch { /* optional preference */ }
}
async function handleLogout() {
  try { await logout() } catch { /* clear local login state */ }
  clearToken()
  router.push('/')
}
const typeText = computed(() => {
  if (route.name === 'article') return 'reading document'
  if (route.name === 'about') return 'cat about.md'
  if (route.name === 'projects') return 'projects.sh'
  return 'ONLINE & READY'
})
const displayed = ref('')
let timer: ReturnType<typeof setTimeout> | undefined
const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
function startTypeWriter() {
  clearTimeout(timer)
  const text = typeText.value
  if (motion.matches) { displayed.value = text; return }
  displayed.value = ''
  let i = 0
  function tick() {
    displayed.value = text.slice(0, ++i)
    if (i < text.length) timer = setTimeout(tick, 45)
  }
  tick()
}
const header = ref<HTMLElement | null>(null)
let resize: ResizeObserver | undefined
onMounted(() => {
  dark.value = document.body.classList.contains('dark-mode')
  startTypeWriter()
  motion.addEventListener('change', startTypeWriter)
  resize = new ResizeObserver(() => {
    if (header.value) document.documentElement.style.setProperty('--header-height', header.value.offsetHeight + 'px')
  })
  if (header.value) resize.observe(header.value)
})
watch(typeText, startTypeWriter)
onUnmounted(() => {
  clearTimeout(timer)
  resize?.disconnect()
  motion.removeEventListener('change', startTypeWriter)
})
</script>

<template>
  <header ref="header" class="app-header">
    <button class="menu-btn" @click="$emit('menu-click')" aria-label="Toggle navigation" aria-controls="site-sidebar" :aria-expanded="sidebarOpen">≡</button>
    <div class="title-block">
      <RouterLink to="/" class="site-title">STDOUT_CMS_ELF</RouterLink>
      <div class="status-line" aria-hidden="true">{{ displayed }}<span class="cursor"></span></div>
    </div>
    <div class="header-actions">
      <button v-if="isLoggedIn" class="logout-btn" @click="handleLogout"><span class="wide-label">[ LOGOUT ]</span><span class="compact-label">EXIT</span></button>
      <button class="theme-btn" @click="toggleTheme" :aria-label="dark ? 'Switch to light theme' : 'Switch to dark theme'" :aria-pressed="dark">
        <span class="wide-label">[ {{ dark ? 'LIGHT' : 'DARK' }} ]</span>
        <span class="compact-label" aria-hidden="true">{{ dark ? '☀' : '◐' }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.app-header { display: flex; align-items: center; gap: 1rem; border-bottom: 2px solid var(--border); padding: 1rem 1.5rem; background: var(--bg); position: sticky; top: 0; z-index: 50; }
.title-block { flex: 1; min-width: 0; }
.site-title { font-size: 2rem; letter-spacing: -1px; font-weight: bold; color: var(--fg); text-decoration: none; }
.status-line { color: var(--muted); font-size: 0.75rem; min-height: 22px; white-space: nowrap; overflow: hidden; }
.status-line .cursor { width: 7px; margin-left: 4px; height: 0.9em; }
.header-actions { display: flex; gap: 6px; flex-shrink: 0; }
.header-actions button { font-size: 0.75rem; white-space: nowrap; }
.menu-btn, .compact-label { display: none; }
@media (max-width: 768px) {
  .app-header { padding: 0.5rem 0.75rem; gap: 0.5rem; }
  .site-title { font-size: clamp(0.88rem, 3.9vw, 1.3rem); letter-spacing: -0.5px; white-space: nowrap; }
  .status-line { font-size: 0.65rem; line-height: 1.5; min-height: 16px; }
  .menu-btn { display: block; font-size: 1.5rem; border: 0; padding: 0; width: 36px; height: 44px; flex-shrink: 0; }
  .header-actions button { padding: 0 5px; min-width: 36px; height: 44px; border-width: 1px; font-size: 0.65rem; }
  .theme-btn .compact-label { font-size: 1.2rem; }
  .wide-label { display: none; }
  .compact-label { display: inline; }
}
</style>
