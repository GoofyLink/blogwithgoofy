<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchAnimes } from '@/api'
import type { AnimeItem } from '@/types'
import AnimeCard from '@/components/AnimeCard.vue'
import { watchStatuses, progressLabel } from '@/utils/anime'
const all = ref<AnimeItem[]>([])
const keyword = ref(''),
  category = ref(''),
  region = ref(''),
  status = ref('')
const loading = ref(true),
  error = ref(false)
const categories = computed(() => [
  ...new Set(all.value.map((a) => a.category).filter(Boolean)),
])
const regions = computed(() => [
  ...new Set(all.value.map((a) => a.region).filter(Boolean)),
])
const filtered = computed(() =>
  all.value.filter(
    (a) =>
      a.title.toLowerCase().includes(keyword.value.trim().toLowerCase()) &&
      (!category.value || a.category === category.value) &&
      (!region.value || a.region === region.value) &&
      (!status.value || a.watchStatus === status.value),
  ),
)
const watching = computed(() =>
  all.value
    .filter((a) => a.watchStatus === 'watching')
    .sort(
      (a, b) =>
        Date.parse(b.progressUpdatedAt || b.updatedAt) -
        Date.parse(a.progressUpdatedAt || a.updatedAt),
    )
    .slice(0, 3),
)
const recommended = computed(() =>
  all.value
    .filter((a) => a.recommended)
    .sort((a, b) => a.recommendOrder - b.recommendOrder || a.id - b.id)
    .slice(0, 5),
)
const ranking = computed(() =>
  all.value
    .filter((a) => a.rank > 0)
    .sort((a, b) => a.rank - b.rank || a.id - b.id)
    .slice(0, 8),
)
const reviews = computed(() =>
  all.value
    .filter((a) => a.content?.trim())
    .sort(
      (a, b) =>
        Date.parse(b.reviewUpdatedAt || b.updatedAt) -
        Date.parse(a.reviewUpdatedAt || a.updatedAt),
    )
    .slice(0, 5),
)
function reset() {
  keyword.value = ''
  category.value = ''
  region.value = ''
  status.value = ''
}
async function load() {
  loading.value = true
  error.value = false
  try {
    all.value = (await fetchAnimes()) || []
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>
<template>
  <div class="container anime-page">
    <header class="heading">
      <h1>追番与观后感</h1>
      <p>收藏喜欢的故事，记录每一次心动。</p>
    </header>
    <div v-if="loading" class="notice" aria-live="polite">正在加载片单…</div>
    <el-empty v-else-if="error" description="片单加载失败"
      ><el-button @click="load">重新加载</el-button></el-empty
    >
    <template v-else>
      <section v-if="watching.length" class="watching">
        <h2>最近在追</h2>
        <div class="watching-list">
          <router-link
            v-for="a in watching"
            :key="a.id"
            :to="`/anime/${a.id}`"
            class="watching-item"
            ><span class="now-dot" aria-hidden="true" />
            <div>
              <strong>{{ a.title }}</strong>
              <p>已看 {{ progressLabel(a) }}</p>
            </div>
            <span>查看记录</span></router-link
          >
        </div>
      </section>
      <div class="layout">
        <section class="library">
          <div class="section-heading">
            <h2>我的片单</h2>
            <span>{{ filtered.length }} 部作品</span>
          </div>
          <div class="filters">
            <el-input
              v-model="keyword"
              clearable
              placeholder="搜索动漫名称"
              aria-label="搜索动漫名称"
            /><el-select
              v-model="status"
              clearable
              placeholder="全部观看状态"
              aria-label="观看状态"
              ><el-option
                v-for="s in watchStatuses"
                :key="s.value"
                :label="s.label"
                :value="s.value" /></el-select
            ><el-select
              v-model="region"
              clearable
              placeholder="全部地区"
              aria-label="地区"
              ><el-option
                v-for="r in regions"
                :key="r"
                :label="r"
                :value="r" /></el-select
            ><el-select
              v-model="category"
              clearable
              placeholder="全部题材"
              aria-label="题材"
              ><el-option
                v-for="c in categories"
                :key="c"
                :label="c"
                :value="c"
            /></el-select>
          </div>
          <div v-if="filtered.length" class="anime-grid">
            <AnimeCard v-for="a in filtered" :key="a.id" :anime="a" />
          </div>
          <el-empty
            v-else
            :description="all.length ? '没有符合条件的作品' : '片单正在整理中'"
            ><el-button v-if="all.length" @click="reset"
              >清除筛选</el-button
            ></el-empty
          >
        </section>
        <aside>
          <section class="side-section">
            <h2>私藏推荐</h2>
            <p v-if="!recommended.length" class="muted">推荐清单正在整理中。</p>
            <router-link
              v-for="a in recommended"
              :key="a.id"
              :to="`/anime/${a.id}`"
              class="side-item"
              ><strong>{{ a.title }}</strong>
              <p>{{ a.review || a.description }}</p></router-link
            >
          </section>
          <section class="side-section">
            <h2>个人喜爱榜</h2>
            <p class="muted">按我的喜爱程度排序。</p>
            <p v-if="!ranking.length" class="muted">还没有设置排名。</p>
            <router-link
              v-for="a in ranking"
              :key="a.id"
              :to="`/anime/${a.id}`"
              class="rank-item"
              ><span>{{ a.rank }}</span
              ><strong>{{ a.title }}</strong></router-link
            >
          </section>
          <section class="side-section">
            <h2>最新观后感</h2>
            <p v-if="!reviews.length" class="muted">看完再来慢慢写。</p>
            <router-link
              v-for="a in reviews"
              :key="a.id"
              :to="`/anime/${a.id}#review`"
              class="side-item"
              ><strong>{{ a.title }}</strong
              ><small
                >{{ (a.reviewUpdatedAt || a.updatedAt).slice(0, 10) }} 更新<span
                  v-if="a.spoiler"
                >
                  · 含剧透</span
                ></small
              ></router-link
            >
          </section>
        </aside>
      </div>
    </template>
  </div>
</template>
<style scoped>
.anime-page {
  max-width: 1200px;
}
.heading {
  margin: 8px 0 32px;
}
.heading h1 {
  font: 700 34px var(--font-heading);
  margin: 0 0 10px;
}
.heading p,
.muted {
  color: var(--muted);
}
h2 {
  font: 700 22px var(--font-heading);
  margin: 0 0 18px;
}
.watching {
  margin-bottom: 36px;
  padding-bottom: 26px;
  border-bottom: 1px solid var(--border);
}
.watching-list {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}
.watching-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
}
.watching-item p {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--muted);
}
.watching-item > span:last-child {
  margin-left: auto;
  font-size: 12px;
  color: var(--accent);
  white-space: nowrap;
}
.now-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--accent);
  flex-shrink: 0;
}
.layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 245px;
  gap: 36px;
}
.section-heading {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}
.section-heading span {
  font-size: 13px;
  color: var(--muted);
}
.filters {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 24px;
}
.anime-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 28px 20px;
}
.side-section {
  margin-bottom: 30px;
  border-top: 2px solid var(--border);
  padding-top: 18px;
}
.side-item {
  display: block;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
  color: var(--text);
}
.side-item strong {
  display: block;
}
.side-item p {
  font-size: 13px;
  color: var(--muted);
  line-height: 1.7;
  margin: 8px 0;
}
.side-item small {
  color: var(--muted);
  font-size: 12px;
}
.rank-item {
  display: flex;
  gap: 15px;
  padding: 12px 0;
  color: var(--text);
}
.rank-item > span {
  color: var(--accent);
  min-width: 18px;
}
.side-item:hover strong,
.rank-item:hover strong {
  color: var(--accent);
}
.notice {
  padding: 30px;
}
@media (max-width: 960px) {
  .layout {
    grid-template-columns: 1fr;
  }
  .watching-list {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 600px) {
  .anime-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px 14px;
  }
  .heading h1 {
    font-size: 28px;
  }
}
</style>
