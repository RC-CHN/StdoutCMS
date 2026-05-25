<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const router = useRouter()
const { setToken } = useAuth()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  error.value = ''
  if (!username.value || !password.value) {
    error.value = 'ERROR: username and password required'
    return
  }

  loading.value = true

  // TODO: 对接后端 API
  // const res = await fetch('/api/v1/admin/login', {
  //   method: 'POST',
  //   headers: { 'Content-Type': 'application/json' },
  //   body: JSON.stringify({ username: username.value, password: password.value }),
  // })
  // if (!res.ok) { error.value = 'ERROR: authentication failed'; loading.value = false; return }
  // const data = await res.json()
  // setToken(data.token)

  // mock: 任意用户名密码都过
  setToken('mock-session-token-' + Date.now())
  loading.value = false
  router.push('/admin')
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
          <label>USERNAME:</label>
          <input v-model="username" type="text" placeholder="root" autocomplete="username" />
        </div>
        <div class="form-row">
          <label>PASSWORD:</label>
          <input v-model="password" type="password" placeholder="********" autocomplete="current-password" />
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

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
  font-weight: bold;
  font-size: 0.85rem;
  margin-bottom: 1rem;
  padding: 6px 10px;
  border: 1px solid #ff4444;
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
