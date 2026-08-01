const BASE_URL = '/api/v1'

export interface RequestOptions {
  redirectOnUnauthorized?: boolean
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string | null

  constructor(message: string, status: number, code: string | null = null) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

async function readErrorResponse(res: Response): Promise<{ message: string; code: string | null }> {
  const raw = await res.text().catch(() => '')
  if (raw) {
    try {
      const data: unknown = JSON.parse(raw)
      if (data && typeof data === 'object') {
        const payload = data as Record<string, unknown>
        const message = typeof payload.error === 'string' ? payload.error : ''
        const code = typeof payload.code === 'string' ? payload.code : null
        if (message) return { message, code }
      }
    } catch {
      return { message: raw, code: null }
    }
  }

  return {
    message: res.statusText || `request failed with HTTP ${res.status}`,
    code: null,
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const headers: Record<string, string> = {}
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  const res = await fetch(BASE_URL + path, {
    method,
    headers,
    credentials: 'include', // send HttpOnly cookie set by backend
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  })

  const error = res.ok ? null : await readErrorResponse(res)

  if (res.status === 401 && options.redirectOnUnauthorized !== false) {
    sessionStorage.removeItem('blog_logged_in')
    window.location.href = '/login?reason=session_expired'
  }

  if (error) {
    throw new ApiError(error.message, res.status, error.code)
  }
  return res.json()
}

export function get<T>(path: string): Promise<T> {
  return request<T>('GET', path)
}

export function post<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
  return request<T>('POST', path, body, options)
}

export function put<T>(path: string, body?: unknown): Promise<T> {
  return request<T>('PUT', path, body)
}

export function del<T>(path: string): Promise<T> {
  return request<T>('DELETE', path)
}
