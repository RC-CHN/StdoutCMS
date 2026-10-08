import { ref } from 'vue'

const LEGACY_KEY = 'blog_editor_draft'
export interface Draft {
  title: string
  slug: string
  tags: string
  excerpt: string
  content: string
  savedAt: string
}

export const emptyDraft = (): Draft => ({ title: '', slug: '', tags: '', excerpt: '', content: '', savedAt: '' })
export const draftKey = (slug?: string) => `${LEGACY_KEY}:${slug ? `post:${slug}` : 'new'}`
export const draftText = (draft: Draft) => JSON.stringify([draft.title, draft.slug, draft.tags, draft.excerpt, draft.content])

// Reading a backup never overwrites the editor; the author chooses whether to restore it.
export function useDraft(slug?: string) {
  const key = draftKey(slug)
  const draft = ref<Draft>(emptyDraft())
  const lastSaved = ref('')
  const storageError = ref('')

  function load(): Draft | null {
    try {
      const current = localStorage.getItem(key)
      const raw = current || localStorage.getItem(LEGACY_KEY)
      if (!raw) return null
      const value = JSON.parse(raw)
      if (!value || Object.keys(emptyDraft()).some(field => typeof value[field] !== 'string')) return null
      if (!current && slug && value.slug !== slug) return null
      if (!current) {
        localStorage.setItem(key, raw)
        localStorage.removeItem(LEGACY_KEY)
      }
      return value as Draft
    } catch {
      storageError.value = 'Local backup is unavailable. Save to the server before leaving.'
      return null
    }
  }

  function save(): boolean {
    try {
      const savedAt = new Date().toISOString()
      localStorage.setItem(key, JSON.stringify({ ...draft.value, savedAt }))
      lastSaved.value = savedAt
      storageError.value = ''
      return true
    } catch {
      storageError.value = 'Local backup failed. Save to the server before leaving.'
      return false
    }
  }

  function discard(): boolean {
    try {
      localStorage.removeItem(key)
      lastSaved.value = ''
      storageError.value = ''
      return true
    } catch {
      storageError.value = 'Could not remove the local backup.'
      return false
    }
  }

  function restore(value: Draft) {
    draft.value = { ...value }
    lastSaved.value = value.savedAt
  }

  return { draft, lastSaved, storageError, load, save, discard, restore }
}
