<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { MetaResponse } from '../api/meta'

const route = useRoute()
const props = defineProps<{ meta: MetaResponse | null }>()

const marqueeText = computed(() => {
  const m = props.meta
  if (route.name === 'article') {
    return '*** READING MODE *** CAT ARTICLE.MD *** MARKDOWN ***'
  }
  if (route.name === 'about') {
    return '*** ABOUT *** CAT ABOUT.MD ***'
  }
  if (route.name === 'projects') {
    return '*** PROJECTS *** LS -LH PROJECTS/ ***'
  }
  const parts = ['*** STDOUT_CMS_ELF ***']
  if (m) parts.push(`ARTICLES: ${m.posts}`)
  if (m) parts.push(`PROJECTS: ${m.projects}`)
  if (m) parts.push(`UPTIME: ${m.uptime}`)
  return parts.join(' *** ')
})

const footerRight = computed(() => {
  if (props.meta) return `BUILD: ${props.meta.version} | UPTIME: ${props.meta.uptime}`
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
