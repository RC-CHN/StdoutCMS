import { ref, computed, onScopeDispose } from 'vue'
import { chatStream } from '../api/chat'

const CHAT_SID_KEY = 'chat_session_id'

export interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp: number
}

export function useChat() {
  const messages = ref<ChatMessage[]>([])
  const input = ref('')
  const loading = ref(false)
  const thinking = ref(false)
  const error = ref('')
  const sid = ref(localStorage.getItem(CHAT_SID_KEY) || '')
  const controller = new AbortController()
  onScopeDispose(() => controller.abort())

  function updateSid(newSid: string) {
    if (sid.value === newSid) return
    sid.value = newSid
    localStorage.setItem(CHAT_SID_KEY, newSid)
  }

  const inputError = computed(() => new TextEncoder().encode(input.value.trim()).length > 500
    ? 'Question exceeds the 500-byte limit. Shorten it to send.'
    : '')
  const canSend = computed(() =>
    !loading.value && input.value.trim().length > 0 && !inputError.value
  )

  async function send(q: string, ctx?: string) {
    if (!q.trim()) return
    loading.value = true
    error.value = ''

    // user message
    messages.value.push({
      role: 'user',
      content: q.trim(),
      timestamp: Date.now(),
    })

    // placeholder assistant message — filled in real-time by deltas
    const aiIdx = messages.value.length
    messages.value.push({
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
    })

    await chatStream(
      { q: q.trim(), ctx, sid: sid.value || undefined },
      {
        onDelta(delta: string, newSid: string) {
          updateSid(newSid)
          messages.value[aiIdx].content += delta
        },
        onThinking(t: boolean) {
          thinking.value = t
        },
        onDone(newSid: string) {
          updateSid(newSid)
          thinking.value = false
          loading.value = false
        },
        onError(msg: string) {
          thinking.value = false
          error.value = msg
          loading.value = false
          if (messages.value[aiIdx]?.content === '') messages.value.splice(aiIdx, 1)
        },
      },
      controller.signal,
    )
  }

  function clear() {
    messages.value = []
    error.value = ''
    sid.value = ''
    localStorage.removeItem(CHAT_SID_KEY)
  }

  return {
    messages,
    input,
    loading,
    thinking,
    error,
    inputError,
    sid,
    canSend,
    send,
    clear,
  }
}
