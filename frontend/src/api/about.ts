import { get, put } from './client'

export interface AboutPayload {
  title: string
  content: string
  updatedAt?: string
}

export function getAbout() {
  return get<AboutPayload>('/about')
}

export function getAboutAdmin() {
  return get<AboutPayload>('/admin/about')
}

export function updateAbout(payload: AboutPayload) {
  return put<{ status: string }>('/admin/about', payload)
}
