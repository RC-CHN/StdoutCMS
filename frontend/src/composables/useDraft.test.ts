import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick, watch } from 'vue'
import { draftKey, draftText, emptyDraft, useDraft } from './useDraft'

let values: Map<string, string>
beforeEach(() => {
  values = new Map()
  vi.stubGlobal('localStorage', {
    getItem: vi.fn((key: string) => values.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => values.set(key, value)),
    removeItem: vi.fn((key: string) => values.delete(key)),
  })
})
afterEach(() => vi.unstubAllGlobals())

describe('article backups', () => {
  it('keeps new articles and different article backups separate', () => {
    for (const slug of [undefined, 'one', 'two']) {
      const editor = useDraft(slug)
      editor.draft.value.content = slug || 'new article'
      expect(editor.save()).toBe(true)
    }
    expect(useDraft('one').load()?.content).toBe('one')
    expect(useDraft('two').load()?.content).toBe('two')
    expect(useDraft().load()?.content).toBe('new article')
  })

  it('offers a backup without overwriting server content until restore is chosen', () => {
    values.set(draftKey('one'), JSON.stringify({ ...emptyDraft(), slug: 'one', content: 'local work', savedAt: '2026-10-08T10:00:00Z' }))
    const editor = useDraft('one')
    editor.draft.value.content = 'server version'
    const backup = editor.load()
    expect(editor.draft.value.content).toBe('server version')
    editor.restore(backup!)
    expect(editor.draft.value.content).toBe('local work')
    expect(editor.lastSaved.value).toBe(backup!.savedAt)
  })

  it('does not trigger the content watcher again when writing a backup', async () => {
    const editor = useDraft('one')
    const changed = vi.fn()
    const stop = watch(() => draftText(editor.draft.value), changed)
    editor.draft.value.content = 'updated'
    await nextTick()
    editor.save()
    await nextTick()
    expect(changed).toHaveBeenCalledTimes(1)
    stop()
  })

  it('removes the backup after saving without clearing the form or another article', () => {
    const editor = useDraft('one')
    editor.draft.value.content = 'keep this on screen'
    editor.save()
    values.set(draftKey('two'), 'other backup')
    editor.discard()
    expect(editor.draft.value.content).toBe('keep this on screen')
    expect(values.has(draftKey('one'))).toBe(false)
    expect(values.get(draftKey('two'))).toBe('other backup')
  })

  it('keeps a legacy backup available to recover and never puts it into a different existing article', () => {
    values.set('blog_editor_draft', JSON.stringify({ ...emptyDraft(), slug: 'one', content: 'legacy' }))
    expect(useDraft('two').load()).toBeNull()
    expect(useDraft('one').load()?.content).toBe('legacy')
    expect(values.has('blog_editor_draft')).toBe(false)
    expect(values.has(draftKey('one'))).toBe(true)
  })

  it('reports storage failures without claiming a backup succeeded', () => {
    vi.mocked(localStorage.setItem).mockImplementation(() => { throw new Error('quota exceeded') })
    const editor = useDraft()
    expect(editor.save()).toBe(false)
    expect(editor.lastSaved.value).toBe('')
    expect(editor.storageError.value).toContain('failed')
  })
})
