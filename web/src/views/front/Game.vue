<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchGames } from '@/api'
import type { GameItem } from '@/types'

const all = ref<GameItem[]>([])
const activeCat = ref('')
const loading = ref(false)

const categories = computed(() => {
  const set = new Set<string>()
  for (const g of all.value) if (g.category) set.add(g.category)
  return [...set]
})

const filtered = computed(() =>
  activeCat.value ? all.value.filter((g) => g.category === activeCat.value) : all.value,
)

/** 🔥 热门游戏榜：按热度取前 8 */
const hotList = computed(() =>
  [...all.value].sort((a, b) => (b.hot || 0) - (a.hot || 0)).slice(0, 8),
)

function formatHot(n: number): string {
  if (!n) return '0'
  return n >= 10000 ? (n / 10000).toFixed(1) + ' 万' : String(n)
}

const parseTags = (tags: string): string[] =>
  tags.split(/[,，]/).map((t) => t.trim()).filter(Boolean)

onMounted(async () => {
  loading.value = true
  try {
    all.value = (await fetchGames()) || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container game-wrap">
    <div class="game-hero card">
      <h1 class="hero-title">🎮 游戏世界</h1>
      <p class="hero-sub">在玩的游戏与电竞项目，按分类浏览。</p>
      <div v-if="categories.length" class="cat-tabs">
        <button class="cat-tab" :class="{ active: activeCat === '' }" @click="activeCat = ''">全部</button>
        <button
          v-for="cat in categories"
          :key="cat"
          class="cat-tab"
          :class="{ active: activeCat === cat }"
          @click="activeCat = activeCat === cat ? '' : cat"
        >
          {{ cat }}
        </button>
      </div>
    </div>

    <div class="game-layout">
      <!-- 左：游戏卡片 -->
      <div class="game-main">
        <div class="game-grid" v-loading="loading">
          <router-link v-for="g in filtered" :key="g.id" :to="`/game/${g.id}`" class="game-card card">
            <div class="cover" :style="g.cover ? { background: `url(${g.cover}) center/cover` } : {}">
              <span v-if="!g.cover" class="cover-fallback">{{ g.title.slice(0, 1) }}</span>
              <span v-if="g.platform" class="platform-tag">{{ g.platform }}</span>
            </div>
            <div class="info">
              <h3 class="title">{{ g.title }}</h3>
              <p class="meta">
                <span v-if="g.category" class="cat-badge">{{ g.category }}</span>
                <span class="hot">🔥 {{ formatHot(g.hot) }}</span>
              </p>
              <p v-if="g.description" class="desc">{{ g.description }}</p>
              <div class="tags">
                <span v-for="t in parseTags(g.tags)" :key="t" class="tag">{{ t }}</span>
              </div>
            </div>
          </router-link>
        </div>
        <el-empty v-if="!loading && !filtered.length" description="这个分类下还没有游戏" />
      </div>

      <!-- 右：热门榜 -->
      <aside v-if="hotList.length" class="hot-panel card">
        <header class="hot-head">
          <h3 class="hot-title">🔥 热门榜</h3>
        </header>
        <ol class="hot-list">
          <li v-for="(g, i) in hotList" :key="g.id" class="hot-item">
            <span class="hot-rank" :class="{ top3: i < 3 }">{{ i + 1 }}</span>
            <router-link :to="`/game/${g.id}`" class="hot-name" :title="g.title">{{ g.title }}</router-link>
            <span v-if="g.platform" class="hot-plat">{{ g.platform.split(',')[0] }}</span>
            <span class="hot-count">{{ formatHot(g.hot) }}</span>
          </li>
        </ol>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.game-wrap {
  max-width: 1200px;
}

.game-hero {
  padding: 32px 40px;
  margin-bottom: 22px;
  background:
    radial-gradient(circle at 92% -30%, var(--accent-soft), transparent 50%),
    var(--card);
}

.hero-title {
  font-family: var(--font-heading);
  font-size: 30px;
  margin: 0 0 8px;
}

.hero-sub {
  color: var(--muted);
  font-size: 14.5px;
  margin: 0 0 16px;
}

.cat-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.cat-tab {
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  padding: 5px 16px;
  border-radius: 999px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.cat-tab.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
  font-weight: 600;
}

/* ===== 两栏布局 ===== */
.game-layout {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 20px;
  align-items: start;
}

.game-main {
  min-width: 0;
}

.game-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  min-height: 200px;
}

.game-card {
  overflow: hidden;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.game-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lift);
}

.cover {
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: linear-gradient(150deg, var(--surface-alt), var(--surface));
  display: grid;
  place-items: center;
}

.cover-fallback {
  font-family: var(--font-heading);
  font-size: 44px;
  font-weight: 700;
  color: var(--muted);
  opacity: 0.55;
}

.platform-tag {
  position: absolute;
  right: 6px;
  bottom: 6px;
  font-size: 11px;
  color: #fff;
  background: rgb(0 0 0 / 0.55);
  border-radius: 4px;
  padding: 1px 7px;
}

.info {
  padding: 14px 16px 16px;
}

.title {
  font-family: var(--font-heading);
  font-size: 17px;
  font-weight: 700;
  color: var(--heading);
  margin: 0 0 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: var(--muted);
  margin: 0 0 8px;
  flex-wrap: wrap;
}

.cat-badge {
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 999px;
  padding: 1px 9px;
}

.hot {
  font-size: 12px;
}

.desc {
  font-size: 13px;
  color: var(--text);
  line-height: 1.6;
  margin: 0 0 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tag {
  font-size: 11px;
  color: var(--muted);
  border: 1px solid var(--border);
  padding: 0 7px;
  border-radius: 999px;
}

/* ===== 热门榜 ===== */
.hot-panel {
  position: sticky;
  top: 86px;
  padding: 16px 18px;
}

.hot-head {
  margin-bottom: 10px;
}

.hot-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--heading);
  margin: 0;
}

.hot-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.hot-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 4px;
  border-bottom: 1px dashed var(--border);
}

.hot-item:last-child {
  border-bottom: none;
}

.hot-rank {
  width: 18px;
  flex-shrink: 0;
  text-align: center;
  font-family: var(--font-mono);
  font-size: 14px;
  font-weight: 800;
  color: var(--muted);
  font-style: italic;
}

.hot-rank.top3 {
  color: var(--accent);
}

.hot-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hot-plat {
  font-size: 11px;
  color: var(--muted);
  flex-shrink: 0;
}

.hot-count {
  font-size: 12px;
  color: var(--muted);
  font-family: var(--font-mono);
  flex-shrink: 0;
}

@media (max-width: 1000px) {
  .game-layout {
    grid-template-columns: 1fr;
  }

  .hot-panel {
    position: static;
    order: 2;
  }
}

@media (max-width: 700px) {
  .game-grid {
    grid-template-columns: 1fr;
  }
}
</style>
