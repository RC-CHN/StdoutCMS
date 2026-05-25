import { post } from './client'

export function uploadImage(file: File): Promise<{ url: string }> {
  const fd = new FormData()
  fd.append('file', file)
  return post<{ url: string }>('/admin/upload', fd)
}
