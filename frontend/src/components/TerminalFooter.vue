<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const marqueeText = computed(() => {
  if (route.name === 'article') {
    return '*** READING_MODE ENGAGED *** CONNECTION SECURE *** NO TRACKERS BLOCKED (BECAUSE THERE ARE NONE) *** ENJOY THE PLAIN TEXT ***'
  }
  if (route.name === 'about') {
    return '*** ABOUT_MODE *** NO SECRETS HERE *** JUST A HUMAN BEING ***'
  }
  if (route.name === 'projects') {
    return '*** PROJECT_MODE *** COMPILING IDEAS *** SHIPPING CODE ***'
  }
  return '*** SYSTEM OPTIMAL *** CLUSTER NODES: 3/3 READY *** UPTIME: 99.99% *** NO STATEFUL SETS DETECTED *** ALL SYSTEMS GO ***'
})

const footerRight = computed(() => {
  if (route.name === 'article') return 'READ_TIME: 3m 12s'
  return 'RENDER_TIME: 12ms'
})
</script>

<template>
  <div class="marquee-container">
    <div class="marquee-content">{{ marqueeText }}</div>
  </div>

  <div class="container" style="padding-top: 0; padding-bottom: 2rem;">
    <footer>
      <span>(c) 202X YOUR_NAME.</span>
      <span>{{ footerRight }}</span>
    </footer>
  </div>
</template>
