import { beforeEach, describe, expect, it, vi } from 'vitest'
import { uploadFile } from '../api/upload'
import { useEditorUploads } from './useEditorUploads'
vi.mock('../api/upload', () => ({ uploadFile: vi.fn() }))
beforeEach(() => vi.resetAllMocks())

function editor() {
  const field = { value: 'before after', selectionStart: 7, selectionEnd: 7, dispatchEvent: vi.fn() }
  return { field, uploads: useEditorUploads(() => field as unknown as HTMLTextAreaElement) }
}
const file = new File(['image'], 'my-image.png', { type: 'image/png' })

describe('editor uploads', () => {
  it('replaces the placeholder and keeps surrounding text on success', async () => {
    vi.mocked(uploadFile).mockResolvedValue({ url: '/img/images/test.png', kind: 'image' })
    const { field, uploads } = editor()
    await uploads.upload(file, 'image')
    expect(field.value).toBe('before ![image:my-image.png](/img/images/test.png)after')
    expect(uploads.uploading.value).toBe(false)
    expect(uploads.failures.value).toHaveLength(0)
  })

  it('offers retry after a failure and replaces the failure marker on success', async () => {
    vi.mocked(uploadFile).mockRejectedValueOnce(new Error('network unavailable'))
    const { field, uploads } = editor()
    await uploads.upload(file, 'image')
    expect(field.value).not.toContain('uploading-')
    expect(field.value).toContain('[upload failed:')
    expect(uploads.failures.value[0]?.error).toBe('network unavailable')
    vi.mocked(uploadFile).mockResolvedValue({ url: '/img/retried.png', kind: 'image' })
    await uploads.retry(uploads.failures.value[0]!.id)
    expect(field.value).toBe('before ![image:my-image.png](/img/retried.png)after')
    expect(uploads.failures.value).toHaveLength(0)
  })

  it('removes only the failed upload when dismissed', async () => {
    vi.mocked(uploadFile).mockRejectedValue(new Error('too large'))
    const { field, uploads } = editor()
    await uploads.upload(file, 'image')
    uploads.remove(uploads.failures.value[0]!.id)
    expect(field.value).toBe('before after')
    expect(uploads.failures.value).toHaveLength(0)
  })

  it('does not write late upload results into an editor after it is disposed', async () => {
    let resolve!: (value: { url: string; kind: 'image' }) => void
    vi.mocked(uploadFile).mockReturnValue(new Promise(done => { resolve = done }))
    const { field, uploads } = editor()
    const upload = uploads.upload(file, 'image')
    uploads.dispose()
    field.value = 'another article'
    resolve({ url: '/img/late.png', kind: 'image' })
    await upload
    expect(field.value).toBe('another article')
  })

  it('treats replacement characters in filenames as literal text', async () => {
    vi.mocked(uploadFile).mockResolvedValue({ url: '/img/test.png', kind: 'image' })
    const { field, uploads } = editor()
    await uploads.upload(new File(['a'], '$&.png', { type: 'image/png' }), 'image')
    expect(field.value).toBe('before ![image:$&.png](/img/test.png)after')
  })
})
