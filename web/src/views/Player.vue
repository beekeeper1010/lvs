<template>
  <div class="player-page">
    <header class="topbar">
      <button class="back-btn" @click="goBack">← 返回</button>
      <h2 class="title" :title="videoName">{{ videoName || '播放中' }}</h2>
    </header>
    <div class="video-wrap">
      <video :src="playUrl" controls autoplay playsinline></video>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const videoName = ref(sessionStorage.getItem('lvs_video_name') || '')

// video 标签无法携带 Authorization 头, token 通过 query 传递
const playUrl = computed(
  () => `/api/video/play?id=${route.params.id}&token=${localStorage.getItem('token')}`
)

function goBack() {
  router.push('/gallery')
}
</script>

<style scoped>
.player-page {
  min-height: 100vh;
}
.topbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(15, 17, 21, 0.9);
  backdrop-filter: blur(8px);
  position: sticky;
  top: 0;
  z-index: 10;
}
.back-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: transparent;
  color: #e6e8eb;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.2s;
}
.back-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}
.title {
  font-size: 16px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.video-wrap {
  max-width: 1080px;
  margin: 24px auto;
  padding: 0 16px;
}
video {
  width: 100%;
  border-radius: 12px;
  background: #000;
  max-height: 82vh;
  outline: none;
}
</style>
