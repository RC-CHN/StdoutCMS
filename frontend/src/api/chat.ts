import { post } from './client'

export interface ChatReq {
  q: string
  ctx?: string
  sid?: string
}

export interface ChatRes {
  a: string
  sid: string
}

/** Non-streaming fallback — calls the same endpoint but returns JSON. */
export function chat(body: ChatReq): Promise<ChatRes> {
  return post<ChatRes>('/ai/chat', body)
}

/**
 * Streaming chat — reads SSE events from the AI endpoint.
 *
 * SSE event format:
 *   data: {"delta":"Hello","sid":"chat_xxx"}
 *   data: {"done":true,"sid":"chat_xxx"}
 */
export async function chatStream(
  body: ChatReq,
  cbs: {
    onDelta: (delta: string, sid: string) => void
    onThinking: (thinking: boolean) => void
    onDone: (sid: string) => void
    onError: (msg: string) => void
  },
): Promise<void> {
  const BASE_URL = '/api/v1'
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'text/event-stream',
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(BASE_URL + '/ai/chat', {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  })

  if (!res.ok) {
    const data = await res.json().catch(() => ({}))
    cbs.onError(data.error || res.statusText)
    return
  }

  const reader = res.body?.getReader()
  if (!reader) {
    cbs.onError('stream not supported')
    return
  }

  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        try {
          const ev = JSON.parse(line.slice(6))
          if (ev.done) {
            cbs.onDone(ev.sid)
          } else if (typeof ev.thinking === 'boolean') {
            cbs.onThinking(ev.thinking)
          } else if (ev.delta) {
            cbs.onDelta(ev.delta, ev.sid)
          }
        } catch {
          // skip malformed JSON
        }
      }
    }
  } catch (e: any) {
    cbs.onError(e.message || 'stream error')
  }
}

function getToken(): string {
  const match = document.cookie.match(/(?:^| )blog_session_token=([^;]+)/)
  if (match) return match[1]
  return localStorage.getItem('blog_session_token') || ''
}

// ---- Meta generation (admin) ----

export interface GenerateMetaReq {
  title: string
  content: string
}

export interface GenerateMetaRes {
  slug: string | null
  tags: string | null
  excerpt: string | null
}

/** Calls the admin AI endpoint to generate slug / tags / excerpt from title + content. */
export function generateMeta(body: GenerateMetaReq): Promise<GenerateMetaRes> {
  return post<GenerateMetaRes>('/admin/ai/generate-meta', body)
}
