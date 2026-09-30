<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import { fetchGame, fetchGamePosts } from '@/api'
import type { GameItem, GamePostItem } from '@/types'

const route = useRoute()
const game = ref<GameItem | null>(null)
const posts = ref<GamePostItem[]>([])
const activeType = ref('')
const loading = ref(false)

const types = computed(() => {
  const set = new Set<string>()
  for (const p of posts.value) if (p.type) set.add(p.type)
  return [...set]
})

const filtered = computed(() =>
  activeType.value ? posts.value.filter((p) => p.type === activeType.value) : posts.value,
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
    const id = route.params.id as string
    const [g, p] = await Promise.all([fetchGame(id), fetchGamePosts(id)])
    game.value = g
    posts.value = p || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container detail-wrap" v-if="game">
    <router-link to="/game" class="back-link">← 游戏世界</router-link>

    <!-- 游戏信息头 -->
    <div class="card head-card">
      <div class="cover" :style="game.cover ? { background: `url(${game.cover}) center/cover` } : {}">
        <span v-if="!game.cover">{{ game.title.slice(0, 1) }}</span>
      </div>
      <div class="head-info">
        <h1 class="head-title">{{ game.title }}</h1>
        <p class="meta">
          <span v-if="game.category" class="cat-badge">{{ game.category }}</span>
          <span v-if="game.platform">{{ game.platform }}</span>
          <span class="hot">🔥 {{ formatHot(game.hot) }}</span>
        </p>
        <div class="tags" v-if="parseTags(game.tags).length">
          <span v-for="t in parseTags(game.tags)" :key="t" class="tag">{{ t }}</span>
        </div>
      </div>
    </div>

    <!-- 攻略资讯圈 -->
    <div class="card posts-card">
      <div class="posts-head">
        <h3 class="posts-title">📕 攻略资讯圈</h3>
        <div v-if="types.length" class="type-tabs">
          <button class="type-tab" :class="{ active: !activeType }" @click="activeType = ''">全部</button>
          <button
            v-for="t in types"
            :key="t"
            class="type-tab"
            :class="{ active: activeType === t }"
            @click="activeType = activeType === t ? '' : t"
          >{{ t }}</button>
        </div>
      </div>

      <div v-loading="loading">
        <div v-if="filtered.length" class="post-list">
          <router-link
            v-for="p in filtered"
            :key="p.id"
            :to="`/game/post/${p.id}`"
            class="post-row"
          >
            <div class="post-main">
              <span class="post-title">
                <em v-if="p.type" class="type-badge">{{ p.type }}</em>
                {{ p.title }}
              </span>
              <span v-if="p.summary" class="post-summary">{{ p.summary }}</span>
            </div>
            <div class="post-side">
              <span class="post-views">{{ p.views }} 浏览</span>
              <span class="post-date">{{ dayjs(p.createdAt).format('MM-DD') }}</span>
            </div>
          </router-link>
        </div>
        <el-empty v-else-if="!loading" description="还没有攻略资讯，去后台「游戏攻略」发布吧" />
      </div>
    </div>

    <!-- 游戏详细介绍 -->
    <div v-if="game.content" class="card content-card md-content" v-html="game.content" />
  </div>
</template>

<style scoped>
.detail-wrap {
  max-width: 900px;
}

.back-link {
  font-size: 13px;
  color: var(--muted);
  display: inline-block;
  margin-bottom: 12px;
}

.back-link:hover {
  color: var(--accent);
}

.head-card {
  display: flex;
  gap: 22px;
  padding: 22px 26px;
  margin-bottom: 18px;
}

.cover {
  width: 200px;
  height: 120px;
  flex-shrink: 0;
  border-radius: 10px;
  background: linear-gradient(150deg, var(--surface-alt), var(--surface));
  display: grid;
  place-items: center;
  font-family: var(--font-heading);
  font-size: 40px;
  font-weight: 700;
  color: var(--muted);
}

.head-info {
  flex: 1;
  min-width: 0;
}

.head-title {
  font-family: var(--font-heading);
  font-size: 26px;
  margin: 0 0 8px;
}

.meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--muted);
  margin: 0 0 10px;
  flex-wrap: wrap;
}

.cat-badge {
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 999px;
  padding: 1px 10px;
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
  padding: 0 8px;
  border-radius: 999px;
}

/* ===== 帖子圈 ===== */
.posts-card {
  padding: 20px 26px;
  margin-bottom: 18px;
}

.posts-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.posts-title {
  font-family: var(--font-heading);
  font-size: 17px;
  margin: 0;
}

.type-tabs {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.type-tab {
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  font-size: 12.5px;
  padding: 3px 12px;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.15s;
}

.type-tab.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
  font-weight: 600;
}

.post-list {
  display: flex;
  flex-direction: column;
}

.post-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 13px 6px;
  border-bottom: 1px dashed var(--border);
}

.post-row:last-child {
  border-bottom: none;
}

.post-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.post-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--heading);
  display: flex;
  align-items: center;
  gap: 8px;
}

.post-row:hover .post-title {
  color: var(--accent);
}

.type-badge {
  font-style: normal;
  font-size: 11px;
  color: var(--accent);
  background: var(--accent-soft);
  padding: 0 8px;
  border-radius: 999px;
  flex-shrink: 0;
}

.post-summary {
  font-size: 12.5px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  flex-shrink: 0;
}

.post-views {
  font-size: 12px;
  color: var(--muted);
}

.post-date {
  font-size: 11.5px;
  color: var(--muted);
  font-family: var(--font-mono);
}

.content-card {
  padding: 26px 32px;
}
</style>
