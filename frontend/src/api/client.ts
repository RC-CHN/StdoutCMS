const BASE_URL = '/api/v1'

function getToken(): string {
  // 优先 cookie，fallback localStorage
  const match = document.cookie.match(/(?:^| )blog_session_token=([^;]+)/)
  if (match) return match[1]
  return localStorage.getItem('blog_session_token') || ''
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const headers: Record<string, string> = {}
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  const token = getToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(BASE_URL + path, {
    method,
    headers,
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    const data = await res.json().catch(() => ({}))
    throw new Error(data.error || res.statusText)
  }
  return res.json()
}

export function get<T>(path: string): Promise<T> {
  return request<T>('GET', path)
}

export function post<T>(path: string, body?: unknown): Promise<T> {
  return request<T>('POST', path, body)
}

export function put<T>(path: string, body?: unknown): Promise<T> {
  return request<T>('PUT', path, body)
}

export function del<T>(path: string): Promise<T> {
  return request<T>('DELETE', path)
}
