import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { useChat } from '../composables/useChat'
import { chatStream } from './chat'

beforeEach(() => {
  vi.stubGlobal('document', { cookie: '' })
  vi.stubGlobal('localStorage', { getItem: () => null })
})
afterEach(() => vi.unstubAllGlobals())

async function readChat(events: object[]) {
  const stream = events.map(event => 'data: ' + JSON.stringify(event) + '\n\n').join('')
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(stream)))
  const callbacks = { onDelta: vi.fn(), onThinking: vi.fn(), onError: vi.fn(), onDone: vi.fn() }
  await chatStream({ q: 'Hello' }, callbacks)
  return callbacks
}

describe('chat status messages', () => {
  it('stops the pending request when the chat component is disposed', async () => {
    let signal: AbortSignal | undefined
    vi.stubGlobal('fetch', vi.fn((_url: string, options: RequestInit) => {
      signal = options.signal ?? undefined
      return new Promise<Response>((_resolve, reject) => {
        signal?.addEventListener('abort', () => reject(signal?.reason), { once: true })
      })
    }))
    const scope = effectScope()
    const chat = scope.run(useChat)!
    const reply = chat.send('Hello')
    expect(signal?.aborted).toBe(false)
    scope.stop()
    expect(signal?.aborted).toBe(true)
    await reply
    expect(chat.error.value).toBe('')
  })

  it('keeps a server error separate from a partial AI reply', async () => {
    const callbacks = await readChat([
      { delta: 'Answer', sid: 'chat_test' },
      { error: 'Chat is unavailable.', delta: '[ ERROR ] Chat is unavailable.', sid: 'chat_test' },
      { done: true, sid: 'chat_test' },
    ])
    expect(callbacks.onDelta).toHaveBeenCalledExactlyOnceWith('Answer', 'chat_test')
    expect(callbacks.onError).toHaveBeenCalledExactlyOnceWith('Chat is unavailable.')
    expect(callbacks.onDone).not.toHaveBeenCalled()
  })

  it('reports a reply stream that closed before completion', async () => {
    const callbacks = await readChat([{ delta: 'Answer', sid: 'chat_test' }])
    expect(callbacks.onError).toHaveBeenCalledExactlyOnceWith('The reply was interrupted. Send your question again.')
    expect(callbacks.onDone).not.toHaveBeenCalled()
  })

  it('completes normal replies without an error', async () => {
    const callbacks = await readChat([{ done: true, sid: 'chat_test' }])
    expect(callbacks.onDone).toHaveBeenCalledExactlyOnceWith('chat_test')
    expect(callbacks.onError).not.toHaveBeenCalled()
  })
})
