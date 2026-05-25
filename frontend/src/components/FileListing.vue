<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import type { PostPayload } from '../api/posts'

defineEmits<{
  navigate: []
}>()

const props = defineProps<{
  directory: 'root' | 'articles'
  activeSlug?: string
  posts?: PostPayload[]
}>()

const route = useRoute()
const { isLoggedIn } = useAuth()

type DirEntry = {
  date: string
  name: string
  suffix: string
  to: string
  isActive: boolean
}

const entries = computed<DirEntry[]>(() => {
  if (props.directory === 'articles') {
    const result: DirEntry[] = [
      {
        date: '',
        name: '../',
        suffix: '',
        to: '/',
        isActive: false,
      },
    ]
    for (const post of props.posts ?? []) {
      result.push({
        date: post.createdAt || '',
        name: post.slug,
        suffix: '.md',
        to: `/article/${post.slug}`,
        isActive: props.activeSlug === post.slug,
      })
    }
    return result
  }

  const base: DirEntry[] = [
    {
      date: '2023-10-25',
      name: 'articles',
      suffix: '/',
      to: '/',
      isActive: route.name === 'home' || route.name === 'article',
    },
    {
      date: '2023-10-01',
      name: 'about',
      suffix: '/',
      to: '/about',
      isActive: route.name === 'about',
    },
    {
      date: '2023-09-15',
      name: 'projects',
      suffix: '/',
      to: '/projects',
      isActive: route.name === 'projects',
    },
  ]

  // 登录后才显示 admin 入口
  if (isLoggedIn.value) {
    base.push({
      date: '',
      name: 'admin',
      suffix: '/',
      to: '/admin',
      isActive: !!route.name?.toString().startsWith('admin-'),
    })
  }

  return base
})
</script>

<template>
  <div class="fs-listing">
    <RouterLink
      v-for="entry in entries"
      :key="entry.name"
      :to="entry.to"
      class="fs-entry"
      :class="{ 'fs-active': entry.isActive }"
      @click="$emit('navigate')"
    >
      <span class="fs-name">{{ entry.name }}<span class="fs-suffix">{{ entry.suffix }}</span></span>
      <span class="fs-date">{{ entry.date }}</span>
    </RouterLink>
  </div>
</template>

<style scoped>
.fs-listing {
  font-family: var(--font-main);
  font-size: 0.82rem;
  line-height: 1.4;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.fs-entry {
  display: flex;
  flex-direction: column;
  padding: 3px 6px;
  text-decoration: none;
  color: var(--fg);
  border: 1px solid transparent;
  border-radius: 2px;
  transition: background 0.1s;
}

.fs-entry:hover {
  background: var(--fg);
  color: var(--bg);
}

.fs-active {
  background: var(--fg);
  color: var(--bg);
}

.fs-name {
  font-weight: bold;
  overflow-wrap: break-word;
  line-height: 1.3;
}

.fs-suffix {
  color: var(--muted);
}

.fs-entry:hover .fs-suffix,
.fs-active .fs-suffix {
  color: inherit;
}

.fs-date {
  font-size: 0.75rem;
  color: var(--muted);
  line-height: 1.3;
}

.fs-entry:hover .fs-date,
.fs-active .fs-date {
  color: inherit;
}
</style>
