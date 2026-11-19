<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getAboutAdmin, updateAbout } from '../../api/about'
import type { AboutPayload } from '../../api/about'
import AdminNav from '../../components/AdminNav.vue'
import ArticleRenderer from '../../components/ArticleRenderer.vue'
import MobileArticleRenderer from '../../components/MobileArticleRenderer.vue'
import PageState from '../../components/PageState.vue'
import { ApiError } from '../../api/client'
import { useBreakpoint } from '../../composables/useBreakpoint'
import TerminalFeedback from '../../components/TerminalFeedback.vue'
import UploadStatus from '../../components/UploadStatus.vue'
import { useEditorUploads } from '../../composables/useEditorUploads'
import { useImagePaste } from '../../composables/useImagePaste'

const about = ref<AboutPayload>({ title: '', content: '' })
const isNew = ref(false)
const saving = ref(false)
const loading = ref(true)
const loadError = ref('')
const { isMobile } = useBreakpoint()
const preview = ref(false)
const feedback = ref<{ msg: string; type: 'ok' | 'err' } | null>(null)
const feedbackTrigger = ref(0)

function showFeedback(msg: string, type: 'ok' | 'err') {
  feedback.value = { msg, type }
  feedbackTrigger.value++
}

// image paste upload
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const uploads = useEditorUploads(() => textareaRef.value)
const { attach, detach } = useImagePaste(() => textareaRef.value, uploads)

onMounted(() => attach())
onUnmounted(() => { detach(); uploads.dispose() })

async function loadAbout() {
  loading.value = true
  loadError.value = ''
  try {
    const data = await getAboutAdmin()
    about.value = data
    isNew.value = false
  } catch (cause) {
    if (cause instanceof ApiError && cause.status === 404) isNew.value = true
    else loadError.value = 'Could not load the About page.'
  } finally {
    loading.value = false
  }
}
onMounted(loadAbout)

async function handleSave() {
  if (loading.value || loadError.value || saving.value || uploads.uploading.value || uploads.failures.value.length) return
  if (/uploading-|\[upload failed:/.test(about.value.content)) { showFeedback('Finish or remove incomplete uploads before saving.', 'err'); return }
  saving.value = true
  try {
    await updateAbout(about.value)
    isNew.value = false
    showFeedback('Saved about.md.', 'ok')
  } catch (e: any) {
    showFeedback(e.message || 'Could not save the About page.', 'err')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AdminNav />

  <div class="editor-header">
    <div class="editor-prompt">
      root@stdoutcms:~/admin# <span class="muted">vim about.md</span>
    </div>
  </div>

  <PageState v-if="loading" mode="loading" message="Loading the About page…" />
  <PageState v-else-if="loadError" mode="error" :message="loadError" retry @retry="loadAbout" />
  <div class="meta-panel">
    <div class="meta-row">
      <label for="about-title">TITLE</label>
      <input id="about-title" v-model="about.title" :disabled="loading || !!loadError || saving" type="text" placeholder="About page title" />
    </div>
  </div>

  <div v-if="isMobile" class="state-actions" role="group" aria-label="About editor view">
    <button :aria-pressed="!preview" @click="preview = false">EDIT</button>
    <button :aria-pressed="preview" @click="preview = true">PREVIEW</button>
  </div>
  <div class="editor-split">
    <div v-show="!isMobile || !preview" class="editor-pane">
      <div class="pane-label">MARKDOWN</div>
      <textarea
        v-model="about.content"
        :disabled="loading || !!loadError || saving"
        aria-label="About content"
        class="editor-textarea"
        ref="textareaRef"
        placeholder="Write your About page in Markdown…"
        spellcheck="false"
      />
    </div>
    <div v-show="!isMobile || preview" class="preview-pane">
      <div class="pane-label">PREVIEW</div>
      <div class="preview-scroll">
        <KeepAlive :max="1">
          <component v-if="!loading && !loadError && (!isMobile || preview)" :is="isMobile ? MobileArticleRenderer : ArticleRenderer" :content="about.content" />
        </KeepAlive>
      </div>
    </div>
  </div>

  <UploadStatus :uploads="uploads" />
  <div class="editor-actions">
    <button class="btn" :disabled="loading || !!loadError || saving || uploads.uploading.value || !!uploads.failures.value.length" @click="handleSave">
      {{ saving ? 'SAVING…' : isNew ? 'CREATE PAGE' : 'SAVE CHANGES' }}
    </button>
  </div>

  <TerminalFeedback
    :message="feedback?.msg ?? null"
    :type="feedback?.type ?? 'info'"
    :duration="feedback?.type === 'err' ? 0 : 4000"
    :trigger="feedbackTrigger"
  />
</template>

<style scoped>
.editor-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; flex-wrap: wrap; gap: 0.5rem; }
.editor-prompt { font-weight: bold; font-size: 0.95rem; }
.muted { color: var(--muted); }
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

.meta-row input { min-width: 0; }
.editor-split { height: 65dvh; min-height: 350px; }
.preview-scroll, .editor-textarea { min-height: 0; }
.editor-actions { position: sticky; bottom: 0; background: var(--bg); padding: 0.75rem; z-index: 10; }
@media (max-width: 768px) {
  .meta-row { flex-wrap: wrap; }
  .meta-row label { width: 100%; }
  .meta-row input { width: 100%; font-size: 16px; }
  .editor-split { flex-direction: row; }
  .editor-textarea { font-size: 16px; }
  .state-actions { margin-bottom: 0.75rem; }
  .state-actions [aria-pressed="true"] { background: var(--fg); color: var(--bg); }
}
</style>
