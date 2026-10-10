<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchGames, fetchLatestGamePosts } from '@/api'
import type { GameItem, GamePostItem } from '@/types'
import GameCard from '@/components/GameCard.vue'
import { playStatuses, splitGameTags } from '@/utils/game'
const games = ref<GameItem[]>([])
const posts = ref<GamePostItem[]>([])
const keyword = ref('')
const category = ref('')
const platform = ref('')
const playStatus = ref('')
const loading = ref(true)
const error = ref(false)
const postsError = ref(false)
const categories = computed(() => [...new Set(games.value.map(g => g.category).filter(Boolean))])
const platforms = computed(() => [...new Set(games.value.flatMap(g => splitGameTags(g.platform)))])
const filtered = computed(() => games.value.filter(g =>
  g.title.toLowerCase().includes(keyword.value.trim().toLowerCase()) &&
  (!category.value || g.category === category.value) &&
  (!platform.value || splitGameTags(g.platform).includes(platform.value)) &&
  (!playStatus.value || g.playStatus === playStatus.value)))
const playing = computed(() => [...games.value].filter(g => g.playStatus === 'playing').sort((a,b) => Date.parse(b.updatedAt) - Date.parse(a.updatedAt))[0])
const recommendations = computed(() => games.value.filter(g => g.recommended).sort((a,b) => a.recommendOrder - b.recommendOrder || a.id - b.id).slice(0, 5))
const gameName = (id: number) => games.value.find(g => g.id === id)?.title || ''
function reset() { keyword.value = ''; category.value = ''; platform.value = ''; playStatus.value = '' }
async function loadPosts() {
  postsError.value = false
  try { posts.value = (await fetchLatestGamePosts()) || [] } catch { postsError.value = true }
}
async function load() {
  loading.value = true; error.value = false
  try { games.value = (await fetchGames()) || [] } catch { error.value = true }
  finally { loading.value = false }
}
onMounted(() => { void load(); void loadPosts() })
</script>
<template>
  <div class="container game-page">
    <header class="page-heading"><h1>游戏与手记</h1><p>记录在玩的世界，也留下值得分享的经验。</p></header>
    <div v-if="error" class="card notice" role="alert">游戏库加载失败。<el-button text @click="load">重新加载</el-button></div>
    <template v-else>
      <section v-if="playing" class="playing card">
        <img v-if="playing.cover" :src="playing.cover" :alt="playing.title" />
        <div><span class="section-label">最近在玩</span><h2>{{ playing.title }}</h2><p>{{ playing.review || playing.description }}</p><p v-if="playing.progress" class="muted">当前进度：{{ playing.progress }}</p><router-link :to="`/game/${playing.id}`" class="entry">查看游玩记录</router-link></div>
      </section>
      <div class="layout">
        <section class="library" aria-labelledby="library-title">
          <div class="library-heading"><h2 id="library-title">我的游戏库</h2><span>{{ filtered.length }} 款游戏</span></div>
          <div class="filters">
            <el-input v-model="keyword" clearable placeholder="搜索游戏名称" aria-label="搜索游戏名称" />
            <el-select v-model="category" clearable placeholder="全部类型" aria-label="游戏类型"><el-option v-for="c in categories" :key="c" :value="c" :label="c" /></el-select>
            <el-select v-model="platform" clearable placeholder="全部平台" aria-label="游戏平台"><el-option v-for="p in platforms" :key="p" :value="p" :label="p" /></el-select>
            <el-select v-model="playStatus" clearable placeholder="全部游玩状态" aria-label="游玩状态"><el-option v-for="s in playStatuses" :key="s.value" :value="s.value" :label="s.label" /></el-select>
          </div>
          <div v-if="loading" class="notice" aria-live="polite">正在加载游戏库…</div>
          <div v-else-if="filtered.length" class="game-grid"><GameCard v-for="g in filtered" :key="g.id" :game="g" /></div>
          <el-empty v-else :description="games.length ? '没有符合条件的游戏' : '游戏手记正在整理中'"><el-button v-if="games.length" @click="reset">清除筛选</el-button></el-empty>
        </section>
        <aside>
          <section class="side-section"><h2>站长推荐</h2><p v-if="!recommendations.length" class="muted">推荐清单正在整理中。</p><router-link v-for="g in recommendations" :key="g.id" :to="`/game/${g.id}`" class="side-item"><strong>{{ g.title }}</strong><p>{{ g.review || g.description }}</p></router-link></section>
          <section class="side-section"><h2>最新攻略</h2><p v-if="postsError" role="alert">攻略加载失败。<el-button text @click="loadPosts">重试</el-button></p><p v-else-if="!posts.length" class="muted">新的攻略正在路上。</p><router-link v-for="p in posts" :key="p.id" :to="`/game/post/${p.id}`" class="side-item"><span class="muted">{{ gameName(p.gameId) }} · {{ p.type }}</span><strong>{{ p.title }}</strong><small>{{ p.updatedAt.slice(0, 10) }} 更新</small></router-link></section>
        </aside>
      </div>
    </template>
  </div>
</template>
<style scoped>
.game-page { max-width: 1200px; }
.page-heading { margin: 6px 0 28px; }
h1 { font: 700 34px var(--font-heading); margin: 0 0 10px; }
.page-heading p,.muted { color: var(--muted); }
h2 { font: 700 22px var(--font-heading); margin: 0 0 16px; }
.playing { display: grid; grid-template-columns: minmax(220px, 42%) 1fr; overflow: hidden; margin-bottom: 36px; }
.playing img { width: 100%; height: 100%; max-height: 320px; object-fit: cover; }
.playing > div { padding: 28px; }
.playing:not(:has(img)) { grid-template-columns: 1fr; }
.playing h2 { font-size: 30px; margin-top: 10px; }
.playing p { line-height: 1.8; }
.section-label,.entry { color: var(--accent); font-size: 14px; }
.entry { text-decoration: underline; text-underline-offset: 5px; }
.layout { display: grid; grid-template-columns: minmax(0,1fr) 260px; gap: 36px; }
.library-heading { display: flex; align-items: baseline; justify-content: space-between; }
.library-heading span { color: var(--muted); font-size: 13px; }
.filters { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 10px; margin-bottom: 20px; }
.game-grid { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 20px; }
.side-section { margin-bottom: 30px; border-top: 2px solid var(--border); padding-top: 18px; }
.side-item { display: block; padding: 14px 0; border-bottom: 1px solid var(--border); color: var(--text); }
.side-item strong { display: block; margin: 5px 0; }
.side-item p { font-size: 13px; color: var(--muted); line-height: 1.7; margin: 5px 0; }
.side-item small,.side-item > span { color: var(--muted); font-size: 12px; }
.side-item:hover strong { color: var(--accent); }
.notice { padding: 30px; }
@media(max-width: 900px) { .layout { grid-template-columns: 1fr; } aside { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; } }
@media(max-width: 600px) { .playing,.game-grid,aside { grid-template-columns: 1fr; } .playing img { max-height: 220px; } .playing > div { padding: 20px; } h1 { font-size: 28px; } }
</style>
