import { ref } from 'vue'
import { uploadFile, type UploadResult } from '../api/upload'

type MediaKind = UploadResult['kind']

interface MediaConfig {
  kind: MediaKind
  accept: string
  symbol: string
  placeholder: string
}

const MEDIA: MediaConfig[] = [
  { kind: 'image', accept: 'image/*', symbol: '\u25FB', placeholder: 'image' },
  { kind: 'audio', accept: 'audio/*', symbol: '\u266B', placeholder: 'audio' },
  { kind: 'video', accept: 'video/*', symbol: '\u25B6', placeholder: 'video' },
  { kind: 'file',  accept: '*/*',    symbol: '\u21E9', placeholder: 'file' },
]

/**
 * Editor toolbar — multi-kind media upload via dynamic file inputs.
 * Works on a plain textarea via cursor-aware text insertion.
 */
export function useEditorToolbar(textarea: () => HTMLTextAreaElement | null) {
  const uploading = ref(false)
  const uploadKind = ref<MediaKind | null>(null)

  function getTa(): HTMLTextAreaElement | null {
    return textarea()
  }

  function fireInput(ta: HTMLTextAreaElement) {
    ta.dispatchEvent(new Event('input', { bubbles: true }))
  }

  /** Trigger file picker for a media kind. Creates and destroys the input. */
  function uploadMedia(kind: MediaKind) {
    const cfg = MEDIA.find(m => m.kind === kind)
    if (!cfg) return

    const input = document.createElement('input')
    input.type = 'file'
    input.accept = cfg.accept
    input.onchange = (e) => handleFileSelected(e, cfg)
    input.click()
  }

  async function handleFileSelected(event: Event, cfg: MediaConfig) {
    const input = event.target as HTMLInputElement
    const file = input.files?.[0]
    if (!file) { input.remove(); return }

    uploading.value = true
    uploadKind.value = cfg.kind

    const ta = getTa()
    const uuid = 'upload-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8)
    const placeholder = `![${cfg.placeholder}...](uploading-${uuid})`

    if (ta) {
      const start = ta.selectionStart
      ta.value = ta.value.slice(0, start) + placeholder + ta.value.slice(ta.selectionEnd)
      ta.selectionStart = ta.selectionEnd = start + placeholder.length
      fireInput(ta)
    }

    try {
      const res = await uploadFile(file)
      if (ta) {
        const alt = file.name.replace(/\.[^.]+$/, '')
        const md = `![${cfg.kind}:${alt}](${res.url})`
        ta.value = ta.value.replace(placeholder, md)
        fireInput(ta)
      }
    } catch {
      // placeholder stays; user can fix manually
    } finally {
      uploading.value = false
      uploadKind.value = null
      input.remove()
    }
  }

  return { uploading, uploadKind, uploadMedia }
}
