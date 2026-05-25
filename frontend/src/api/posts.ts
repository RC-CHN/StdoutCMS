import { get, post, put, del } from './client'

export interface PostPayload {
  slug: string
  title: string
  content: string
  excerpt?: string
  tags?: string[]
  author?: string
  wordCount?: number
  readTime?: string
  published?: boolean
  createdAt?: string
  updatedAt?: string
}

export function listPosts(page = 1, size = 10) {
  return get<{
    posts: PostPayload[]
    total: number
    page: number
    pageSize: number
  }>(`/posts?page=${page}&size=${size}`)
}

export function getPost(slug: string) {
  return get<PostPayload>(`/posts/${slug}`)
}

export function listPostsAdmin(page = 1, size = 10) {
  return get<{
    posts: PostPayload[]
    total: number
    page: number
    pageSize: number
  }>(`/admin/posts?page=${page}&size=${size}`)
}

export function createPost(payload: PostPayload) {
  return post<{ slug: string }>('/admin/posts', payload)
}

export function updatePost(slug: string, payload: PostPayload) {
  return put<{ slug: string; status: string }>(`/admin/posts/${slug}`, payload)
}

export function deletePost(slug: string) {
  return del<{ slug: string; status: string }>(`/admin/posts/${slug}`)
}

export function getPostAdmin(slug: string) {
  return get<PostPayload>(`/admin/posts/${slug}`)
}
