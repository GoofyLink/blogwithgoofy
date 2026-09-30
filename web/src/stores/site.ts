import { defineStore } from 'pinia'
import { fetchSettings } from '@/api'

/** 站点个性化配置（后台「个人信息」设置，前台首页/页脚使用） */
export const useSiteStore = defineStore('site', {
  state: () => ({
    settings: {} as Record<string, string>,
    loaded: false,
  }),
  getters: {
    get: (state) => (key: string, fallback = '') => state.settings[key] || fallback,
  },
  actions: {
    async load(force = false) {
      if (this.loaded && !force) return
      try {
        this.settings = (await fetchSettings()) || {}
        this.loaded = true
      } catch {
        /* 拉取失败时使用组件内默认值 */
      }
    },
  },
})
