import { post } from './client'

export interface UploadResult {
  url: string
  kind: 'image' | 'audio' | 'video' | 'file'
}

export function uploadFile(file: File): Promise<UploadResult> {
  const fd = new FormData()
  fd.append('file', file)
  return post<UploadResult>('/admin/upload', fd)
}
