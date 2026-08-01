<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { login } from '../api/auth'
import { ApiError } from '../api/client'

interface LoginError {
  title: string
  detail: string
  status: string
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
        detail: 'Your previous session is no longer valid. Sign in again to continue.',
        status: 'HTTP 401 / SESSION_INVALID',
      }
    : null,
)
const errorBox = ref<HTMLElement | null>(null)
const loading = ref(false)

function describeLoginError(cause: unknown): LoginError {
  if (cause instanceof ApiError) {
    if (cause.status === 401) {
      return {
        title: 'AUTHENTICATION FAILED',
        detail: 'The username or password is incorrect. Check spelling and capitalization, then try again.',
        status: 'HTTP 401 / INVALID_CREDENTIALS',
      }
    }
    if (cause.status === 400) {
      return {
        title: 'REQUEST REJECTED',
        detail: cause.message || 'The server could not read the submitted credentials.',
        status: `HTTP 400${cause.code ? ` / ${cause.code}` : ''}`,
      }
    }
    if (cause.status === 429) {
      return {
        title: 'TOO MANY ATTEMPTS',
        detail: 'The server is temporarily limiting login attempts. Wait a moment before retrying.',
        status: `HTTP 429${cause.code ? ` / ${cause.code}` : ''}`,
      }
    }
    if (cause.status >= 500) {
      return {
        title: 'AUTH SERVICE UNAVAILABLE',
        detail: 'The server could not complete authentication. Please try again later.',
        status: `HTTP ${cause.status}${cause.code ? ` / ${cause.code}` : ''}`,
      }
    }
    return {
      title: 'LOGIN REQUEST FAILED',
      detail: cause.message,
      status: `HTTP ${cause.status}${cause.code ? ` / ${cause.code}` : ''}`,
    }
  }

  return {
    title: 'CONNECTION FAILED',
    detail: 'The authentication server could not be reached. Check your connection and try again.',
    status: 'NETWORK_ERROR',
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
      detail: 'Enter both a username and password before submitting.',
      status: 'INPUT_REQUIRED',
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
        <h2>> AUTH_REQUIRED</h2>
        <div class="cursor-line">
          <span>mode: login | </span><span>awaiting credentials...</span><span class="cursor"></span>
        </div>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-row">
          <label for="login-username">USERNAME:</label>
          <input
            id="login-username"
            v-model="username"
            type="text"
            placeholder="root"
            autocomplete="username"
            :aria-invalid="!!error"
            aria-describedby="login-error"
          />
        </div>
        <div class="form-row">
          <label for="login-password">PASSWORD:</label>
          <input
            id="login-password"
            v-model="password"
            type="password"
            placeholder="********"
            autocomplete="current-password"
            :aria-invalid="!!error"
            aria-describedby="login-error"
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
          <div class="error-title"><span>[FAIL]</span> {{ error.title }}</div>
          <div class="error-status">{{ error.status }}</div>
          <p>{{ error.detail }}</p>
        </div>

        <div class="form-actions">
          <button type="submit" class="btn" :disabled="loading">
            {{ loading ? '[ AUTHENTICATING... ]' : '[ LOGIN ]' }}
          </button>
        </div>
      </form>

      <div class="login-footer">
        <RouterLink to="/" class="btn btn-sm">cd ~</RouterLink>
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

.login-header h2 {
  font-size: 1.4rem;
  margin-bottom: 0.5rem;
}

.cursor-line {
  font-size: 0.85rem;
  color: var(--muted);
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
  color: #ff4444;
  font-size: 0.85rem;
  margin-bottom: 1rem;
  padding: 10px 12px;
  border: 1px solid #ff4444;
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
