import { ref, computed } from 'vue'

const SESSION_KEY = 'blog_session_token'
const TOKEN_EXPIRE_DAYS = 7

const token = ref<string>('')
const isLoggedIn = computed(() => !!token.value)

function loadToken() {
  // 优先从 cookie 读，再 fallback 到 localStorage
  const match = document.cookie.match(new RegExp('(^| )' + SESSION_KEY + '=([^;]+)'))
  if (match) {
    token.value = match[2]
    return
  }
  const saved = localStorage.getItem(SESSION_KEY)
  if (saved) token.value = saved
}

function setToken(newToken: string) {
  token.value = newToken
  // 存 cookie，带 HttpOnly 是后端的事，前端用非 HttpOnly 的临时方案
  const expires = new Date(Date.now() + TOKEN_EXPIRE_DAYS * 864e5).toUTCString()
  document.cookie = `${SESSION_KEY}=${newToken}; expires=${expires}; path=/; SameSite=Strict`
  localStorage.setItem(SESSION_KEY, newToken)
}

function clearToken() {
  token.value = ''
  document.cookie = `${SESSION_KEY}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`
  localStorage.removeItem(SESSION_KEY)
}

// 初始化
loadToken()

export function useAuth() {
  return {
    token,
    isLoggedIn,
    setToken,
    clearToken,
  }
}
