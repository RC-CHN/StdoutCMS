import type { EditorUploads, MediaKind } from './useEditorUploads'

const accept: Record<MediaKind, string> = { image: 'image/*', audio: 'audio/*', video: 'video/*', file: '*/*' }

export function useEditorToolbar(uploads: EditorUploads) {
  function uploadMedia(kind: MediaKind) {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = accept[kind]
    input.onchange = () => {
      const file = input.files?.[0]
      if (file) void uploads.upload(file, kind)
      input.remove()
    }
    input.click()
  }
  return { uploadMedia, uploading: uploads.uploading, uploadKind: uploads.uploadKind }
}
