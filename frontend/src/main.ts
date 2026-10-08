import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'

try {
  const saved = localStorage.getItem('blog_theme_pref')
  if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.body.classList.add('dark-mode')
  }
} catch { /* Theme still works when storage is unavailable. */ }

createApp(App).use(router).mount('#app')
