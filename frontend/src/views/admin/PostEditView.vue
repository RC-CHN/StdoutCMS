<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import { getPostAdmin, createPost, updatePost } from '../../api/posts'
import type { PostPayload } from '../../api/posts'
import { generateMeta, type GenerateMetaRes } from '../../api/chat'
import { fetchMeta } from '../../api/meta'
import { useDraft, emptyDraft, draftText, type Draft } from '../../composables/useDraft'
import { useImagePaste } from '../../composables/useImagePaste'
import { useEditorToolbar } from '../../composables/useEditorToolbar'
import { useEditorUploads } from '../../composables/useEditorUploads'
import { useBreakpoint } from '../../composables/useBreakpoint'
import ArticleRenderer from '../../components/ArticleRenderer.vue'
import MobileArticleRenderer from '../../components/MobileArticleRenderer.vue'
import AdminNav from '../../components/AdminNav.vue'
import PageState from '../../components/PageState.vue'
import UploadStatus from '../../components/UploadStatus.vue'
import TerminalFeedback from '../../components/TerminalFeedback.vue'

const route = useRoute()
const router = useRouter()
const slugParam = route.params.slug as string | undefined
const isEdit = !!slugParam
const { isMobile } = useBreakpoint()
const { draft, lastSaved, storageError, load, save, discard, restore } = useDraft(slugParam)
const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const generatingField = ref<'slug' | 'tags' | 'excerpt' | null>(null)
const published = ref(false)
const serverSaved = ref('')
const baseline = ref(draftText(emptyDraft()))
const pendingDraft = ref<Draft | null>(null)
const dirty = computed(() => draftText(draft.value) !== baseline.value)
const feedbackMsg = ref<string | null>(null)
const feedbackType = ref<'ok' | 'err'>('ok')
const feedbackTrigger = ref(0)
const author = ref('root')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const uploads = useEditorUploads(() => textareaRef.value)
const { attach, detach } = useImagePaste(() => textareaRef.value, uploads)
const toolbar = useEditorToolbar(uploads)
const blocked = computed(() => saving.value || loading.value || !!generatingField.value || !!loadError.value || !!pendingDraft.value)
const canSave = computed(() => !blocked.value && !uploads.uploading.value && !uploads.failures.value.length)

function showFeedback(message: string, type: 'ok' | 'err') {
  feedbackMsg.value = message
  feedbackType.value = type
  feedbackTrigger.value++
}
async function loadEditor() {
  loading.value = true
  loadError.value = ''
  try {
    if (slugParam) {
      const post = await getPostAdmin(slugParam)
      restore({ ...emptyDraft(), ...post, tags: (post.tags || []).join(', '), excerpt: post.excerpt || '' })
      published.value = !!post.published
      author.value = post.author || 'root'
      serverSaved.value = post.updatedAt || ''
    }
    baseline.value = draftText(draft.value)
    const backup = load()
    if (backup && draftText(backup) !== baseline.value) pendingDraft.value = backup
    else if (backup) discard()
  } catch (cause) {
    loadError.value = cause instanceof Error ? cause.message : 'Could not load this article.'
  } finally { loading.value = false }
}
function restoreBackup() {
  if (pendingDraft.value) restore({ ...pendingDraft.value, slug: slugParam || pendingDraft.value.slug })
  pendingDraft.value = null
}
function discardBackup() {
  if (discard()) pendingDraft.value = null
}
let autoSaveTimer: ReturnType<typeof setTimeout> | undefined
function flushBackup() {
  clearTimeout(autoSaveTimer)
  if (!loading.value && !loadError.value && !pendingDraft.value && dirty.value) return save()
  return true
}
watch(() => draftText(draft.value), () => {
  clearTimeout(autoSaveTimer)
  if (blocked.value || !dirty.value) return
  autoSaveTimer = setTimeout(flushBackup, 1000)
})
watch(textareaRef, attach, { flush: 'post' })
function beforeUnload(event: BeforeUnloadEvent) {
  if (!flushBackup() || uploads.uploading.value) {
    event.preventDefault()
    event.returnValue = ''
  }
}
onBeforeRouteLeave(() => {
  const stored = flushBackup()
  if (!stored || uploads.uploading.value) return window.confirm('Some changes or uploads are not saved. Leave this editor?')
})
onMounted(() => {
  void loadEditor()
  window.addEventListener('beforeunload', beforeUnload)
  fetchMeta().then(meta => { aiEnabled.value = meta.ai }).catch(() => {})
})
onUnmounted(() => {
  flushBackup()
  detach()
  uploads.dispose()
  window.removeEventListener('beforeunload', beforeUnload)
})
type ViewMode = 'split' | 'edit' | 'preview'
const selectedMode = ref<ViewMode>('split')
const viewMode = computed(() => isMobile.value && selectedMode.value === 'split' ? 'edit' : selectedMode.value)
const aiEnabled = ref(false)
const metaCache = ref<GenerateMetaRes | null>(null)
watch(() => [draft.value.title, draft.value.content], () => { metaCache.value = null })
async function generateField(field: 'slug' | 'tags' | 'excerpt') {
  if (generatingField.value || blocked.value) return
  generatingField.value = field
  try {
    const result = metaCache.value || await generateMeta({ title: draft.value.title, content: draft.value.content })
    metaCache.value = result
    const value = result[field]?.trim()
    if (!value) throw new Error('No suggestion returned. Try again.')
    draft.value[field] = value
  } catch (cause) {
    showFeedback(cause instanceof Error ? cause.message : 'Generation failed. Try again.', 'err')
  } finally { generatingField.value = null }
}
async function saveToServer(nextPublished = published.value) {
  if (!canSave.value) return
  if (!draft.value.title.trim() || !/^[a-z0-9]+(-[a-z0-9]+)*$/.test(draft.value.slug)) {
    showFeedback('Enter a title and a URL slug using lowercase letters, numbers and single hyphens.', 'err')
    return
  }
  if (/uploading-|\[upload failed:/.test(draft.value.content)) {
    showFeedback('Finish or remove the incomplete uploads before saving.', 'err')
    return
  }
  saving.value = true
  clearTimeout(autoSaveTimer)
  const wordCount = draft.value.content.replace(/\s/g, '').length
  const seconds = Math.round(wordCount / 400 * 60)
  const payload: PostPayload = {
    slug: slugParam || draft.value.slug, title: draft.value.title, content: draft.value.content,
    excerpt: draft.value.excerpt, tags: draft.value.tags.split(',').map(tag => tag.trim()).filter(Boolean),
    author: author.value, wordCount,
    readTime: seconds < 60 ? seconds + 'S' : Math.round(seconds / 60) + ' MIN',
    published: nextPublished,
  }
  try {
    if (slugParam) await updatePost(slugParam, payload)
    else await createPost(payload)
    published.value = nextPublished
    serverSaved.value = new Date().toISOString()
    baseline.value = draftText(draft.value)
    discard()
    showFeedback(payload.slug + '.md saved to server', 'ok')
    if (!isEdit) await router.replace('/admin/edit/' + payload.slug)
  } catch (cause) {
    save()
    showFeedback(cause instanceof Error ? cause.message : 'Server save failed. Try again.', 'err')
  } finally { saving.value = false }
}
function unpublish() {
  if (window.confirm('Unpublish this article? It will no longer be visible to readers.')) void saveToServer(false)
}
function clearEditor() {
  if (!window.confirm('Clear the editor? This replaces the local backup; the server copy is unchanged.')) return
  draft.value = { ...emptyDraft(), slug: slugParam || '' }
}
</script>

<template>
  <AdminNav />
  <div class="editor-header">
    <span class="editor-prompt">root@k8s-node $ vim {{ slugParam || 'new_post' }}.md</span>
    <span class="publication-state">{{ published ? '[ PUBLISHED ]' : '[ DRAFT ]' }}</span>
  </div>
  <PageState v-if="loading" mode="loading" message="Loading article…" />
  <PageState v-else-if="loadError" mode="error" :message="loadError" retry @retry="loadEditor" />
  <template v-else>
    <div v-if="pendingDraft" class="recovery-notice" role="status">
      <p>Local backup found from {{ new Date(pendingDraft.savedAt).toLocaleString() }}.</p>
      <p>Restore it to continue writing, or keep the server version.</p>
      <div class="state-actions">
        <button @click="restoreBackup">RESTORE BACKUP</button>
        <button @click="discardBackup">USE SERVER VERSION</button>
      </div>
    </div>
    <p v-if="storageError" class="error-notice" role="alert">{{ storageError }}</p>
    <fieldset class="editor-fields" :disabled="blocked">
      <label class="title-label" for="post-title">TITLE</label>
      <input id="post-title" v-model="draft.title" class="title-input" placeholder="Article title…" />
      <details class="meta-panel" :open="!isEdit">
        <summary>ARTICLE DETAILS <span>URL / tags / summary</span></summary>
        <div class="meta-row">
          <label for="post-slug">SLUG</label>
          <input id="post-slug" v-model="draft.slug" :readonly="isEdit" placeholder="url-slug" />
          <button v-if="aiEnabled && !isEdit" :disabled="!!generatingField" @click="generateField('slug')">GEN_SLUG</button>
        </div>
        <div class="meta-row">
          <label for="post-tags">TAGS</label>
          <input id="post-tags" v-model="draft.tags" placeholder="tag1, tag2" />
          <button v-if="aiEnabled" :disabled="!!generatingField" @click="generateField('tags')">GEN_TAGS</button>
        </div>
        <div class="meta-row">
          <label for="post-excerpt">SUMMARY</label>
          <textarea id="post-excerpt" v-model="draft.excerpt" rows="2" placeholder="A short introduction…" />
          <button v-if="aiEnabled" :disabled="!!generatingField" @click="generateField('excerpt')">GEN_EXCERPT</button>
        </div>
      </details>
      <div class="view-tabs" role="group" aria-label="Editor view">
        <button v-if="!isMobile" :aria-pressed="viewMode === 'split'" @click="selectedMode = 'split'">SPLIT VIEW</button>
        <button :aria-pressed="viewMode === 'edit'" @click="selectedMode = 'edit'">EDIT</button>
        <button :aria-pressed="viewMode === 'preview'" @click="selectedMode = 'preview'">PREVIEW</button>
      </div>
      <div class="editor-area" :class="'mode-' + viewMode">
        <div v-show="viewMode !== 'preview'" class="editor-pane">
          <div class="editor-toolbar">
            <label for="post-content">MARKDOWN</label>
            <div class="toolbar-actions">
              <button v-for="kind in (['image', 'audio', 'video', 'file'] as const)" :key="kind"
                :disabled="uploads.uploading.value" :title="'Upload ' + kind" @click="toolbar.uploadMedia(kind)">
                + {{ kind }}
              </button>
            </div>
          </div>
          <textarea id="post-content" ref="textareaRef" v-model="draft.content" class="editor-textarea"
            placeholder="Write your markdown here…" spellcheck="false" />
        </div>
        <div v-show="viewMode !== 'edit'" class="preview-pane">
          <div class="pane-label">PREVIEW</div>
          <div class="preview-scroll">
            <MobileArticleRenderer v-if="isMobile" :content="draft.content" />
            <ArticleRenderer v-else :content="draft.content" />
          </div>
        </div>
      </div>
    </fieldset>
    <UploadStatus :uploads="uploads" />
    <div class="editor-actions">
      <div class="action-left">
        <button class="btn-primary" :disabled="!canSave" @click="saveToServer()">{{ saving ? 'SAVING…' : published ? 'SAVE CHANGES' : 'SAVE DRAFT' }}</button>
        <button v-if="!published" :disabled="!canSave" @click="saveToServer(true)">PUBLISH</button>
        <button v-else :disabled="!canSave" @click="unpublish">UNPUBLISH</button>
        <button :disabled="blocked || uploads.uploading.value" @click="clearEditor">CLEAR</button>
      </div>
      <div class="save-status" role="status">
        <span>{{ dirty ? 'Changes not saved to server' : serverSaved ? 'Saved to server' : 'New article' }}</span>
        <span v-if="lastSaved">Local backup: {{ new Date(lastSaved).toLocaleTimeString() }}</span>
        <span v-else-if="serverSaved">Server save: {{ new Date(serverSaved).toLocaleString() }}</span>
      </div>
    </div>
    <TerminalFeedback :message="feedbackMsg" :type="feedbackType" :duration="feedbackType === 'err' ? 0 : 4000" :trigger="feedbackTrigger" />
    <RouterLink to="/admin" class="back-link">← Back to posts</RouterLink>
  </template>
</template>

<style scoped>
.editor-header { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.75rem; margin-bottom: 1.5rem; font-size: 0.8rem; overflow-wrap: anywhere; }
.publication-state { color: var(--accent); font-weight: bold; }
.editor-fields { border: 0; min-width: 0; }
.title-label { display: block; color: var(--muted); font-size: 0.75rem; margin-bottom: 0.4rem; }
.title-input { width: 100%; padding: 0.75rem; font-size: 1.4rem; font-weight: bold; margin-bottom: 1rem; }
input, textarea { background: var(--bg); color: var(--fg); border: 1px solid var(--border); font-family: inherit; min-width: 0; }
.meta-panel { padding: 0.8rem 1rem; margin-bottom: 1.5rem; border: 1px solid var(--border); }
summary { font-size: 0.8rem; font-weight: bold; }
summary span { color: var(--muted); font-weight: normal; }
.meta-row { display: flex; align-items: center; gap: 0.75rem; margin-top: 1rem; }
.meta-row label { width: 70px; flex-shrink: 0; font-size: 0.75rem; }
.meta-row input, .meta-row textarea { flex: 1; padding: 0.6rem; font-size: 0.9rem; }
.meta-row button { font-size: 0.7rem; }
.view-tabs { display: flex; border: 1px solid var(--border); border-bottom: 0; }
.view-tabs button { flex: 1; border: 0; padding: 0.75rem; font-size: 0.8rem; }
.view-tabs [aria-pressed="true"] { background: var(--fg); color: var(--bg); }
.editor-area { display: flex; height: min(68vh, 800px); min-height: 380px; border: 1px solid var(--border); }
.editor-pane, .preview-pane { flex: 1; min-width: 0; display: flex; flex-direction: column; min-height: 0; }
.mode-split .editor-pane { border-right: 1px solid var(--border); }
.editor-toolbar, .pane-label { min-height: 62px; padding: 0.5rem; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; font-size: 0.7rem; color: var(--muted); }
.toolbar-actions { display: flex; flex-wrap: wrap; gap: 4px; }
.toolbar-actions button { font-size: 0.65rem; padding: 4px 6px; border-width: 1px; }
.editor-textarea { flex: 1; min-height: 0; width: 100%; padding: 1rem; font-family: var(--font-main); font-size: 0.95rem; line-height: 1.8; resize: none; border: 0; }
.preview-scroll { flex: 1; min-height: 0; padding: 1rem; overflow: auto; }
.editor-actions { position: sticky; bottom: 0; background: var(--bg); border: 1px solid var(--border); border-top: 2px solid var(--border); padding: 0.75rem; display: flex; flex-wrap: wrap; gap: 0.75rem; justify-content: space-between; z-index: 10; margin-block: 1rem; }
.action-left { display: flex; flex-wrap: wrap; gap: 0.5rem; align-items: center; }
.action-left button { font-size: 0.75rem; }
.save-status { display: flex; flex-direction: column; font-size: 0.7rem; color: var(--muted); }
.recovery-notice { padding: 1rem; border: 2px solid var(--accent); margin-bottom: 1rem; }
.recovery-notice .state-actions { margin-top: 0.75rem; }
.error-notice { color: var(--danger); margin-bottom: 1rem; }
.back-link { display: inline-block; margin-block: 1rem; font-size: 0.85rem; color: var(--muted); }
@media (max-width: 768px) {
  .meta-row { flex-wrap: wrap; }
  .meta-row label { width: 100%; }
  .meta-row input, .meta-row textarea { flex-basis: 100%; font-size: 16px; }
  .title-input { font-size: 1.15rem; }
  summary span { display: block; }
  .editor-area { height: 60dvh; min-height: 350px; }
  .editor-textarea { font-size: 16px; padding: 0.75rem; }
  .toolbar-actions button { min-height: 36px; }
  .editor-actions { padding-bottom: max(0.75rem, env(safe-area-inset-bottom)); }
  .action-left { width: 100%; }
  .action-left button { flex: 1; padding-inline: 6px; }
}
</style>
