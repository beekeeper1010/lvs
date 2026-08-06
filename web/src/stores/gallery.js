import { defineStore } from 'pinia'

// 广场页 UI 状态（组件销毁重建后仍保留，供从播放页返回时恢复）
export const useGalleryStore = defineStore('gallery', {
  state: () => ({
    page: 1,
  }),
})
