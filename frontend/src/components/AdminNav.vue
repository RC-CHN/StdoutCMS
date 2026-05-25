<script setup lang="ts">
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
</script>

<template>
  <nav class="admin-nav">
    <RouterLink to="/admin" class="admin-nav-item" :class="{ active: !route.name?.toString().includes('projects') && !route.name?.toString().includes('about') && route.name?.toString().startsWith('admin') }">
      [ posts ]
    </RouterLink>
    <RouterLink to="/admin/projects" class="admin-nav-item" :class="{ active: !!route.name?.toString().startsWith('admin-projects') }">
      [ projects ]
    </RouterLink>
    <RouterLink to="/admin/about" class="admin-nav-item" :class="{ active: route.name === 'admin-about' }">
      [ about ]
    </RouterLink>
  </nav>
</template>

<style scoped>
.admin-nav {
  display: flex;
  gap: 0;
  border-bottom: 2px solid var(--border);
  margin-bottom: 1.5rem;
}

.admin-nav-item {
  padding: 6px 16px;
  font-weight: bold;
  font-size: 0.85rem;
  color: var(--muted);
  text-decoration: none;
  border: 2px solid transparent;
  border-bottom: none;
  text-transform: uppercase;
  transition: all 0.1s;
}

.admin-nav-item:hover {
  color: var(--fg);
  border-color: var(--border);
}

.admin-nav-item.active {
  color: var(--fg);
  border-color: var(--border);
  background: var(--bg);
}
</style>
