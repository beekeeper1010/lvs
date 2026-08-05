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
          <el-avatar :size="28" class="avatar">{{ userStore.avatarText }}</el-avatar>
          <span class="nickname">{{ userStore.displayName }}</span>
        </div>
        <el-dropdown trigger="click" @command="onCommand">
          <el-button class="menu-btn" size="default">
            操作
            <el-icon class="arrow"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="userStore.isAdmin" command="users" :icon="UserFilled">
                用户管理
              </el-dropdown-item>
              <el-dropdown-item command="settings" :icon="Setting">用户设置</el-dropdown-item>
              <el-dropdown-item command="logout" divided :icon="SwitchButton">注销</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <section class="hero">
      <h1>视频广场</h1>
      <p>共 <b>{{ total }}</b> 部影片 · 发现你的本地收藏</p>
    </section>

    <div v-loading="loading" class="grid" element-loading-text="加载中…">
      <el-empty
        v-if="!loading && videos.length === 0"
        description="暂无视频，请先运行 lvs scan 扫描视频"
        class="empty-box"
      />
      <div
        v-for="(v, i) in videos"
        :key="v.id"
        class="card"
        :style="{ animationDelay: `${i * 45}ms` }"
        @click="play(v)"
      >
        <div class="thumb">
          <el-image v-if="v.thumb_path" :src="thumbUrl(v)" fit="cover" lazy />
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
          <el-button
            v-if="userStore.isAdmin"
            class="del-btn"
            size="small"
            circle
            type="danger"
            plain
            :icon="Delete"
            @click.stop="onDelete(v)"
          />
        </div>
        <div class="meta">
          <div class="name" :title="v.name">{{ v.name }}</div>
          <div class="date">{{ v.created_at }}</div>
        </div>
      </div>
    </div>

    <footer v-if="total > 0" class="pager">
      <el-pagination
        background
        layout="prev, pager, next, total"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="changePage"
      />
    </footer>

    <!-- 用户设置弹窗 -->
    <el-dialog v-model="showSettings" title="用户设置" width="420px" align-center>
      <el-form label-position="top">
        <el-form-item label="昵称">
          <el-input v-model.trim="settingsForm.nickname" placeholder="显示昵称" />
        </el-form-item>
        <el-form-item label="当前密码">
          <el-input
            v-model="settingsForm.oldPassword"
            type="password"
            show-password
            placeholder="修改密码时填写"
          />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input
            v-model="settingsForm.newPassword"
            type="password"
            show-password
            placeholder="留空则不修改"
          />
        </el-form-item>
        <el-form-item label="确认新密码">
          <el-input
            v-model="settingsForm.confirmPassword"
            type="password"
            show-password
            placeholder="再次输入新密码"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSettings = false">取消</el-button>
        <el-button type="primary" :loading="savingSettings" @click="saveSettings">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, UserFilled, Setting, SwitchButton, Delete } from '@element-plus/icons-vue'
import { fetchVideos, fetchUserInfo, updateProfile, deleteVideo, logout } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const videos = ref([])
const page = ref(1)
const pageSize = 12
const total = ref(0)
const loading = ref(false)

// 用户设置弹窗
const showSettings = ref(false)
const savingSettings = ref(false)
const settingsForm = ref({ nickname: '', oldPassword: '', newPassword: '', confirmPassword: '' })

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
    // 拦截器已处理
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

function onCommand(cmd) {
  if (cmd === 'users') router.push('/users')
  else if (cmd === 'settings') openSettings()
  else if (cmd === 'logout') onLogout()
}

function openSettings() {
  settingsForm.value = {
    nickname: userStore.nickname || userStore.username,
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
  showSettings.value = true
}

async function saveSettings() {
  if (!settingsForm.value.nickname) {
    ElMessage.warning('昵称不能为空')
    return
  }
  if (settingsForm.value.newPassword && !settingsForm.value.oldPassword) {
    ElMessage.warning('修改密码需填写当前密码')
    return
  }
  if (settingsForm.value.newPassword !== settingsForm.value.confirmPassword) {
    ElMessage.warning('两次输入的新密码不一致')
    return
  }
  const changedPwd = !!settingsForm.value.newPassword
  savingSettings.value = true
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
    ElMessage.success('设置已保存')
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    savingSettings.value = false
  }
}

function changePage(p) {
  page.value = p
  load()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

// 删除视频（仅移除库记录与缩略图，不删除源文件）
async function onDelete(v) {
  try {
    await ElMessageBox.confirm(`确定删除视频 "${v.name}" 吗？\n仅移除视频库记录，不会删除源文件。`, '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch (e) {
    return // 取消
  }
  try {
    await deleteVideo(v.id)
    ElMessage.success('视频已删除')
    if (videos.value.length === 1 && page.value > 1) page.value -= 1
    load()
  } catch (e) {
    ElMessage.error(e.message || '删除失败')
  }
}

async function onLogout() {
  try {
    await ElMessageBox.confirm('确定要注销当前账号吗？', '注销确认', {
      confirmButtonText: '注销',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch (e) {
    return // 用户取消
  }
  try {
    await logout()
  } catch (e) {
    /* ignore */
  }
  userStore.logout()
  router.push('/login')
}

onMounted(() => {
  load()
  syncUserInfo()
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
  background: var(--grad);
  font-weight: 700;
}
.nickname {
  max-width: 120px;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.menu-btn {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text);
}
.menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.3);
}
.arrow {
  margin-left: 2px;
  transition: transform 0.2s;
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

.empty-box {
  grid-column: 1 / -1;
  padding: 60px 0;
}

/* 卡片网格 */
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 22px;
  padding: 28px;
  max-width: 1440px;
  margin: 0 auto;
  min-height: 200px;
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
.thumb :deep(.el-image) {
  width: 100%;
  height: 100%;
  display: block;
  transition: transform 0.45s ease;
}
.thumb :deep(.el-image img) {
  transition: transform 0.45s ease;
}
.card:hover .thumb :deep(.el-image img) {
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
.del-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 3;
  opacity: 0;
  transition: opacity 0.2s;
}
.card:hover .del-btn {
  opacity: 1;
}
.shade {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 55%, rgba(0, 0, 0, 0.55));
  opacity: 0;
  transition: opacity 0.25s;
  pointer-events: none;
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
  pointer-events: none;
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
  padding: 10px 0 40px;
}
</style>
