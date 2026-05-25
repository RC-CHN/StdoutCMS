<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getPostAdmin, createPost, updatePost } from '../../api/posts'
import type { PostPayload } from '../../api/posts'
import { useDraft } from '../../composables/useDraft'
import ArticleRenderer from '../../components/ArticleRenderer.vue'
import AdminNav from '../../components/AdminNav.vue'

const route = useRoute()
const router = useRouter()
const slugParam = route.params.slug as string | undefined
const isEdit = !!slugParam

const { draft, lastSaved, save, clear, restoreFromPost } = useDraft()
const status = ref('')
const saving = ref(false)

// load existing post from API
onMounted(async () => {
  if (isEdit && slugParam) {
    status.value = `loading ${slugParam}.md...`
    try {
      const post = await getPostAdmin(slugParam)
      restoreFromPost({
        title: post.title,
        slug: post.slug,
        tags: post.tags || [],
        excerpt: post.excerpt || '',
        content: post.content,
      })
      status.value = `loaded: ${post.slug}.md`
    } catch (e: any) {
      status.value = 'ERROR: ' + (e.message || 'not found')
    }
  }
})

// 自动保存（防抖 1s）
let autoSaveTimer: ReturnType<typeof setTimeout>
watch(
  () => draft.value,
  () => {
    clearTimeout(autoSaveTimer)
    autoSaveTimer = setTimeout(() => {
      save()
      status.value = `draft auto-saved at ${new Date().toLocaleTimeString()}`
    }, 1000)
  },
  { deep: true }
)

const previewContent = computed(() => draft.value.content)

/* 视图模式 */
type ViewMode = 'split' | 'edit' | 'preview'
const viewMode = ref<ViewMode>('split')

function generateSlug() {
  const base = draft.value.title
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .slice(0, 50)
  draft.value.slug = base || 'untitled'
}

async function handlePublish() {
  saving.value = true
  const payload: PostPayload = {
    slug: draft.value.slug,
    title: draft.value.title,
    content: draft.value.content,
    excerpt: draft.value.excerpt,
    tags: draft.value.tags.split(',').map(s => s.trim()).filter(Boolean),
    author: 'root',
    wordCount: draft.value.content.replace(/\s/g, '').length,
    published: true,
  }
  try {
    if (isEdit && slugParam) {
      await updatePost(slugParam, payload)
      status.value = `PUBLISHED: ${slugParam}.md updated`
    } else {
      await createPost(payload)
      status.value = `PUBLISHED: ${draft.value.slug}.md created`
      router.push(`/admin/edit/${draft.value.slug}`)
    }
    clear()
  } catch (e: any) {
    status.value = 'ERROR: ' + (e.message || 'publish failed')
  } finally {
    saving.value = false
  }
}

async function handleSave() {
  saving.value = true
  const payload: PostPayload = {
    slug: draft.value.slug,
    title: draft.value.title,
    content: draft.value.content,
    excerpt: draft.value.excerpt,
    tags: draft.value.tags.split(',').map(s => s.trim()).filter(Boolean),
    author: 'root',
    wordCount: draft.value.content.replace(/\s/g, '').length,
    published: false,
  }
  try {
    if (isEdit && slugParam) {
      await updatePost(slugParam, payload)
      status.value = `SAVED: ${slugParam}.md`
    } else {
      await createPost(payload)
      status.value = `SAVED: ${draft.value.slug}.md`
      router.push(`/admin/edit/${draft.value.slug}`)
    }
    clear()
  } catch (e: any) {
    status.value = 'ERROR: ' + (e.message || 'save failed')
  } finally {
    saving.value = false
  }
}

function statusClass() {
  if (status.value.startsWith('ERROR')) return 'status-error'
  return 'status-ok'
}
</script>

<template>
  <AdminNav />

  <div class="editor-header">
    <div class="editor-prompt">
      root@k8s-node ~/admin $ <span class="muted">{{ isEdit ? `vim ${slugParam}.md` : 'vim new_post.md' }}</span>
    </div>
    <div class="editor-status" :class="statusClass()">
      {{ status || 'READY' }}
    </div>
  </div>

  <!-- 元数据编辑区 -->
  <div class="meta-panel">
    <div class="meta-row">
      <label>TITLE:</label>
      <input v-model="draft.title" type="text" placeholder="article title..." @blur="generateSlug" />
    </div>
    <div class="meta-row">
      <label>SLUG:</label>
      <input v-model="draft.slug" type="text" placeholder="url-slug" />
      <button class="btn btn-sm" @click="generateSlug">AUTO_GEN</button>
    </div>
    <div class="meta-row">
      <label>TAGS:</label>
      <input v-model="draft.tags" type="text" placeholder="tag1, tag2, tag3" />
    </div>
    <div class="meta-row">
      <label>EXCERPT:</label>
      <input v-model="draft.excerpt" type="text" placeholder="short summary..." />
    </div>
  </div>

  <!-- 视图选项卡 -->
  <div class="view-tabs">
    <button
      class="tab-btn"
      :class="{ active: viewMode === 'split' }"
      @click="viewMode = 'split'"
    >[ SPLIT_VIEW ]</button>
    <button
      class="tab-btn"
      :class="{ active: viewMode === 'edit' }"
      @click="viewMode = 'edit'"
    >[ EDIT_ONLY ]</button>
    <button
      class="tab-btn"
      :class="{ active: viewMode === 'preview' }"
      @click="viewMode = 'preview'"
    >[ PREVIEW_ONLY ]</button>
  </div>

  <!-- 编辑区 -->
  <div class="editor-area" :class="`mode-${viewMode}`">
    <!-- 编辑面板 -->
    <div class="editor-pane" v-show="viewMode !== 'preview'">
      <div class="pane-label">RAW // MARKDOWN</div>
      <textarea
        v-model="draft.content"
        class="editor-textarea"
        placeholder="# Title\n\nWrite your markdown here..."
        spellcheck="false"
      />
    </div>
    <!-- 预览面板 -->
    <div class="preview-pane" v-show="viewMode !== 'edit'">
      <div class="pane-label">PREVIEW // RENDERED</div>
      <div class="preview-scroll">
        <ArticleRenderer :content="previewContent" />
      </div>
    </div>
  </div>

  <!-- 操作栏 -->
  <div class="editor-actions">
    <div class="action-left">
      <button class="btn" @click="handleSave">[ SAVE_DRAFT ]</button>
      <button class="btn" @click="handlePublish">[ PUBLISH ]</button>
      <button class="btn" @click="clear">[ CLEAR ]</button>
    </div>
    <div class="action-right">
      <span class="meta-info">WORDS: {{ draft.content.replace(/\s/g, '').length }}</span>
      <span class="meta-info" v-if="lastSaved">LAST_SAVE: {{ new Date(lastSaved).toLocaleTimeString() }}</span>
    </div>
  </div>

  <div class="back-link">
    <RouterLink to="/admin" class="btn">cd ..</RouterLink>
  </div>
</template>

<style scoped>
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.editor-prompt {
  font-weight: bold;
  font-size: 0.95rem;
}

.muted {
  color: var(--muted);
}

.editor-status {
  font-size: 0.8rem;
  font-weight: bold;
  padding: 2px 8px;
  border: 1px solid var(--border);
}

.status-ok {
  background: var(--bg);
  color: var(--muted);
}

.status-error {
  background: #ff4444;
  color: #fff;
}

/* 元数据面板 */
.meta-panel {
  border: 2px solid var(--border);
  padding: 1rem;
  margin-bottom: 1.5rem;
  background: var(--bg);
  box-shadow: var(--shadow);
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 0.8rem;
}

.meta-row:last-child {
  margin-bottom: 0;
}

.meta-row label {
  font-weight: bold;
  font-size: 0.85rem;
  min-width: 70px;
  color: var(--muted);
}

.meta-row input {
  flex: 1;
  background: var(--bg);
  color: var(--fg);
  border: 1px solid var(--border);
  padding: 6px 10px;
  font-family: var(--font-main);
  font-size: 0.9rem;
  outline: none;
}

.meta-row input:focus {
  border-width: 2px;
}

/* 视图选项卡 */
.view-tabs {
  display: flex;
  gap: 0;
  margin-bottom: 0;
  border: 2px solid var(--border);
  border-bottom: none;
}

.tab-btn {
  flex: 1;
  background: var(--bg);
  color: var(--muted);
  border: none;
  border-right: 2px solid var(--border);
  padding: 8px 12px;
  font-family: inherit;
  font-size: 0.8rem;
  font-weight: bold;
  cursor: pointer;
  text-transform: uppercase;
  transition: all 0.1s;
}

.tab-btn:last-child {
  border-right: none;
}

.tab-btn:hover {
  background: var(--fg);
  color: var(--bg);
}

.tab-btn.active {
  background: var(--fg);
  color: var(--bg);
}

/* 编辑区 */
.editor-area {
  display: flex;
  gap: 0;
  border: 2px solid var(--border);
  margin-bottom: 1.5rem;
  min-height: 500px;
  box-shadow: var(--shadow);
}

.editor-area.mode-edit,
.editor-area.mode-preview {
  min-height: 600px;
}

.editor-pane,
.preview-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.editor-pane {
  border-right: 2px solid var(--border);
}

.mode-edit .editor-pane,
.mode-preview .preview-pane {
  border-right: none;
}

.pane-label {
  font-size: 0.75rem;
  font-weight: bold;
  padding: 6px 10px;
  border-bottom: 2px solid var(--border);
  background: var(--fg);
  color: var(--bg);
  text-transform: uppercase;
}

.editor-textarea {
  flex: 1;
  width: 100%;
  background: var(--bg);
  color: var(--fg);
  border: none;
  padding: 1rem;
  font-family: var(--font-main);
  font-size: 0.9rem;
  line-height: 1.6;
  resize: none;
  outline: none;
}

.preview-scroll {
  flex: 1;
  padding: 1rem;
  overflow-y: auto;
}

/* 操作栏 */
.editor-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1.5rem;
  border-top: 2px dashed var(--border);
  padding-top: 1.5rem;
}

.action-left {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.action-right {
  display: flex;
  gap: 15px;
  font-size: 0.8rem;
  color: var(--muted);
}

.back-link {
  margin-bottom: 2rem;
}

.btn-sm {
  padding: 2px 8px;
  font-size: 0.8rem;
}

@media (max-width: 768px) {
  .editor-split {
    flex-direction: column;
    min-height: auto;
  }
  .editor-pane {
    border-right: none;
    border-bottom: 2px solid var(--border);
    min-height: 300px;
  }
  .preview-pane {
    min-height: 300px;
  }
  .meta-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .meta-row label {
    min-width: auto;
  }
  .meta-row input {
    width: 100%;
  }
}
</style>
