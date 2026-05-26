import { ref, computed } from 'vue'
import { chat } from '../api/chat'

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
  const error = ref('')
  const sid = ref(localStorage.getItem(CHAT_SID_KEY) || '')

  const canSend = computed(() =>
    !loading.value && input.value.trim().length > 0 && input.value.length <= 500
  )

  async function send(q: string, ctx?: string) {
    if (!q.trim()) return
    loading.value = true
    error.value = ''

    messages.value.push({
      role: 'user',
      content: q.trim(),
      timestamp: Date.now(),
    })

    try {
      const res = await chat({ q: q.trim(), ctx, sid: sid.value || undefined })

      sid.value = res.sid
      localStorage.setItem(CHAT_SID_KEY, res.sid)

      messages.value.push({
        role: 'assistant',
        content: res.a,
        timestamp: Date.now(),
      })
    } catch (e: any) {
      error.value = e.message || 'connection error'
    } finally {
      loading.value = false
    }
  }

  function clear() {
    messages.value = []
    sid.value = ''
    localStorage.removeItem(CHAT_SID_KEY)
  }

  return {
    messages,
    input,
    loading,
    error,
    sid,
    canSend,
    send,
    clear,
  }
}
