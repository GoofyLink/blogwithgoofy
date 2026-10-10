<script setup lang="ts">
import { ref, watch } from 'vue'
import type { AnimeItem } from '@/types'
import { watchLabel, progressLabel } from '@/utils/anime'
const props = defineProps<{ anime: AnimeItem }>()
const failed = ref(false)
watch(
  () => props.anime.cover,
  () => {
    failed.value = false
  },
)
</script>
<template>
  <router-link :to="`/anime/${anime.id}`" class="anime-card">
    <div class="poster">
      <img
        v-if="anime.cover && !failed"
        :src="anime.cover"
        :alt="anime.title"
        loading="lazy"
        @error="failed = true"
      /><span v-else>{{ anime.title.slice(0, 1) }}</span
      ><span v-if="anime.watchStatus" class="state">{{
        watchLabel(anime.watchStatus)
      }}</span>
    </div>
    <div class="line">
      <span>{{ anime.category || '未分类' }}</span
      ><span>{{
        anime.rating == null ? '未评分' : `${anime.rating.toFixed(1)} 分`
      }}</span>
    </div>
    <h3>{{ anime.title }}</h3>
    <p class="progress">已看 {{ progressLabel(anime) }}</p>
    <progress
      v-if="anime.totalEpisodes"
      :value="anime.watched"
      :max="anime.totalEpisodes"
      :aria-label="`${anime.title}观看进度`"
    />
    <p class="review">
      {{ anime.review || anime.description || '观影心得待更新' }}
    </p>
  </router-link>
</template>
<style scoped>
.anime-card {
  display: block;
  min-width: 0;
  color: var(--text);
}
.poster {
  position: relative;
  aspect-ratio: 2/3;
  background: var(--surface);
  border-radius: 10px;
  overflow: hidden;
  display: grid;
  place-items: center;
  color: var(--muted);
  font: 48px var(--font-heading);
}
.poster img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.poster .state {
  position: absolute;
  bottom: 12px;
  left: 12px;
  border-radius: 5px;
  padding: 4px 9px;
  font: 12px var(--font-body);
  background: var(--card);
  color: var(--text);
}
.line {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  margin-top: 14px;
  color: var(--muted);
  font-size: 12px;
}
h3 {
  font: 700 19px var(--font-heading);
  margin: 6px 0;
  overflow-wrap: anywhere;
}
.progress {
  font-size: 12px;
  color: var(--muted);
  margin: 5px 0;
}
progress {
  width: 100%;
  height: 4px;
  accent-color: var(--accent);
  display: block;
  margin: 9px 0;
}
.review {
  color: var(--muted);
  font-size: 13px;
  line-height: 1.7;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.anime-card:hover h3 {
  color: var(--accent);
}
.anime-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 5px;
}
</style>
