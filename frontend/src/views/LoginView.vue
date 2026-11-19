<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { login } from '../api/auth'
import { ApiError } from '../api/client'

interface LoginError {
  title: string
  detail: string
  status?: string
}

const router = useRouter()
const route = useRoute()
const { setToken } = useAuth()

const username = ref('')
const password = ref('')
const error = ref<LoginError | null>(
  route.query.reason === 'session_expired'
    ? {
        title: 'SESSION EXPIRED',
        detail: 'Sign in again to continue.',
      }
    : null,
)
const errorBox = ref<HTMLElement | null>(null)
const loading = ref(false)

function describeLoginError(cause: unknown): LoginError {
  if (cause instanceof ApiError) {
    const status = cause.code ? `HTTP ${cause.status} / ${cause.code}` : `HTTP ${cause.status}`
    if (cause.status === 401) {
      return {
        title: 'SIGN IN FAILED',
        detail: 'Check your username and password, then try again.',
        status,
      }
    }
    if (cause.status === 400) {
      return {
        title: 'REQUEST REJECTED',
        detail: 'Enter your username and password, then try again.',
        status,
      }
    }
    if (cause.status === 429) {
      return {
        title: 'TOO MANY ATTEMPTS',
        detail: 'Wait a moment before signing in again.',
        status,
      }
    }
    if (cause.status >= 500) {
      return {
        title: 'SIGN IN UNAVAILABLE',
        detail: 'The server could not sign you in. Try again later.',
        status,
      }
    }
    return {
      title: 'SIGN IN FAILED',
      detail: cause.message,
      status,
    }
  }

  return {
    title: 'CONNECTION FAILED',
    detail: 'Could not reach the server. Check your connection and try again.',
  }
}

async function showError(value: LoginError) {
  error.value = value
  await nextTick()
  errorBox.value?.focus()
}

async function handleLogin() {
  error.value = null
  if (!username.value || !password.value) {
    await showError({
      title: 'MISSING CREDENTIALS',
      detail: 'Enter your username and password.',
    })
    return
  }

  loading.value = true
  try {
    const data = await login(username.value, password.value)
    setToken(data.token)
    await router.push('/admin')
  } catch (cause: unknown) {
    await showError(describeLoginError(cause))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrapper">
    <div class="login-box">
      <div class="login-header">
        <h1>ADMIN SIGN IN</h1>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-row">
          <label for="login-username">USERNAME</label>
          <input
            id="login-username"
            v-model="username"
            type="text"
            autocomplete="username"
            :aria-invalid="!!error"
            :aria-describedby="error ? 'login-error' : undefined"
          />
        </div>
        <div class="form-row">
          <label for="login-password">PASSWORD</label>
          <input
            id="login-password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            :aria-invalid="!!error"
            :aria-describedby="error ? 'login-error' : undefined"
          />
        </div>

        <div
          v-if="error"
          id="login-error"
          ref="errorBox"
          class="error-msg"
          role="alert"
          aria-live="assertive"
          tabindex="-1"
        >
          <div class="error-title"><span>[ ERROR ]</span> {{ error.title }}</div>
          <div v-if="error.status" class="error-status">{{ error.status }}</div>
          <p>{{ error.detail }}</p>
        </div>

        <div class="form-actions">
          <button type="submit" class="btn" :disabled="loading">
            {{ loading ? 'SIGNING IN…' : 'SIGN IN' }}
          </button>
        </div>
      </form>

      <div class="login-footer">
        <RouterLink to="/" class="btn btn-sm">← ARTICLES</RouterLink>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-wrapper {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 4rem;
  min-height: 60vh;
}

.login-box {
  border: 2px solid var(--border);
  padding: 2rem;
  width: 100%;
  max-width: 450px;
  box-shadow: var(--shadow);
  background: var(--bg);
}

.login-header {
  border-bottom: 2px solid var(--border);
  padding-bottom: 1rem;
  margin-bottom: 1.5rem;
}

.login-header h1 {
  font-size: 1.4rem;
}

.login-form {
  margin-bottom: 1.5rem;
}

.form-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 1rem;
}

.form-row label {
  font-weight: bold;
  font-size: 0.85rem;
  min-width: 80px;
  color: var(--muted);
}

.form-row input {
  flex: 1;
  background: var(--bg);
  color: var(--fg);
  border: 1px solid var(--border);
  padding: 8px 10px;
  font-family: var(--font-main);
  font-size: 0.9rem;
  outline: none;
}

.form-row input:focus {
  border-width: 2px;
}

.error-msg {
  color: var(--danger);
  font-size: 0.85rem;
  margin-bottom: 1rem;
  padding: 10px 12px;
  border: 1px solid var(--danger);
  outline: none;
}

.error-title {
  font-weight: bold;
}

.error-title span {
  display: inline-block;
  margin-right: 4px;
}

.error-status {
  margin: 3px 0 5px;
  color: var(--muted);
  font-size: 0.75rem;
  font-weight: bold;
}

.error-msg p {
  color: var(--fg);
  line-height: 1.45;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.login-footer {
  border-top: 2px dashed var(--border);
  padding-top: 1rem;
}

.btn-sm {
  padding: 4px 12px;
  font-size: 0.8rem;
}

@media (max-width: 600px) {
  .login-wrapper {
    padding-top: 2rem;
  }
  .login-box {
    padding: 1.5rem;
  }
  .form-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .form-row label {
    min-width: auto;
  }
  .form-row input {
    width: 100%;
  }
}
</style>
