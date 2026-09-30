<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { fetchAnimes } from '@/api'
import type { AnimeItem } from '@/types'

const all = ref<AnimeItem[]>([])
const activeCat = ref('')
const loading = ref(false)

/** 分类标签（从数据中提取去重） */
const categories = computed(() => {
  const set = new Set<string>()
  for (const a of all.value) if (a.category) set.add(a.category)
  return [...set]
})

const filtered = computed(() =>
  activeCat.value ? all.value.filter((a) => a.category === activeCat.value) : all.value,
)

/** 🔥 热播榜：总榜 + 每个分类一个榜；按播放量取前 8，「换一换」轮换起点 */
const hotTab = ref('') // '' = 总榜，其他 = 分类名
const hotOffset = ref(0)
const hotList = computed(() => {
  const base = hotTab.value ? all.value.filter((a) => a.category === hotTab.value) : all.value
  const sorted = base.slice().sort((a, b) => (b.playCount || 0) - (a.playCount || 0))
  const top = sorted.filter((a) => (a.playCount || 0) > 0)
  if (top.length <= 8) return top
  const start = hotOffset.value % top.length
  return [...top.slice(start), ...top.slice(0, start)].slice(0, 8)
})

watch(hotTab, () => (hotOffset.value = 0))

function shuffleHot() {
  hotOffset.value += 3
}

function formatCount(n: number): string {
  if (!n) return '0'
  return n >= 10000 ? (n / 10000).toFixed(1) + ' 万' : String(n)
}

/** 徽章：播放量前 3 = 热，7 天内新增 = 新 */
function hotBadge(a: AnimeItem, idx: number): '热' | '新' | '' {
  if (idx < 3) return '热'
  if (Date.now() - new Date(a.createdAt).getTime() < 7 * 24 * 3600 * 1000) return '新'
  return ''
}

/** 排名徽章：前三名特殊色 */
function rankClass(rank: number): string {
  if (rank === 1) return 'gold'
  if (rank === 2) return 'silver'
  if (rank === 3) return 'bronze'
  return ''
}

onMounted(async () => {
  loading.value = true
  try {
    all.value = (await fetchAnimes()) || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container anime-wrap">
    <div class="anime-hero card">
      <h1 class="hero-title">动漫排行榜</h1>
      <p class="hero-sub">我在追的和觉得值得安利的动漫，按喜爱程度排序。</p>
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

    <div class="anime-layout">
      <!-- 左：排行卡片 -->
      <div class="anime-main">
        <div class="anime-grid" v-loading="loading">
          <div v-for="(a, i) in filtered" :key="a.id" class="anime-card card">
            <div class="rank-badge" :class="rankClass(a.rank || i + 1)">
              {{ a.rank || i + 1 }}
            </div>
            <div class="cover" :style="a.cover ? { background: `url(${a.cover}) center/cover` } : {}">
              <span v-if="!a.cover" class="cover-fallback">{{ a.title.slice(0, 1) }}</span>
              <span v-if="a.episodes" class="episodes-tag">{{ a.episodes }}</span>
            </div>
            <div class="info">
              <h3 class="title">{{ a.title }}</h3>
              <p class="meta">
                <span v-if="a.category" class="cat-badge">{{ a.category }}</span>
                <span v-if="a.region">{{ a.region }}</span>
                <span class="meta-play">▶ {{ formatCount(a.playCount) }}</span>
              </p>
              <p v-if="a.description" class="desc">{{ a.description }}</p>
            </div>
          </div>
        </div>
        <el-empty v-if="!loading && !filtered.length" description="这个分类下还没有动漫" />
      </div>

      <!-- 右：热播榜 -->
      <aside v-if="hotList.length" class="hot-panel card">
        <header class="hot-head">
          <h3 class="hot-title">🔥 热播榜</h3>
          <button class="hot-shuffle" title="换一批" @click="hotOffset += 3">
            <el-icon><Refresh /></el-icon> 换一换
          </button>
        </header>

        <div class="hot-tabs">
          <button class="hot-tab" :class="{ active: hotTab === '' }" @click="hotTab = ''">总榜</button>
          <button
            v-for="cat in categories"
            :key="cat"
            class="hot-tab"
            :class="{ active: hotTab === cat }"
            @click="hotTab = cat"
          >
            {{ cat }}
          </button>
        </div>

        <ol v-if="hotList.length" class="hot-list">
          <li v-for="(a, i) in hotList" :key="a.id" class="hot-item">
            <span class="hot-rank" :class="{ 'top3': i < 3 }">{{ i + 1 }}</span>
            <router-link :to="`/anime`" class="hot-title" :title="a.title">{{ a.title }}</router-link>
            <span v-if="hotBadge(a, i)" class="hot-badge" :class="hotBadge(a, i) === '热' ? 'hot' : 'new'">
              {{ hotBadge(a, i) }}
            </span>
            <span class="hot-count">{{ formatCount(a.playCount) }}</span>
          </li>
        </ol>
        <p v-else class="hot-empty">这个分类下还没有播放数据</p>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.anime-wrap {
  max-width: 1200px;
}

.anime-hero {
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

/* ===== 两栏：卡片 + 热播榜 ===== */
.anime-layout {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 20px;
  align-items: start;
}

.anime-main {
  min-width: 0;
}

.anime-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 18px;
  min-height: 200px;
}

.anime-card {
  position: relative;
  overflow: hidden;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.anime-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lift);
}

.rank-badge {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 2;
  min-width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  font-family: var(--font-mono);
  font-size: 15px;
  font-weight: 800;
  color: #fff;
  border-radius: 0 0 10px 0;
  background: var(--accent);
  box-shadow: 2px 2px 6px rgb(0 0 0 / 0.15);
}

.rank-badge.gold {
  background: linear-gradient(135deg, #ffb02e, #ff8c00);
}

.rank-badge.silver {
  background: linear-gradient(135deg, #a8b4c4, #8494a8);
}

.rank-badge.bronze {
  background: linear-gradient(135deg, #d29a6b, #b0784a);
}

.cover {
  position: relative;
  width: 100%;
  aspect-ratio: 3 / 4;
  background: linear-gradient(150deg, var(--surface-alt), var(--surface));
  display: grid;
  place-items: center;
}

.cover-fallback {
  font-family: var(--font-heading);
  font-size: 52px;
  font-weight: 700;
  color: var(--muted);
  opacity: 0.55;
}

.episodes-tag {
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

.meta-play {
  font-size: 12px;
}

.desc {
  font-size: 13px;
  color: var(--text);
  line-height: 1.6;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ===== 热播榜 ===== */
.hot-panel {
  position: sticky;
  top: 86px;
  padding: 16px 18px;
}

.hot-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.hot-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--heading);
  margin: 0;
}

.hot-shuffle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: none;
  background: transparent;
  color: var(--muted);
  font-size: 12px;
  cursor: pointer;
}

.hot-shuffle:hover {
  color: var(--accent);
}

.hot-tabs {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.hot-tab {
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  font-size: 12px;
  padding: 3px 10px;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.15s;
}

.hot-tab.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
  font-weight: 600;
}

.hot-empty {
  font-size: 12.5px;
  color: var(--muted);
  text-align: center;
  padding: 16px 0;
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

.hot-title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hot-item:hover .hot-title {
  color: var(--accent);
}

.hot-badge {
  flex-shrink: 0;
  font-size: 10px;
  border-radius: 3px;
  padding: 0 4px;
  line-height: 1.5;
}

.hot-badge.hot {
  color: #fff;
  background: var(--accent);
}

.hot-badge.new {
  color: var(--accent);
  background: var(--accent-soft);
}

.hot-count {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--muted);
  font-family: var(--font-mono);
}

@media (max-width: 1000px) {
  .anime-layout {
    grid-template-columns: 1fr;
  }

  .hot-panel {
    position: static;
    order: 2;
  }
}

@media (max-width: 760px) {
  .anime-grid {
    grid-template-columns: 1fr;
  }
}
</style>
