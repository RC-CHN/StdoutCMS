import { afterEach, describe, expect, it, vi } from 'vitest'
import { login } from './auth'
import { ApiError, get } from './client'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('API error handling', () => {
  it('returns a 401 login error without invoking the session redirect path', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ error: 'invalid credentials' }),
      {
        status: 401,
        statusText: 'Unauthorized',
        headers: { 'Content-Type': 'application/json' },
      },
    ))
    vi.stubGlobal('fetch', fetchMock)

    await expect(login('root', 'wrong-password')).rejects.toMatchObject({
      name: 'ApiError',
      message: 'invalid credentials',
      status: 401,
    })
    expect(fetchMock).toHaveBeenCalledOnce()
  })

  it('preserves the HTTP status and server error code', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ error: 'temporarily unavailable', code: 'AUTH_BACKEND_DOWN' }),
      {
        status: 503,
        statusText: 'Service Unavailable',
        headers: { 'Content-Type': 'application/json' },
      },
    )))

    const failure = await get('/broken').catch((cause: unknown) => cause)

    expect(failure).toBeInstanceOf(ApiError)
    expect(failure).toMatchObject({
      message: 'temporarily unavailable',
      status: 503,
      code: 'AUTH_BACKEND_DOWN',
    })
  })
})
