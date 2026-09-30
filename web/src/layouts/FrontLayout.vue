<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { motion, AnimatePresence, useScroll, useSpring } from 'motion-v'
import { loadOml2d } from 'oh-my-live2d'
import { useAuthStore } from '@/stores/auth'
import { useSiteStore } from '@/stores/site'
import ThemeSwitcher from '@/components/ThemeSwitcher.vue'
import StarField from '@/components/StarField.vue'
import PetRope from '@/components/PetRope.vue'

const site = useSiteStore()
site.load()

/* 滚动进度条：scrollYProgress 经弹簧平滑后驱动 scaleX */
const { scrollYProgress } = useScroll()
const scrollScaleX = useSpring(scrollYProgress, {
  stiffness: 100,
  damping: 30,
  restDelta: 0.001,
})

/* ===== 页脚：看板娘 + 站点运行时间 ===== */
const SITE_BIRTH = new Date('2026-09-28T00:00:00') // 建站日期，可改
const runtime = ref({ days: 0, hours: '00', minutes: '00', seconds: '00' })
let runtimeTimer: ReturnType<typeof setInterval> | null = null

function tickRuntime() {
  const diff = Date.now() - SITE_BIRTH.getTime()
  const totalSec = Math.floor(diff / 1000)
  const days = Math.floor(totalSec / 86400)
  const hours = String(Math.floor((totalSec % 86400) / 3600)).padStart(2, '0')
  const minutes = String(Math.floor((totalSec % 3600) / 60)).padStart(2, '0')
  const seconds = String(totalSec % 60).padStart(2, '0')
  runtime.value = { days, hours, minutes, seconds }
}

// oh-my-live2d 会把“休息”状态单独持久化到 OML2D_STATUS。两种隐藏来源
// 必须合并判断，否则模型实际已经滑出页面，页脚却仍显示“隐藏看板娘”。
const live2dHidden = ref(
  localStorage.getItem('live2dHidden') === '1' ||
  localStorage.getItem('OML2D_STATUS') === 'sleep',
)

function loadLive2d() {
  if (live2dHidden.value) return
  loadOml2d({
    models: [{ path: 'https://model.hacxy.cn/HK416-2-normal/model.json', scale: 0.08, position: [0, 60] }],
    menus: {
      // 只保留第一个默认菜单项（休息），其余隐藏
      items: (defaultItems) => defaultItems.slice(0, 1),
    },
    statusBar: { disable: true },
  })
}

/** 看板娘隐藏/唤回：唤回走整页重建，避免热重挂载失败 */
function toggleLive2d() {
  if (live2dHidden.value) {
    // 唤回：清除标记并整页刷新，保证初始化路径与首次访问完全一致
    localStorage.removeItem('live2dHidden')
    localStorage.removeItem('OML2D_STATUS')
    location.reload()
  } else {
    live2dHidden.value = true
    localStorage.setItem('live2dHidden', '1')
    // 隐藏：立即从页面移除看板娘节点
    document.querySelector('#oml2d-stage')?.remove()
  }
}

onMounted(() => {
  tickRuntime()
  runtimeTimer = setInterval(tickRuntime, 1000)
  loadLive2d()
})

// 看板娘节点由 oh-my-live2d 直接挂到 document.body，不属于当前布局的
// Vue 子树。离开前台时手动清理，否则切到后台后节点仍会留在页面上。
onUnmounted(() => {
  if (runtimeTimer) clearInterval(runtimeTimer)
  document.querySelector('#oml2d-stage')?.remove()
  document.querySelector('#oml2d-global-style')?.remove()
})

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const keyword = ref('')

function doSearch() {
  const kw = keyword.value.trim()
  if (kw) router.push({ path: '/', query: { keyword: kw } })
}

const navs = [
  { path: '/', label: '首页' },
  { path: '/essays', label: '随笔' },
  { path: '/learn', label: '学习', extra: 'EN · DE' },
  { path: '/novel', label: '小说' },
  { path: '/anime', label: '动漫' },
  { path: '/art', label: '艺术' },
  { path: '/ai', label: 'AI' },
  { path: '/game', label: '游戏' },
  { path: '/archives', label: '归档' },
  { path: '/links', label: '友链' },
  { path: '/about', label: '关于' },
]

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <!-- 全局装饰：左星星 / 右萌宠挂绳（所有前台页面常驻，仅宽屏显示） -->
  <StarField />
  <PetRope />

  <div class="site">
    <header class="site-header">
      <div class="container header-inner">
        <router-link to="/" class="logo">
          <span class="logo-mark">天</span>
          <span class="logo-text">小天 <em>Goofy · Blog</em></span>
        </router-link>

        <nav class="nav">
          <router-link
            v-for="n in navs"
            :key="n.path"
            :to="n.path"
            class="nav-item"
            :class="{ active: isActive(n.path) }"
          >
            <!-- 共享布局动画：激活背景块在导航项之间平滑滑动 -->
            <motion.span
              v-if="isActive(n.path)"
              layout-id="nav-indicator"
              class="nav-indicator"
              :transition="{ type: 'spring', stiffness: 380, damping: 32 }"
            />
            <span class="nav-text">
              {{ n.label }}<sup v-if="n.extra" class="nav-sup">{{ n.extra }}</sup>
            </span>
          </router-link>
        </nav>

        <div class="header-actions">
          <el-input
            v-model="keyword"
            placeholder="搜索文章…"
            class="search-input"
            :prefix-icon="Search"
            size="small"
            @keyup.enter="doSearch"
          />
          <ThemeSwitcher />
          <router-link v-if="auth.isLoggedIn" to="/admin/dashboard" class="admin-entry">
            管理后台
          </router-link>
        </div>
      </div>

      <!-- 滚动进度条：scrollYProgress → scaleX，随页面滚动拉满 -->
      <motion.div class="scroll-progress" :style="{ scaleX: scrollScaleX }" />
    </header>

    <main class="site-main">
      <router-view v-slot="{ Component, route }">
        <!-- AnimatePresence：跨栏目切换时旧页上移淡出、新页从下淡入；
             key 为一级栏目，栏目内切换（如下一章）不触发动画 -->
        <AnimatePresence mode="wait">
          <motion.div
            :key="route.path.split('/')[1] || 'home'"
            :initial="{ opacity: 0, y: 16 }"
            :animate="{ opacity: 1, y: 0 }"
            :exit="{ opacity: 0, y: -10 }"
            :transition="{ duration: 0.22, ease: 'easeOut' }"
          >
            <component :is="Component" />
          </motion.div>
        </AnimatePresence>
      </router-view>
    </main>

    <footer class="site-footer">
      <div class="container footer-inner">
        <div class="footer-left">
          <div class="runtime">
            <span class="runtime-label">本站已稳定运行</span>
            <span class="runtime-nums">
              <b class="rt-num">{{ runtime.days }}</b>天
              <b class="rt-num">{{ runtime.hours }}</b>时
              <b class="rt-num">{{ runtime.minutes }}</b>分
              <b class="rt-num rt-sec">{{ runtime.seconds }}</b>秒
            </span>
          </div>
          <span class="footer-copy">© {{ new Date().getFullYear() }} {{ site.get('nickname', '小天') }} Goofy · 用 Vue3 + Go 认真写字</span>
        </div>
        <nav class="footer-links">
          <router-link to="/about">关于</router-link>
          <a :href="site.get('github', 'https://github.com/')" target="_blank" rel="noopener">GitHub</a>
          <a v-if="site.get('email')" :href="`mailto:${site.get('email')}`">邮箱</a>
          <router-link to="/links">友链</router-link>
          <button class="live2d-toggle" title="显示/隐藏看板娘" @click="toggleLive2d">
            👧 {{ live2dHidden ? '唤回看板娘' : '隐藏看板娘' }}
          </button>
        </nav>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.site {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 毛玻璃导航栏：吸顶 + 半透明背景 + 背景模糊（玻璃拟态） */
.site-header {
  position: sticky;
  top: 0;
  z-index: 100;

  /* 不支持背景模糊时的实色兜底 */
  background: var(--card);
  border-bottom: 1px solid rgb(0 0 0 / 6%);
}

/* 滚动进度条：贴在导航栏底边 */
.scroll-progress {
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 3px;
  background: var(--accent);
  transform-origin: 0 50%;
  z-index: 1;
}

@supports (
  (backdrop-filter: blur(16px)) or
  (-webkit-backdrop-filter: blur(16px))
) {
  .site-header {
    background: var(--header-bg); /* 各主题 78%~80% 透明度 */

    -webkit-backdrop-filter: blur(16px) saturate(140%);
    backdrop-filter: blur(16px) saturate(140%);
  }
}

.header-inner {
  display: flex;
  align-items: center;
  gap: 28px;
  height: 62px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.logo-mark {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  background: var(--accent);
  color: #fff;
  font-family: var(--font-heading);
  font-size: 19px;
  border-radius: 9px;
  box-shadow: var(--shadow);
}

.logo-text {
  font-family: var(--font-heading);
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 1px;
}

.logo-text em {
  font-style: normal;
  font-size: 12px;
  color: var(--muted);
  margin-left: 6px;
  letter-spacing: 0;
}

.nav {
  display: flex;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.nav-item {
  position: relative;
  padding: 6px 9px;
  border-radius: 8px;
  font-size: 14.5px;
  color: var(--muted);
  white-space: nowrap;
  transition: color 0.2s;
}

.nav-text {
  position: relative;
  z-index: 1;
}

/* 共享布局动画的激活背景块（跟随主题色） */
.nav-indicator {
  position: absolute;
  inset: 0;
  border-radius: 8px;
  background: color-mix(in srgb, var(--accent) 9%, transparent);
}

.nav-item:hover {
  color: var(--text);
  background: color-mix(in srgb, var(--accent) 6%, transparent);
}

.nav-item.active {
  color: var(--accent);
  font-weight: 600;
}

.nav-sup {
  font-size: 10px;
  color: var(--accent);
  margin-left: 3px;
  transform: translateY(-4px);
  display: inline-block;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-shrink: 0;
}

.search-input {
  width: 150px;
}

.admin-entry {
  font-size: 13px;
  color: var(--muted);
  border: 1px solid var(--border);
  padding: 6px 14px;
  border-radius: 999px;
  white-space: nowrap;
  transition: all 0.2s;
}

.admin-entry:hover {
  color: var(--accent);
  border-color: var(--accent);
}

.site-main {
  flex: 1;
  padding: 34px 0 60px;
}

.site-footer {
  border-top: 1px solid var(--border);
  padding: 22px 0;
}

.footer-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--muted);
  font-size: 13px;
  gap: 14px;
  flex-wrap: wrap;
}

.footer-links {
  display: flex;
  gap: 18px;
}

.footer-links a {
  color: var(--muted);
}

.footer-links a:hover {
  color: var(--accent);
}

/* ===== 运行时间计数器 ===== */
.footer-left {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.runtime {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.runtime-label {
  font-size: 12px;
  color: var(--muted);
}

.runtime-nums {
  display: inline-flex;
  align-items: baseline;
  gap: 3px;
  font-size: 12px;
  color: var(--muted);
}

.rt-num {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 700;
  color: var(--accent);
  min-width: 1.4em;
  text-align: center;
  display: inline-block;
}

.footer-copy {
  font-size: 13px;
  color: var(--muted);
}

/* ===== 看板娘找回按钮 ===== */
.live2d-toggle {
  border: none;
  background: transparent;
  font-size: 13px;
  color: var(--muted);
  cursor: pointer;
  padding: 0;
}

.live2d-toggle:hover {
  color: var(--accent);
}

@media (max-width: 1240px) {
  .nav-item {
    padding: 6px 7px;
    font-size: 14px;
  }

  .search-input {
    width: 130px;
  }
}

@media (max-width: 900px) {
  .nav-sup,
  .search-input {
    display: none;
  }

  .header-inner {
    gap: 10px;
  }
}
</style>
