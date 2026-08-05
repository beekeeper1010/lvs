<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="onLogin">
      <div class="logo">LVS</div>
      <p class="subtitle">本地视频服务</p>
      <div class="field">
        <input v-model.trim="username" placeholder="用户名" autocomplete="username" />
      </div>
      <div class="field">
        <input
          v-model="password"
          type="password"
          placeholder="密码"
          autocomplete="current-password"
        />
      </div>
      <p v-if="error" class="error">{{ error }}</p>
      <button class="submit" type="submit" :disabled="loading">
        {{ loading ? '登录中…' : '登 录' }}
      </button>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')

async function onLogin() {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const data = await login(username.value, password.value)
    localStorage.setItem('token', data.token)
    router.push('/gallery')
  } catch (e) {
    error.value = e.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f1115 0%, #1a2030 100%);
}
.login-card {
  width: 360px;
  padding: 40px 36px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
}
.logo {
  font-size: 44px;
  font-weight: 800;
  text-align: center;
  background: linear-gradient(90deg, #6366f1, #22d3ee);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.subtitle {
  text-align: center;
  color: #8b93a7;
  margin: 8px 0 28px;
  font-size: 14px;
}
.field {
  margin-bottom: 16px;
}
.field input {
  width: 100%;
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.06);
  color: #e6e8eb;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}
.field input:focus {
  border-color: #6366f1;
}
.error {
  color: #f87171;
  font-size: 13px;
  margin-bottom: 12px;
  text-align: center;
}
.submit {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(90deg, #6366f1, #22d3ee);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}
.submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
