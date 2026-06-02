<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getAboutAdmin, updateAbout } from '../../api/about'
import type { AboutPayload } from '../../api/about'
import AdminNav from '../../components/AdminNav.vue'
import ArticleRenderer from '../../components/ArticleRenderer.vue'
import TerminalFeedback from '../../components/TerminalFeedback.vue'
import { useImagePaste } from '../../composables/useImagePaste'

const about = ref<AboutPayload>({ title: '', content: '' })
const status = ref('')
const saving = ref(false)
const feedback = ref<{ msg: string; type: 'ok' | 'err' } | null>(null)
const feedbackTrigger = ref(0)

function showFeedback(msg: string, type: 'ok' | 'err') {
  feedback.value = { msg, type }
  feedbackTrigger.value++
}

// image paste upload
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const { attach, detach } = useImagePaste(() => textareaRef.value)

onMounted(() => attach())
onUnmounted(() => detach())

onMounted(async () => {
  try {
    const data = await getAboutAdmin()
    about.value = data
    status.value = 'loaded: about'
  } catch {
    status.value = 'creating new about'
  }
})

async function handleSave() {
  saving.value = true
  try {
    await updateAbout(about.value)
    status.value = 'saved: about'
    showFeedback('about saved', 'ok')
  } catch (e: any) {
    status.value = 'ERROR: ' + (e.message || 'save failed')
    showFeedback(e.message || 'save failed', 'err')
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
      <input v-model="about.title" type="text" placeholder="About page title" />
    </div>
  </div>

  <div class="editor-split">
    <div class="editor-pane">
      <div class="pane-label">RAW // MARKDOWN</div>
      <textarea
        v-model="about.content"
        class="editor-textarea"
          ref="textareaRef"
        placeholder="Write your about page in markdown..."
        spellcheck="false"
      />
    </div>
    <div class="preview-pane">
      <div class="pane-label">PREVIEW // RENDERED</div>
      <div class="preview-scroll">
        <ArticleRenderer :content="about.content" />
      </div>
    </div>
  </div>

  <div class="editor-actions">
    <button class="btn" :disabled="saving" @click="handleSave">
      {{ saving ? '[ SAVING... ]' : '[ SAVE ]' }}
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
