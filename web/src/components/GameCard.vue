<script setup lang="ts">
import type { GameItem } from '@/types'
import { playLabel, splitGameTags } from '@/utils/game'
defineProps<{ game: GameItem }>()
</script>
<template>
  <router-link :to="`/game/${game.id}`" class="game-card card">
    <div class="cover"><img v-if="game.cover" :src="game.cover" :alt="game.title" loading="lazy" @error="($event.target as HTMLImageElement).style.visibility = 'hidden'" /><span v-else>{{ game.title.slice(0, 1) }}</span></div>
    <div class="body">
      <div class="meta"><span>{{ game.category || '未分类' }}</span><span v-if="game.playStatus" class="state">{{ playLabel(game.playStatus) }}</span></div>
      <h3>{{ game.title }}</h3>
      <p class="review">{{ game.review || game.description || '游玩心得待更新' }}</p>
      <p v-if="game.progress" class="progress">{{ game.progress }}</p>
      <div class="foot"><span>{{ splitGameTags(game.platform).join(' / ') }}</span><span>{{ game.postCount || 0 }} 篇攻略</span></div>
    </div>
  </router-link>
</template>
<style scoped>
.game-card { overflow: hidden; color: var(--text); display: flex; flex-direction: column; }
.game-card:hover, .game-card:focus-visible { border-color: var(--accent); }
.cover { aspect-ratio: 16/9; background: var(--surface); display: grid; place-items: center; overflow: hidden; font: 42px var(--font-heading); color: var(--muted); }
.cover img { width: 100%; height: 100%; object-fit: cover; }
.body { padding: 18px; flex: 1; display: flex; flex-direction: column; }
.meta,.foot { display: flex; justify-content: space-between; gap: 10px; color: var(--muted); font-size: 12px; flex-wrap: wrap; }
.state { color: var(--accent); }
h3 { font: 700 21px var(--font-heading); margin: 10px 0; }
.review { line-height: 1.7; margin: 0 0 15px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.progress { color: var(--muted); font-size: 13px; margin: 0 0 12px; }
.foot { margin-top: auto; padding-top: 12px; border-top: 1px solid var(--border); }
</style>
