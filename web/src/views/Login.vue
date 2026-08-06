<template>
  <div class="login-page">
    <div class="orb orb-1"></div>
    <div class="orb orb-2"></div>
    <div class="grid-overlay"></div>

    <el-form class="login-card" @submit.prevent="onLogin">
      <div class="brand">
        <div class="brand-icon">
          <PlayIcon :size="24" />
        </div>
        <div class="brand-name">LVS</div>
        <p class="brand-sub">LOCAL VIDEO SERVICE</p>
      </div>

      <el-form-item>
        <el-input
          v-model.trim="username"
          placeholder="用户名"
          size="large"
          :prefix-icon="User"
          clearable
        />
      </el-form-item>
      <el-form-item>
        <el-input
          v-model="password"
          type="password"
          placeholder="密码"
          size="large"
          :prefix-icon="Lock"
          show-password
          @keyup.enter="onLogin"
        />
      </el-form-item>

      <el-alert v-if="notice" :title="notice" type="success" :closable="false" class="tip" />
      <el-alert v-if="error" :title="error" type="error" :closable="false" class="tip" />

      <el-button class="submit" type="primary" size="large" native-type="submit" :loading="loading">
        登 录
      </el-button>
    </el-form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '../api'
import { useUserStore } from '../stores/user'
import PlayIcon from '../components/PlayIcon.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const username = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const notice = ref(route.query.msg || '')

async function onLogin() {
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  error.value = ''
  try {
    const data = await login(username.value, password.value)
    userStore.setLogin(data)
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
  position: relative;
  overflow: hidden;
  padding: 24px;
}

/* 漂浮光斑 */
.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(110px);
  opacity: 0.55;
  pointer-events: none;
}
.orb-1 {
  width: 480px;
  height: 480px;
  top: -12%;
  right: -6%;
  background: radial-gradient(circle, rgba(124, 92, 255, 0.8), transparent 70%);
  animation: orbFloat 14s ease-in-out infinite alternate;
}
.orb-2 {
  width: 420px;
  height: 420px;
  bottom: -16%;
  left: -8%;
  background: radial-gradient(circle, rgba(34, 211, 238, 0.7), transparent 70%);
  animation: orbFloat 17s ease-in-out infinite alternate-reverse;
}
@keyframes orbFloat {
  from {
    transform: translate(0, 0) scale(1);
  }
  to {
    transform: translate(40px, -30px) scale(1.12);
  }
}

/* 网格底纹 */
.grid-overlay {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 44px 44px;
  -webkit-mask-image: radial-gradient(ellipse 60% 55% at 50% 45%, #000 30%, transparent 75%);
  mask-image: radial-gradient(ellipse 60% 55% at 50% 45%, #000 30%, transparent 75%);
}

.login-card {
  position: relative;
  z-index: 1;
  width: 380px;
  max-width: 100%;
  padding: 44px 40px 36px;
  background: rgba(19, 20, 28, 0.72);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid rgba(255, 255, 255, 0.09);
  border-radius: 22px;
  box-shadow: 0 30px 90px rgba(0, 0, 0, 0.55), inset 0 1px 0 rgba(255, 255, 255, 0.06);
  animation: fadeUp 0.5s ease both;
}
.login-card :deep(.el-form-item) {
  margin-bottom: 18px;
}
.login-card :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06);
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.1) inset;
}
.login-card :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset, 0 0 0 4px rgba(124, 92, 255, 0.16);
}

.brand {
  text-align: center;
  margin-bottom: 30px;
}
.brand-icon {
  width: 58px;
  height: 58px;
  margin: 0 auto 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 18px;
  background: var(--grad);
  color: #fff;
  box-shadow: 0 14px 36px rgba(124, 92, 255, 0.45);
}
.brand-name {
  font-size: 34px;
  font-weight: 800;
  letter-spacing: 10px;
  text-indent: 10px;
  background: var(--grad);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.brand-sub {
  margin-top: 8px;
  font-size: 10px;
  letter-spacing: 3.5px;
  color: var(--text-3);
}

.tip {
  margin-bottom: 16px;
  border-radius: 8px;
}

.submit {
  width: 100%;
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 4px;
  text-indent: 4px;
  border: none;
  box-shadow: 0 10px 28px rgba(124, 92, 255, 0.35);
  transition: transform 0.15s, box-shadow 0.2s;
}
.submit:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 36px rgba(124, 92, 255, 0.5);
}
</style>
