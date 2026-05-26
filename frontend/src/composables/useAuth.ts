import { ref, computed } from 'vue'

const token = ref<string>('')
const isLoggedIn = computed(() => !!token.value)

function loadToken() {
  // Token lives only in memory. Backend sets HttpOnly cookie.
  // We check sessionStorage for a lightweight "was logged in" flag
  // so the UI can show login state after page refresh.
  if (sessionStorage.getItem('blog_logged_in') === '1') {
    // Temp placeholder — actual validation happens on first API call.
    token.value = 'restored'
  }
}

function setToken(newToken: string) {
  token.value = newToken
  sessionStorage.setItem('blog_logged_in', '1')
}

function clearToken() {
  token.value = ''
  sessionStorage.removeItem('blog_logged_in')
}

// initialize
loadToken()

export function useAuth() {
  return {
    token,
    isLoggedIn,
    setToken,
    clearToken,
  }
}
