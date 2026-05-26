import { get } from './client'

export interface MetaResponse {
  app: string
  version: string
  uptime: string
  posts: number
  projects: number
  goVersion: string
}

export function fetchMeta() {
  return get<MetaResponse>('/meta')
}
