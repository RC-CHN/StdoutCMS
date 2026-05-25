import { get, post, put, del } from './client'

export interface ProjectPayload {
  id: number
  name: string
  description: string
  lang: string
  status: string
  url: string
  sortOrder: number
}

export function listProjects() {
  return get<{ projects: ProjectPayload[] }>('/projects')
}

export function listProjectsAdmin() {
  return get<{ projects: ProjectPayload[] }>('/admin/projects')
}

export function createProject(payload: ProjectPayload) {
  return post<{ status: string }>('/admin/projects', payload)
}

export function updateProject(id: number, payload: ProjectPayload) {
  return put<{ status: string }>(`/admin/projects/${id}`, payload)
}

export function deleteProject(id: number) {
  return del<{ status: string }>(`/admin/projects/${id}`)
}
