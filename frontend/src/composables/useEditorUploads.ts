import { computed, ref } from 'vue'
import { uploadFile, type UploadResult } from '../api/upload'

export type MediaKind = UploadResult['kind']
interface UploadJob {
  id: string
  file: File
  kind: MediaKind
  marker: string
  state: 'uploading' | 'failed'
  error: string
}

export function useEditorUploads(textarea: () => HTMLTextAreaElement | null) {
  const jobs = ref<UploadJob[]>([])
  const pending = computed(() => jobs.value.filter(job => job.state === 'uploading'))
  const failures = computed(() => jobs.value.filter(job => job.state === 'failed'))
  const uploading = computed(() => pending.value.length > 0)
  const uploadKind = computed(() => pending.value[0]?.kind ?? null)
  let disposed = false

  function replace(marker: string, text: string) {
    const ta = textarea()
    if (disposed || !ta || !ta.value.includes(marker)) return
    const offset = ta.value.indexOf(marker)
    const adjust = (position: number) => position <= offset ? position : position >= offset + marker.length
      ? position + text.length - marker.length : offset + text.length
    const start = adjust(ta.selectionStart)
    const end = adjust(ta.selectionEnd)
    ta.value = ta.value.replace(marker, () => text)
    ta.selectionStart = start
    ta.selectionEnd = end
    ta.dispatchEvent(new Event('input', { bubbles: true }))
  }

  function insert(text: string) {
    const ta = textarea()
    if (!ta || disposed) return false
    const start = ta.selectionStart
    ta.value = ta.value.slice(0, start) + text + ta.value.slice(ta.selectionEnd)
    ta.selectionStart = ta.selectionEnd = start + text.length
    ta.dispatchEvent(new Event('input', { bubbles: true }))
    return true
  }

  async function run(job: UploadJob) {
    try {
      const result = await uploadFile(job.file)
      const label = job.file.name.replace(/[\[\]\\\r\n]/g, '_')
      replace(job.marker, `![${result.kind}:${label}](${result.url})`)
      jobs.value = jobs.value.filter(item => item.id !== job.id)
    } catch (cause) {
      if (disposed) return
      const failed = `[upload failed: ${job.file.name.replace(/[\]\r\n]/g, '_')} (${job.id})]`
      replace(job.marker, failed)
      job.marker = failed
      job.state = 'failed'
      job.error = cause instanceof Error ? cause.message : 'Upload failed'
    }
  }

  async function upload(file: File, kind: MediaKind) {
    if (disposed) return
    const id = crypto.randomUUID()
    const marker = `![${kind}:uploading](uploading-${id})`
    if (!insert(marker)) return
    jobs.value.push({ id, file, kind, marker, state: 'uploading', error: '' })
    await run(jobs.value[jobs.value.length - 1]!)
  }

  async function retry(id: string) {
    const job = jobs.value.find(item => item.id === id)
    if (!job || job.state !== 'failed' || disposed) return
    const marker = `![${job.kind}:uploading](uploading-${job.id})`
    if (textarea()?.value.includes(job.marker)) replace(job.marker, marker)
    else if (!insert(marker)) return
    job.marker = marker
    job.state = 'uploading'
    job.error = ''
    await run(job)
  }

  function remove(id: string) {
    const job = jobs.value.find(item => item.id === id)
    if (!job || job.state === 'uploading') return
    replace(job.marker, '')
    jobs.value = jobs.value.filter(item => item.id !== id)
  }

  return { uploading, uploadKind, pending, failures, upload, retry, remove, dispose: () => { disposed = true } }
}

export type EditorUploads = ReturnType<typeof useEditorUploads>
