<template>
  <div class="player-page">
    <!-- 影院氛围层 -->
    <div class="glow"></div>
    <div class="beam"></div>
    <div class="lights"></div>

    <header class="topbar">
      <h2 class="title" :class="{ hidden: !showTitle }" :title="videoName">
        {{ videoName || '播放中' }}
      </h2>
      <el-button class="back-btn" @click="goBack">← 返回广场</el-button>
    </header>

    <div class="video-wrap" @mouseenter="onMouseEnter" @mouseleave="onMouseLeave">
      <button class="nav-btn nav-prev" :disabled="!adjacent.prev" :title="adjacent.prev ? '上一个' : ''" @click="goPrev">
        <el-icon :size="24"><ArrowLeft /></el-icon>
      </button>
      <div class="video-frame">
        <video :key="route.params.id" :src="playUrl" controls autoplay playsinline @play="onVideoPlay"></video>
      </div>
      <button class="nav-btn nav-next" :disabled="!adjacent.next" :title="adjacent.next ? '下一个' : ''" @click="goNext">
        <el-icon :size="24"><ArrowRight /></el-icon>
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { fetchAdjacent } from '../api'

const route = useRoute()
const router = useRouter()

const video = ref(null)
try {
  video.value = JSON.parse(sessionStorage.getItem('lvs_video') || 'null')
} catch (e) {
  video.value = null
}

const videoName = computed(() => video.value?.name || '')
const adjacent = ref({ prev: null, next: null })
const showTitle = ref(false)
let hideTimer = null
let hovering = false
let forcedUntil = 0 // 播放后强制显示标题的截止时间戳

// video 标签无法携带 Authorization 头, token 通过 query 传递
const playUrl = computed(
  () => `/api/video/play?id=${route.params.id}&token=${localStorage.getItem('token')}`
)

// 加载上一个/下一个视频
async function loadAdjacent() {
  try {
    adjacent.value = await fetchAdjacent(route.params.id)
  } catch (e) {
    adjacent.value = { prev: null, next: null }
  }
}

function switchTo(v) {
  if (!v) return
  sessionStorage.setItem('lvs_video', JSON.stringify(v))
  video.value = v
  router.replace(`/play/${v.id}`)
}

function goPrev() {
  switchTo(adjacent.value.prev)
}

function goNext() {
  switchTo(adjacent.value.next)
}

// 开始播放时显示标题，3 秒后自动隐藏（强制显示期内不因鼠标离开而隐藏）
function onVideoPlay() {
  clearTimeout(hideTimer)
  showTitle.value = true
  forcedUntil = Date.now() + 3000
  hideTimer = setTimeout(() => {
    if (!hovering) showTitle.value = false
  }, 3000)
}

// 鼠标进入播放器时显示标题
function onMouseEnter() {
  hovering = true
  clearTimeout(hideTimer)
  showTitle.value = true
}

function onMouseLeave() {
  hovering = false
  clearTimeout(hideTimer)
  if (Date.now() < forcedUntil) {
    // 仍处于播放后强制显示期：到期后再隐藏
    hideTimer = setTimeout(() => {
      showTitle.value = false
    }, forcedUntil - Date.now())
  } else {
    showTitle.value = false
  }
}

function goBack() {
  router.push('/gallery')
}

watch(() => route.params.id, loadAdjacent, { immediate: true })

onUnmounted(() => {
  clearTimeout(hideTimer)
})
</script>

<style scoped>
.player-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(1100px 560px at 50% -8%, rgba(124, 92, 255, 0.22), transparent 62%),
    radial-gradient(900px 460px at 50% 118%, rgba(34, 211, 238, 0.13), transparent 62%),
    radial-gradient(700px 380px at 18% 45%, rgba(255, 255, 255, 0.035), transparent 70%),
    linear-gradient(180deg, #0a0b12 0%, #0e0f19 48%, #090a10 100%);
  animation: fadeUp 0.35s ease both;
}

/* 屏幕上方光晕 */
.glow {
  position: fixed;
  left: 50%;
  top: 0;
  width: 72vw;
  height: 46vh;
  transform: translateX(-50%);
  background: radial-gradient(ellipse 55% 100% at 50% 0%, rgba(124, 92, 255, 0.28), rgba(34, 211, 238, 0.08) 45%, transparent 72%);
  pointer-events: none;
  z-index: 0;
}

/* 放映机光束 */
.beam {
  position: fixed;
  top: 0;
  left: 50%;
  width: 980px;
  height: 100%;
  transform: translateX(-50%);
  background: linear-gradient(180deg, rgba(170, 160, 255, 0.12) 0%, rgba(170, 160, 255, 0.03) 48%, transparent 78%);
  clip-path: polygon(41% 0, 59% 0, 100% 100%, 0 100%);
  filter: blur(7px);
  pointer-events: none;
  z-index: 0;
}

/* 底部座位区光带 */
.lights {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 340px;
  background: repeating-linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.05) 0 1px,
    transparent 1px 46px
  );
  -webkit-mask-image: linear-gradient(180deg, transparent, #000 55%);
  mask-image: linear-gradient(180deg, transparent, #000 55%);
  pointer-events: none;
  z-index: 0;
}

.topbar {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 14px 28px;
}
.back-btn {
  border-color: rgba(255, 255, 255, 0.18);
  background: rgba(0, 0, 0, 0.35);
  color: #fff;
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
}
.back-btn:hover {
  background: rgba(124, 92, 255, 0.35);
  border-color: rgba(124, 92, 255, 0.6);
  color: #fff;
}
.title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  max-width: 55%;
  font-size: 16px;
  font-weight: 600;
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.6);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  opacity: 1;
  transition: opacity 0.25s ease;
}
.title.hidden {
  opacity: 0;
}

.video-wrap {
  position: relative;
  z-index: 1;
  max-width: 1080px;
  margin: 36px auto 60px;
  padding: 0 20px;
}
.video-frame {
  position: relative;
  border-radius: 8px;
  padding: 2px;
  background: linear-gradient(140deg, rgba(124, 92, 255, 0.6), rgba(34, 211, 238, 0.4), rgba(255, 255, 255, 0.1));
  box-shadow: 0 40px 110px rgba(0, 0, 0, 0.78), 0 8px 40px rgba(0, 0, 0, 0.5);
}
/* 屏幕表面光泽 */
.video-frame::after {
  content: '';
  position: absolute;
  inset: 2px;
  border-radius: 6px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.07), transparent 26%);
  pointer-events: none;
  z-index: 1;
}
video {
  position: relative;
  display: block;
  width: 100%;
  border-radius: 6px;
  background: #000;
  max-height: 82vh;
  outline: none;
}

/* 上一集/下一集切换按钮 */
.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 52px;
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(0, 0, 0, 0.4);
  color: #fff;
  cursor: pointer;
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: all 0.2s;
  z-index: 5;
}
.nav-prev {
  left: -118px;
}
.nav-next {
  right: -118px;
}
.nav-btn:hover:not(:disabled) {
  background: rgba(124, 92, 255, 0.4);
  border-color: rgba(124, 92, 255, 0.6);
}
.nav-btn:disabled {
  opacity: 0.25;
  cursor: not-allowed;
}
</style>
