<template>
  <div class="gallery-page">
    <header class="topbar">
      <div class="brand">
        <div class="brand-badge">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
            <path d="M8 5.14v13.72c0 .81.89 1.3 1.57.87l10.6-6.86a1.04 1.04 0 0 0 0-1.74L9.57 4.27A1.03 1.03 0 0 0 8 5.14z" />
          </svg>
        </div>
        <span class="brand-name">LVS</span>
      </div>
      <div class="actions">
        <div class="user-chip" :title="userStore.username">
          <span class="avatar">{{ userStore.avatarText }}</span>
          <span class="nickname">{{ userStore.displayName }}</span>
        </div>
        <div class="dropdown">
          <button class="menu-btn" @click.stop="menuOpen = !menuOpen">
            操作
            <svg
              viewBox="0 0 24 24"
              width="14"
              height="14"
              fill="none"
              stroke="currentColor"
              stroke-width="2.2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="arrow"
              :class="{ rotated: menuOpen }"
            >
              <path d="M6 9l6 6 6-6" />
            </svg>
          </button>
          <transition name="dropdown">
            <div v-if="menuOpen" class="menu" @click.stop>
              <button v-if="userStore.isAdmin" class="menu-item" @click="goUsers">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
                  <circle cx="9" cy="7" r="4" />
                  <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
                  <path d="M16 3.13a4 4 0 0 1 0 7.75" />
                </svg>
                用户管理
              </button>
              <button class="menu-item" @click="openSettings">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="3" />
                  <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
                </svg>
                用户设置
              </button>
              <button class="menu-item danger" @click="onLogout">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                  <path d="M16 17l5-5-5-5" />
                  <path d="M21 12H9" />
                </svg>
                注销
              </button>
            </div>
          </transition>
        </div>
      </div>
    </header>

    <section class="hero">
      <h1>视频广场</h1>
      <p>共 <b>{{ total }}</b> 部影片 · 发现你的本地收藏</p>
    </section>

    <div v-if="loading" class="status"><div class="spinner"></div></div>

    <div v-else-if="videos.length === 0" class="status empty">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" width="34" height="34" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <rect x="2" y="2" width="20" height="20" rx="3" />
          <circle cx="12" cy="12" r="4" />
          <line x1="12" y1="2" x2="12" y2="6" />
        </svg>
      </div>
      <p>暂无视频</p>
      <span>请先运行 <code>lvs scan --dir &lt;目录&gt;</code> 扫描视频</span>
    </div>

    <div v-else class="grid">
      <div
        v-for="(v, i) in videos"
        :key="v.id"
        class="card"
        :style="{ animationDelay: `${i * 45}ms` }"
        @click="play(v)"
      >
        <div class="thumb">
          <img v-if="v.thumb_path" :src="thumbUrl(v)" loading="lazy" alt="" />
          <div v-else class="no-thumb">
            <svg viewBox="0 0 24 24" width="26" height="26" fill="currentColor">
              <path d="M8 5.14v13.72c0 .81.89 1.3 1.57.87l10.6-6.86a1.04 1.04 0 0 0 0-1.74L9.57 4.27A1.03 1.03 0 0 0 8 5.14z" />
            </svg>
          </div>
          <div class="shade"></div>
          <div class="play-btn">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
              <path d="M8 5.14v13.72c0 .81.89 1.3 1.57.87l10.6-6.86a1.04 1.04 0 0 0 0-1.74L9.57 4.27A1.03 1.03 0 0 0 8 5.14z" />
            </svg>
          </div>
        </div>
        <div class="meta">
          <div class="name" :title="v.name">{{ v.name }}</div>
          <div class="date">{{ v.created_at }}</div>
        </div>
      </div>
    </div>

    <footer v-if="total > 0" class="pager">
      <button class="pg-btn" :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
      <span class="pg-info">{{ page }} / {{ totalPages }}</span>
      <button class="pg-btn" :disabled="page >= totalPages" @click="changePage(page + 1)">下一页</button>
    </footer>

    <!-- 用户设置弹窗 -->
    <div v-if="showSettings" class="modal-mask" @click.self="showSettings = false">
      <div class="modal">
        <h3>用户设置</h3>
        <div class="field">
          <label>昵称</label>
          <input v-model.trim="settingsForm.nickname" placeholder="显示昵称" />
        </div>
        <div class="field">
          <label>当前密码</label>
          <input v-model="settingsForm.oldPassword" type="password" placeholder="修改密码时填写" />
        </div>
        <div class="field">
          <label>新密码</label>
          <input v-model="settingsForm.newPassword" type="password" placeholder="留空则不修改" />
        </div>
        <div class="field">
          <label>确认新密码</label>
          <input v-model="settingsForm.confirmPassword" type="password" placeholder="再次输入新密码" />
        </div>
        <p v-if="settingsError" class="error">{{ settingsError }}</p>
        <div class="modal-actions">
          <button class="btn cancel" @click="showSettings = false">取消</button>
          <button class="btn ok" :disabled="savingSettings" @click="saveSettings">
            {{ savingSettings ? '保存中…' : '保存' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { fetchVideos, fetchUserInfo, updateProfile, logout } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const videos = ref([])
const menuOpen = ref(false)

// 用户设置弹窗
const showSettings = ref(false)
const savingSettings = ref(false)
const settingsError = ref('')
const settingsForm = ref({ nickname: '', oldPassword: '', newPassword: '', confirmPassword: '' })
const page = ref(1)
const pageSize = 12
const total = ref(0)
const loading = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

function thumbUrl(v) {
  const token = localStorage.getItem('token')
  return `/api/video/thumb?id=${v.id}&token=${token}`
}

function play(v) {
  sessionStorage.setItem('lvs_video', JSON.stringify(v))
  router.push(`/play/${v.id}`)
}

async function load() {
  loading.value = true
  try {
    const data = await fetchVideos(page.value, pageSize)
    videos.value = data.list
    total.value = data.total
  } catch (e) {
    // 拦截器已处理跳转
  } finally {
    loading.value = false
  }
}

// 从服务端同步最新用户信息（昵称可能已被修改）
async function syncUserInfo() {
  try {
    userStore.setInfo(await fetchUserInfo())
  } catch (e) {
    // 忽略，沿用本地缓存
  }
}

function changePage(p) {
  page.value = p
  load()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function goUsers() {
  menuOpen.value = false
  router.push('/users')
}

function openSettings() {
  menuOpen.value = false
  settingsForm.value = {
    nickname: userStore.nickname || userStore.username,
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  settingsError.value = ''
  showSettings.value = true
}

async function saveSettings() {
  if (!settingsForm.value.nickname) {
    settingsError.value = '昵称不能为空'
    return
  }
  if (settingsForm.value.newPassword && !settingsForm.value.oldPassword) {
    settingsError.value = '修改密码需填写当前密码'
    return
  }
  if (settingsForm.value.newPassword !== settingsForm.value.confirmPassword) {
    settingsError.value = '两次输入的新密码不一致'
    return
  }
  const changedPwd = !!settingsForm.value.newPassword
  savingSettings.value = true
  settingsError.value = ''
  try {
    await updateProfile({
      nickname: settingsForm.value.nickname,
      password: settingsForm.value.newPassword,
      old_password: settingsForm.value.oldPassword,
    })
    showSettings.value = false
    if (changedPwd) {
      // 密码已修改，旧 token 已失效，强制重新登录
      userStore.logout()
      router.push({ path: '/login', query: { msg: '密码已修改，请重新登录' } })
      return
    }
    syncUserInfo()
  } catch (e) {
    settingsError.value = e.message || '保存失败'
  } finally {
    savingSettings.value = false
  }
}

async function onLogout() {
  menuOpen.value = false
  try {
    await logout()
  } catch (e) {
    /* ignore */
  }
  userStore.logout()
  router.push('/login')
}

// 点击页面空白区域收起下拉菜单
function onDocClick() {
  menuOpen.value = false
}

onMounted(() => {
  load()
  syncUserInfo()
  document.addEventListener('click', onDocClick)
})

onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
})
</script>

<style scoped>
.gallery-page {
  min-height: 100vh;
  animation: fadeUp 0.4s ease both;
}

/* 顶栏 */
.topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 28px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(11, 12, 17, 0.72);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
}
.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}
.brand-badge {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 9px;
  background: var(--grad);
  color: #fff;
  box-shadow: 0 6px 18px rgba(124, 92, 255, 0.4);
}
.brand-name {
  font-size: 17px;
  font-weight: 800;
  letter-spacing: 4px;
  background: var(--grad);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.actions {
  display: flex;
  align-items: center;
  gap: 14px;
}
.dropdown {
  position: relative;
  z-index: 32;
}
.menu-btn {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}
.menu-btn .arrow {
  transition: transform 0.2s;
}
.menu-btn .arrow.rotated {
  transform: rotate(180deg);
}
.menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 150px;
  padding: 6px;
  border-radius: 12px;
  background: #1a1c26;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 16px 44px rgba(0, 0, 0, 0.55);
  z-index: 31;
}
.menu-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s;
}
.menu-item:hover {
  background: rgba(124, 92, 255, 0.16);
  color: #c4b5fd;
}
.menu-item.danger:hover {
  background: rgba(248, 113, 113, 0.14);
  color: #fca5a5;
}
.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px 4px 4px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.05);
}
.avatar {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: var(--grad);
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}
.nickname {
  max-width: 120px;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 页首 */
.hero {
  padding: 44px 28px 10px;
  text-align: center;
}
.hero h1 {
  font-size: 34px;
  font-weight: 800;
  letter-spacing: 2px;
  background: linear-gradient(120deg, #fff 20%, #a78bfa 60%, #22d3ee 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.hero p {
  margin-top: 10px;
  font-size: 14px;
  color: var(--text-2);
}
.hero p b {
  color: var(--accent-2);
  font-weight: 700;
}

/* 状态 */
.status {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 90px 0;
  color: var(--text-2);
}
.empty-icon {
  width: 76px;
  height: 76px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px dashed rgba(255, 255, 255, 0.18);
  color: var(--text-3);
}
.empty p {
  font-size: 16px;
  font-weight: 600;
}
.empty span {
  font-size: 13px;
  color: var(--text-3);
}
.empty code {
  padding: 2px 7px;
  border-radius: 6px;
  background: rgba(124, 92, 255, 0.14);
  color: #b4a5ff;
  font-size: 12px;
}

/* 卡片网格 */
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 22px;
  padding: 28px;
  max-width: 1440px;
  margin: 0 auto;
}
.card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  opacity: 0;
  animation: fadeUp 0.45s ease forwards;
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease,
    background 0.25s ease;
}
.card:hover {
  transform: translateY(-6px);
  background: var(--surface-hover);
  border-color: rgba(124, 92, 255, 0.4);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(124, 92, 255, 0.18);
}

.thumb {
  position: relative;
  aspect-ratio: 16 / 9;
  background: linear-gradient(135deg, #151724, #1c1e2b);
  overflow: hidden;
}
.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.45s ease;
}
.card:hover .thumb img {
  transform: scale(1.07);
}
.no-thumb {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.28);
}
.shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 55%, rgba(0, 0, 0, 0.55));
  opacity: 0;
  transition: opacity 0.25s;
}
.card:hover .shade {
  opacity: 1;
}
.play-btn {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: scale(0.8);
  transition: opacity 0.25s, transform 0.25s;
}
.play-btn svg {
  width: 52px;
  height: 52px;
  padding: 14px;
  border-radius: 50%;
  background: var(--grad);
  color: #fff;
  box-shadow: 0 10px 30px rgba(124, 92, 255, 0.5);
}
.card:hover .play-btn {
  opacity: 1;
  transform: scale(1);
}

.meta {
  padding: 13px 15px 14px;
}
.name {
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.date {
  margin-top: 5px;
  font-size: 12px;
  color: var(--text-3);
}

/* 分页 */
.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 18px;
  padding: 10px 0 40px;
}
.pg-btn {
  padding: 9px 20px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.pg-btn:hover:not(:disabled) {
  background: rgba(124, 92, 255, 0.18);
  border-color: rgba(124, 92, 255, 0.5);
}
.pg-btn:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.pg-info {
  font-size: 13px;
  color: var(--text-2);
  font-variant-numeric: tabular-nums;
}

/* 设置弹窗 */
.modal-mask {
  position: fixed;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(6px);
}
.modal {
  width: 400px;
  max-width: calc(100vw - 40px);
  padding: 30px 32px;
  border-radius: 18px;
  background: #1a1c26;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 30px 90px rgba(0, 0, 0, 0.6);
  animation: fadeUp 0.25s ease both;
}
.modal h3 {
  margin-bottom: 22px;
  font-size: 18px;
  font-weight: 700;
}
.field {
  margin-bottom: 16px;
}
.field label {
  display: block;
  margin-bottom: 7px;
  font-size: 13px;
  color: var(--text-2);
}
.field input {
  width: 100%;
  padding: 11px 14px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: var(--text);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.field input:focus {
  border-color: rgba(124, 92, 255, 0.7);
  box-shadow: 0 0 0 4px rgba(124, 92, 255, 0.16);
}
.error {
  color: var(--danger);
  font-size: 13px;
  margin: 2px 0 12px;
}
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
}
.btn {
  padding: 10px 22px;
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}
.btn.cancel {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-2);
  border-color: rgba(255, 255, 255, 0.14);
}
.btn.cancel:hover {
  color: var(--text);
  background: rgba(255, 255, 255, 0.1);
}
.btn.ok {
  background: var(--grad);
  color: #fff;
  font-weight: 600;
}
.btn.ok:hover:not(:disabled) {
  filter: brightness(1.1);
}
.btn.ok:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
