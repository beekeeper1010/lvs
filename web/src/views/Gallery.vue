<template>
  <div class="gallery-page">
    <header class="topbar">
      <h2>视频广场</h2>
      <button class="logout-btn" @click="onLogout">注销</button>
    </header>

    <div v-if="loading" class="status">加载中…</div>
    <div v-else-if="videos.length === 0" class="status">暂无视频，请先运行 lvs scan</div>
    <div v-else class="grid">
      <div v-for="v in videos" :key="v.id" class="card" @click="play(v)">
        <div class="thumb">
          <img v-if="v.thumb_path" :src="thumbUrl(v)" loading="lazy" alt="" />
          <div v-else class="no-thumb">无预览</div>
          <div class="play-icon">▶</div>
        </div>
        <div class="name" :title="v.name">{{ v.name }}</div>
      </div>
    </div>

    <footer v-if="total > 0" class="pager">
      <button :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
      <span>{{ page }} / {{ totalPages }}</span>
      <button :disabled="page >= totalPages" @click="changePage(page + 1)">下一页</button>
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
  sessionStorage.setItem('lvs_video_name', v.name)
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
}
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  position: sticky;
  top: 0;
  background: rgba(15, 17, 21, 0.9);
  backdrop-filter: blur(8px);
  z-index: 10;
}
.topbar h2 {
  font-size: 18px;
}
.logout-btn {
  padding: 8px 18px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: transparent;
  color: #e6e8eb;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}
.logout-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}
.status {
  text-align: center;
  padding: 100px 0;
  color: #8b93a7;
}
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
  padding: 24px;
}
.card {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}
.card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
}
.thumb {
  position: relative;
  aspect-ratio: 16 / 9;
  background: #000;
}
.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.no-thumb {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #555d6e;
  font-size: 13px;
}
.play-icon {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 38px;
  color: #fff;
  opacity: 0;
  background: rgba(0, 0, 0, 0.25);
  transition: opacity 0.2s;
}
.card:hover .play-icon {
  opacity: 1;
}
.name {
  padding: 12px 14px;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 24px;
}
.pager button {
  padding: 8px 18px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: transparent;
  color: #e6e8eb;
  font-size: 14px;
  cursor: pointer;
}
.pager button:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
}
.pager button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.pager span {
  color: #8b93a7;
  font-size: 14px;
}
</style>
