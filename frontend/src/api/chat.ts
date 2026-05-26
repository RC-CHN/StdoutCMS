import { post } from './client'

export interface ChatReq {
  q: string
  ctx?: string
  sid?: string
}

export interface ChatRes {
  a: string
  sid: string
}

export function chat(body: ChatReq): Promise<ChatRes> {
  return post<ChatRes>('/ai/chat', body)
}
