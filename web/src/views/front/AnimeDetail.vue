<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchAnime } from '@/api'
import type { AnimeItem } from '@/types'
import { airLabel, watchLabel, progressLabel } from '@/utils/anime'
import { renderMarkdown } from '@/utils/markdown'
const route = useRoute()
const anime = ref<AnimeItem | null>(null)
const loading = ref(true),
  error = ref(false),
  imageFailed = ref(false)
const html = computed(() => renderMarkdown(anime.value?.content || ''))
const safeWatchURL = computed(() => {
  try {
    const u = new URL(anime.value?.watchUrl || '')
    return ['https:', 'http:'].includes(u.protocol) ? u.href : ''
  } catch {
    return ''
  }
})
let version = 0
async function load() {
  const current = ++version
  loading.value = true
  error.value = false
  anime.value = null
  imageFailed.value = false
  try {
    const data = await fetchAnime(String(route.params.id))
    if (current === version) anime.value = data
  } catch {
    if (current === version) error.value = true
  } finally {
    if (current === version) loading.value = false
  }
  if (current === version && route.hash === '#review') {
    await nextTick()
    document.getElementById('review')?.scrollIntoView()
  }
}
watch(() => route.params.id, load, { immediate: true })
</script>
<template>
  <div class="container detail-page">
    <router-link to="/anime" class="back">返回我的片单</router-link>
    <p v-if="loading" aria-live="polite">正在加载作品…</p>
    <el-empty v-else-if="error" description="作品不存在、已下架或暂时无法加载"
      ><el-button @click="load">重试</el-button></el-empty
    >
    <template v-else-if="anime">
      <header class="detail-head">
        <div class="poster">
          <img
            v-if="anime.cover && !imageFailed"
            :src="anime.cover"
            :alt="anime.title"
            @error="imageFailed = true"
          /><span v-else>{{ anime.title.slice(0, 1) }}</span>
        </div>
        <div class="intro">
          <p class="muted">
            {{
              [anime.region, anime.year || '', anime.category, anime.source]
                .filter(Boolean)
                .join(' · ')
            }}
          </p>
          <h1>{{ anime.title }}</h1>
          <p>
            {{ airLabel(anime.airStatus)
            }}<span v-if="anime.totalEpisodes">
              · 全 {{ anime.totalEpisodes }} 集</span
            >
          </p>
          <p class="description">
            {{ anime.description || '作品简介待补充。' }}
          </p>
          <a
            v-if="safeWatchURL"
            :href="safeWatchURL"
            data-analytics-action="watch_link"
            target="_blank"
            rel="noopener noreferrer"
            class="watch-link"
            >前往{{ anime.watchPlatform || '正版平台' }}观看 ↗</a
          >
        </div>
      </header>
      <section class="record card">
        <h2>我的观看记录</h2>
        <div class="record-meta">
          <span>{{ watchLabel(anime.watchStatus) }}</span
          ><span>已看 {{ progressLabel(anime) }}</span
          ><span
            >个人评分：{{
              anime.rating == null
                ? '未评分'
                : `${anime.rating.toFixed(1)} / 10`
            }}</span
          >
        </div>
        <progress
          v-if="anime.totalEpisodes"
          :value="anime.watched"
          :max="anime.totalEpisodes"
          aria-label="观看进度"
        />
        <blockquote v-if="anime.review">{{ anime.review }}</blockquote>
        <p v-else class="muted">短评待更新。</p>
      </section>
      <section id="review" class="review-section">
        <div class="review-head">
          <h2>观后感</h2>
          <small v-if="anime.content && anime.reviewUpdatedAt"
            >{{ anime.reviewUpdatedAt.slice(0, 10) }} 更新</small
          >
        </div>
        <p v-if="!anime.content" class="muted">这部作品的观后感还在整理中。</p>
        <details v-else-if="anime.spoiler" :key="anime.id" class="spoiler">
          <summary>含剧情剧透，点击展开观后感</summary>
          <div class="md-content" v-html="html" />
        </details>
        <div v-else class="md-content" v-html="html" />
      </section>
    </template>
  </div>
</template>
<style scoped>
.detail-page {
  max-width: 960px;
}
.back {
  display: inline-block;
  color: var(--muted);
  margin-bottom: 24px;
}
.detail-head {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 32px;
  margin-bottom: 32px;
}
.poster {
  aspect-ratio: 2/3;
  display: grid;
  place-items: center;
  background: var(--surface);
  border-radius: 10px;
  overflow: hidden;
  font: 60px var(--font-heading);
  color: var(--muted);
}
.poster img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
h1 {
  font: 700 32px var(--font-heading);
  margin: 10px 0 18px;
}
h2 {
  font: 700 22px var(--font-heading);
}
.description {
  line-height: 1.9;
}
.muted,
small {
  color: var(--muted);
}
.watch-link {
  color: var(--accent);
  text-decoration: underline;
  text-underline-offset: 5px;
  display: inline-block;
  margin-top: 14px;
}
.record {
  padding: 24px;
}
.record h2 {
  margin-top: 0;
}
.record-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}
progress {
  width: 100%;
  height: 6px;
  accent-color: var(--accent);
  margin-top: 20px;
}
blockquote {
  margin: 24px 0 0;
  padding-left: 16px;
  border-left: 3px solid var(--accent);
  line-height: 1.8;
}
.review-section {
  margin-top: 32px;
  scroll-margin-top: 125px;
}
.review-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}
.spoiler {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 18px;
}
.spoiler summary {
  cursor: pointer;
  color: var(--accent);
}
.spoiler[open] summary {
  margin-bottom: 24px;
}
.md-content {
  overflow-wrap: anywhere;
}
@media (max-width: 600px) {
  .detail-head {
    grid-template-columns: 110px minmax(0, 1fr);
    gap: 18px;
    align-items: start;
  }
  h1 {
    font-size: 25px;
  }
  .record {
    padding: 18px;
  }
  .intro {
    overflow-wrap: anywhere;
  }
}
</style>
