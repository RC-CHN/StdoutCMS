import { ref, onMounted } from 'vue'

const DRAFT_KEY = 'blog_editor_draft'

export interface Draft {
  title: string
  slug: string
  tags: string
  excerpt: string
  content: string
  savedAt: string
}

const emptyDraft: Draft = {
  title: '',
  slug: '',
  tags: '',
  excerpt: '',
  content: '',
  savedAt: '',
}

export function useDraft() {
  const draft = ref<Draft>({ ...emptyDraft })
  const lastSaved = ref('')

  function load() {
    try {
      const raw = localStorage.getItem(DRAFT_KEY)
      if (raw) {
        const parsed = JSON.parse(raw) as Draft
        draft.value = parsed
        lastSaved.value = parsed.savedAt
      }
    } catch {
      // ignore
    }
  }

  function save() {
    const now = new Date().toISOString()
    draft.value.savedAt = now
    localStorage.setItem(DRAFT_KEY, JSON.stringify(draft.value))
    lastSaved.value = now
  }

  function clear() {
    draft.value = { ...emptyDraft }
    lastSaved.value = ''
    localStorage.removeItem(DRAFT_KEY)
  }

  function restoreFromPost(post: { title: string; slug: string; tags: string[]; excerpt: string; content: string }) {
    draft.value = {
      title: post.title,
      slug: post.slug,
      tags: post.tags.join(', '),
      excerpt: post.excerpt,
      content: post.content,
      savedAt: '',
    }
  }

  onMounted(load)

  // auto-save on change (debounced by caller)
  return {
    draft,
    lastSaved,
    save,
    clear,
    load,
    restoreFromPost,
  }
}
