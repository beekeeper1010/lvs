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
      <button class="logout-btn" @click="onLogout">
        <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
          <path d="M16 17l5-5-5-5" />
          <path d="M21 12H9" />
        </svg>
        注销
      </button>
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { fetchVideos, logout } from '../api'

const router = useRouter()
const videos = ref([])
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

function changePage(p) {
  page.value = p
  load()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function onLogout() {
  try {
    await logout()
  } catch (e) {
    /* ignore */
  }
  localStorage.removeItem('token')
  router.push('/login')
}

onMounted(load)
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
.logout-btn {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-2);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.logout-btn:hover {
  color: var(--text);
  background: rgba(248, 113, 113, 0.12);
  border-color: rgba(248, 113, 113, 0.4);
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
</style>
