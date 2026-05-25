<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getPostAdmin, createPost, updatePost } from '../../api/posts'
import type { PostPayload } from '../../api/posts'
import AdminNav from '../../components/AdminNav.vue'
import ArticleRenderer from '../../components/ArticleRenderer.vue'

const SLUG = 'about'
const post = ref<PostPayload>({ slug: SLUG, title: '', content: '', published: true })
const isNew = ref(true)
const status = ref('')
const saving = ref(false)

onMounted(async () => {
  try {
    const data = await getPostAdmin(SLUG)
    post.value = data
    isNew.value = false
    status.value = `loaded: ${SLUG}.md`
  } catch {
    status.value = 'creating new about.md'
  }
})

async function handleSave() {
  saving.value = true
  try {
    if (isNew.value) {
      await createPost(post.value)
      isNew.value = false
      status.value = `created: ${SLUG}.md`
    } else {
      await updatePost(SLUG, post.value)
      status.value = `saved: ${SLUG}.md`
    }
  } catch (e: any) {
    status.value = 'ERROR: ' + (e.message || 'save failed')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AdminNav />

  <div class="editor-header">
    <div class="editor-prompt">
      root@k8s-node ~/admin $ <span class="muted">vim about.md</span>
    </div>
    <div class="editor-status" :class="{ 'status-error': status.startsWith('ERROR') }">
      {{ status || 'READY' }}
    </div>
  </div>

  <div class="meta-panel">
    <div class="meta-row">
      <label>TITLE:</label>
      <input v-model="post.title" type="text" placeholder="About page title" />
    </div>
  </div>

  <div class="editor-split">
    <div class="editor-pane">
      <div class="pane-label">RAW // MARKDOWN</div>
      <textarea
        v-model="post.content"
        class="editor-textarea"
        placeholder="Write your about page in markdown..."
        spellcheck="false"
      />
    </div>
    <div class="preview-pane">
      <div class="pane-label">PREVIEW // RENDERED</div>
      <div class="preview-scroll">
        <ArticleRenderer :content="post.content" />
      </div>
    </div>
  </div>

  <div class="editor-actions">
    <button class="btn" :disabled="saving" @click="handleSave">
      {{ saving ? '[ SAVING... ]' : '[ SAVE ]' }}
    </button>
  </div>
</template>

<style scoped>
.editor-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; gap: 0.5rem; }
.editor-prompt { font-weight: bold; font-size: 0.95rem; }
.muted { color: var(--muted); }
.editor-status { font-size: 0.8rem; font-weight: bold; padding: 2px 8px; border: 1px solid var(--border); color: var(--muted); }
.status-error { background: #ff4444; color: #fff !important; }
.meta-panel { border: 2px solid var(--border); padding: 1rem; margin-bottom: 1.5rem; background: var(--bg); box-shadow: var(--shadow); }
.meta-row { display: flex; align-items: center; gap: 10px; }
.meta-row label { font-weight: bold; font-size: 0.85rem; min-width: 60px; color: var(--muted); }
.meta-row input { flex: 1; background: var(--bg); color: var(--fg); border: 1px solid var(--border); padding: 6px 10px; font-family: var(--font-main); font-size: 0.9rem; outline: none; }
.meta-row input:focus { border-width: 2px; }
.editor-split { display: flex; gap: 0; border: 2px solid var(--border); margin-bottom: 1.5rem; min-height: 500px; box-shadow: var(--shadow); }
.editor-pane, .preview-pane { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.editor-pane { border-right: 2px solid var(--border); }
.pane-label { font-size: 0.75rem; font-weight: bold; padding: 6px 10px; border-bottom: 2px solid var(--border); background: var(--fg); color: var(--bg); text-transform: uppercase; }
.editor-textarea { flex: 1; width: 100%; background: var(--bg); color: var(--fg); border: none; padding: 1rem; font-family: var(--font-main); font-size: 0.9rem; line-height: 1.6; resize: none; outline: none; }
.preview-scroll { flex: 1; padding: 1rem; overflow-y: auto; }
.editor-actions { display: flex; gap: 10px; border-top: 2px dashed var(--border); padding-top: 1.5rem; }
@media (max-width: 768px) { .editor-split { flex-direction: column; min-height: auto; } .editor-pane { border-right: none; border-bottom: 2px solid var(--border); min-height: 300px; } .preview-pane { min-height: 300px; } }
</style>
