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
  signal?: AbortSignal,
): Promise<void> {
  const BASE_URL = '/api/v1'
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'text/event-stream',
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  let res: Response
  try {
    res = await fetch(BASE_URL + '/ai/chat', {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
      signal,
    })
  } catch {
    if (!signal?.aborted) cbs.onError('Could not reach the chat service. Check your connection and try again.')
    return
  }

  if (!res.ok) {
    if (res.status === 429) cbs.onError('Daily chat limit reached. Try again tomorrow.')
    else if (res.status === 400) cbs.onError('Use a shorter question and try again.')
    else cbs.onError('Chat is unavailable. Try again later.')
    return
  }

  const reader = res.body?.getReader()
  if (!reader) {
    cbs.onError('Could not open the reply stream. Try again.')
    return
  }

  const decoder = new TextDecoder()
  let buffer = ''
  const interrupted = 'The reply was interrupted. Send your question again.'

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) {
        cbs.onError(interrupted)
        return
      }

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        try {
          const ev = JSON.parse(line.slice(6))
          if (typeof ev.error === 'string' && ev.error) {
            cbs.onError(ev.error)
            return
          } else if (ev.done) {
            cbs.onDone(ev.sid)
            return
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
  } catch {
    if (!signal?.aborted) cbs.onError(interrupted)
  } finally {
    reader.releaseLock()
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
export function generateMeta(body: GenerateMetaReq, signal?: AbortSignal): Promise<GenerateMetaRes> {
  return post<GenerateMetaRes>('/admin/ai/generate-meta', body, { signal })
}
