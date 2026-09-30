<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchArts } from '@/api'
import type { ArtworkItem } from '@/types'

const list = ref<ArtworkItem[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    list.value = (await fetchArts()) || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container art-wrap">
    <div class="art-hero card">
      <h1 class="hero-title">艺术鉴赏</h1>
      <p class="hero-sub">收集喜欢的画面与创作，点击图片可放大欣赏。</p>
    </div>

    <div class="art-grid" v-loading="loading">
      <div v-for="a in list" :key="a.id" class="art-card card">
        <el-image
          v-if="a.image"
          :src="a.image"
          :preview-src-list="[a.image]"
          fit="cover"
          class="art-img"
          hide-on-click-modal
          preview-teleported
        />
        <div v-else class="art-img art-placeholder">
          <span>{{ a.title.slice(0, 1) }}</span>
        </div>
        <div class="art-info">
          <div class="art-head">
            <h3 class="art-title">{{ a.title }}</h3>
            <span v-if="a.author" class="art-author">{{ a.author }}</span>
          </div>
          <p v-if="a.description" class="art-desc">{{ a.description }}</p>
        </div>
      </div>
    </div>
    <el-empty v-if="!loading && !list.length" description="画廊还是空的，去后台「艺术鉴赏管理」添加作品吧" />
  </div>
</template>

<style scoped>
.art-wrap {
  max-width: 1020px;
}

.art-hero {
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
  margin: 0;
}

.art-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  min-height: 200px;
}

.art-card {
  overflow: hidden;
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.art-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lift);
}

.art-img {
  width: 100%;
  height: 320px;
  display: block;
}

.art-img :deep(img) {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.4s ease;
}

.art-card:hover .art-img :deep(img) {
  transform: scale(1.03);
}

.art-placeholder {
  display: grid;
  place-items: center;
  background: linear-gradient(150deg, var(--surface-alt), var(--surface));
}

.art-placeholder span {
  font-family: var(--font-heading);
  font-size: 60px;
  font-weight: 700;
  color: var(--muted);
  opacity: 0.5;
}

.art-info {
  padding: 16px 20px 20px;
}

.art-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.art-title {
  font-family: var(--font-heading);
  font-size: 18px;
  font-weight: 700;
  color: var(--heading);
  margin: 0 0 6px;
}

.art-author {
  font-size: 12px;
  color: var(--muted);
  font-style: italic;
  flex-shrink: 0;
}

.art-desc {
  font-size: 13.5px;
  color: var(--text);
  line-height: 1.7;
  margin: 0;
}

@media (max-width: 760px) {
  .art-grid {
    grid-template-columns: 1fr;
  }
}
</style>
