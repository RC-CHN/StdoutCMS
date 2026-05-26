import { uploadImage } from '../api/upload'

/**
 * Handle paste event: if clipboard contains an image file,
 * insert a placeholder and upload it asynchronously.
 *
 * @param textarea - the target textarea ref
 * @returns a cleanup function to remove the event listener
 */
export function useImagePaste(textarea: () => HTMLTextAreaElement | null) {
  async function onPaste(e: ClipboardEvent) {
    const items = e.clipboardData?.items
    if (!items) return

    for (const item of items) {
      if (!item.type.startsWith('image/')) continue

      e.preventDefault()
      const file = item.getAsFile()
      if (!file) continue

      const uuid = 'upload-' + Date.now() + '-' + Math.random().toString(36).slice(2, 8)
      const placeholder = `![image](uploading-${uuid})`

      insertAtCursor(placeholder)

      try {
        const data = await uploadImage(file)
        replacePlaceholder(uuid, data.url)
      } catch {
        // 失败：占位符保留，用户手动处理
      }
      break
    }
  }

  function insertAtCursor(text: string) {
    const ta = textarea()
    if (!ta) return
    const start = ta.selectionStart
    const end = ta.selectionEnd
    ta.value = ta.value.slice(0, start) + text + ta.value.slice(end)
    ta.selectionStart = ta.selectionEnd = start + text.length
    ta.focus()
    ta.dispatchEvent(new Event('input', { bubbles: true }))
  }

  function replacePlaceholder(uuid: string, url: string) {
    const ta = textarea()
    if (!ta) return
    const pattern = `![image](uploading-${uuid})`
    ta.value = ta.value.replace(pattern, `![image](${url})`)
    ta.dispatchEvent(new Event('input', { bubbles: true }))
  }

  function attach() {
    document.addEventListener('paste', onPaste)
  }

  function detach() {
    document.removeEventListener('paste', onPaste)
  }

  return { attach, detach }
}
