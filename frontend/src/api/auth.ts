import { post, del } from './client'

export function login(username: string, password: string) {
  return post<{ token: string }>('/admin/login', { username, password })
}

export function logout() {
  return del<{ status: string }>('/admin/logout')
}
