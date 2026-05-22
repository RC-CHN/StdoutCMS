<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

defineEmits<{
  'menu-click': []
}>()

const route = useRoute()
const router = useRouter()

const THEME_KEY = 'blog_theme_pref'

function toggleTheme() {
  document.body.classList.toggle('dark-mode')
  const isDark = document.body.classList.contains('dark-mode')
  localStorage.setItem(THEME_KEY, isDark ? 'dark' : 'light')
}

const typeText = computed(() => {
  if (route.name === 'article') return 'rendering document...'
  if (route.name === 'about') return 'cat about.md...'
  if (route.name === 'projects') return 'executing projects.sh...'
  return 'ONLINE & READY...'
})

const displayed = ref('')
let i = 0

function typeWriter() {
  const text = typeText.value
  if (i < text.length) {
    displayed.value += text.charAt(i)
    i++
    setTimeout(typeWriter, Math.random() * 80 + 40)
  }
}

onMounted(() => {
  displayed.value = ''
  i = 0
  setTimeout(typeWriter, 300)
})
</script>

<template>
  <header class="app-header">
    <button class="menu-btn" @click="$emit('menu-click')" aria-label="Toggle sidebar">
      [ ≡ ]
    </button>

    <div class="title-block">
      <h1 @click="router.push('/')">SYS_BLOG.EXE</h1>
      <div class="status-line">
        <span>mode: {{ route.name === 'home' ? 'browse' : route.name }} | </span>
        <span>{{ displayed }}</span><span class="cursor"></span>
      </div>
    </div>

    <button class="theme-btn" @click="toggleTheme">[ INVERT_COLORS ]</button>
  </header>
</template>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-bottom: 3px solid var(--border);
  padding: 1rem 1.5rem;
  flex-wrap: wrap;
  gap: 1rem;
  background: var(--bg);
  position: sticky;
  top: 0;
  z-index: 50;
}

.title-block h1 {
  font-size: 2.5rem;
  text-transform: uppercase;
  letter-spacing: -1px;
  font-weight: bold;
  cursor: pointer;
}

.status-line {
  margin-top: 5px;
}

.theme-btn {
  flex-shrink: 0;
}

/* 汉堡按钮 */
.menu-btn {
  display: none;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .app-header {
    padding: 0.75rem 1rem;
    align-items: center;
  }

  .title-block h1 { font-size: 1.5rem; }

  .menu-btn {
    display: inline-block;
  }
}

@media (max-width: 600px) {
  .app-header { flex-direction: row; align-items: center; }
  .title-block h1 { font-size: 1.4rem; }
}
</style>
