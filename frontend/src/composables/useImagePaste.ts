import type { EditorUploads } from './useEditorUploads'

export function useImagePaste(textarea: () => HTMLTextAreaElement | null, uploads: EditorUploads) {
  let attached: HTMLTextAreaElement | null = null
  function onPaste(event: ClipboardEvent) {
    for (const item of event.clipboardData?.items ?? []) {
      if (!item.type.startsWith('image/')) continue
      const file = item.getAsFile()
      if (!file) continue
      event.preventDefault()
      void uploads.upload(file, 'image')
      break
    }
  }
  function detach() {
    attached?.removeEventListener('paste', onPaste)
    attached = null
  }
  function attach() {
    detach()
    attached = textarea()
    attached?.addEventListener('paste', onPaste)
  }
  return { attach, detach }
}
