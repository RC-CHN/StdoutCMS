<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { fetchMeta } from '../api/meta'
import type { MetaResponse } from '../api/meta'

const route = useRoute()
const meta = ref<MetaResponse | null>(null)

onMounted(async () => {
  try { meta.value = await fetchMeta() } catch { /* ignore */ }
})

const marqueeText = computed(() => {
  const m = meta.value
  if (route.name === 'article') {
    return '*** READING_MODE *** 0 TRACKERS *** 0 ANALYTICS *** PLAIN TEXT ONLY *** ENJOY ***'
  }
  if (route.name === 'about') {
    return '*** ABOUT *** A HUMAN BEING *** POWERED BY CURIOSITY AND COFFEE ***'
  }
  if (route.name === 'projects') {
    return '*** PROJECTS *** COMPILED FROM IDEAS *** SHIPPED WITH LOVE ***'
  }
  const parts = ['*** STDOUT_CMS_ELF ***']
  if (m) parts.push(`POSTS: ${m.posts}`)
  if (m) parts.push(`PROJECTS: ${m.projects}`)
  if (m) parts.push(`UPTIME: ${m.uptime}`)
  parts.push('STATELESS *** ALL GREEN ***')
  return parts.join(' *** ')
})

const footerRight = computed(() => {
  if (meta.value) return `UPTIME: ${meta.value.uptime}`
  return 'LOADING...'
})
</script>

<template>
  <div class="marquee-container">
    <div class="marquee-content">{{ marqueeText }}</div>
  </div>

  <div class="container" style="padding-top: 0; padding-bottom: 2rem;">
    <footer>
      <span>(c) 2026 Ruochen_Pan.</span>
      <span>{{ footerRight }}</span>
    </footer>
  </div>
</template>
