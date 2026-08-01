import { post, del } from './client'

export function login(username: string, password: string) {
  return post<{ token: string }>(
    '/admin/login',
    { username, password },
    { redirectOnUnauthorized: false },
  )
}

export function logout() {
  return del<{ status: string }>('/admin/logout')
}
