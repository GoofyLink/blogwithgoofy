<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { fetchBook } from '@/api'
import type { BookDetailData } from '@/types'

const route = useRoute()
const data = ref<BookDetailData | null>(null)
const book = computed(() => data.value?.book)
const chapters = computed(() => data.value?.chapters || [])

const firstChapterId = computed(() => chapters.value[0]?.id)

onMounted(async () => {
  data.value = await fetchBook(route.params.id as string)
})
</script>

<template>
  <div class="container book-wrap" v-if="book">
    <div class="book-head card">
      <router-link to="/learn" class="back-link">← 双语书房</router-link>
      <div class="head-main">
        <div class="head-cover" :class="book.language">
          <span class="cover-lang">{{ book.language === 'de' ? 'DE' : 'EN' }}</span>
          <span class="cover-title">{{ book.subtitle || book.title }}</span>
        </div>
        <div class="head-info">
          <h1 class="head-title">{{ book.title }}</h1>
          <p v-if="book.subtitle && book.title !== book.subtitle" class="head-subtitle">{{ book.subtitle }}</p>
          <p class="head-author">{{ book.author || '佚名' }} · {{ book.language === 'de' ? '德语' : '英语' }} · {{ chapters.length }} 章</p>
          <p class="head-desc">{{ book.description }}</p>
          <router-link v-if="firstChapterId" :to="`/learn/chapter/${firstChapterId}`">
            <el-button type="primary">📖 开始阅读第一章</el-button>
          </router-link>
        </div>
      </div>
    </div>

    <div class="card chapter-panel">
      <h3 class="panel-title">Chapters · 章节目录</h3>
      <ol class="chapter-list">
        <li v-for="(c, i) in chapters" :key="c.id">
          <router-link :to="`/learn/chapter/${c.id}`" class="chapter-item">
            <span class="chapter-no">{{ i + 1 }}.</span>
            <span class="chapter-title">{{ c.title }}</span>
          </router-link>
        </li>
      </ol>
      <el-empty v-if="!chapters.length" description="还没有章节" />
    </div>
  </div>
</template>

<style scoped>
.book-wrap {
  max-width: 1200px;
}

.book-head {
  padding: 24px 28px;
  margin-bottom: 18px;
}

.back-link {
  font-size: 13px;
  color: var(--muted);
}

.back-link:hover {
  color: var(--accent);
}

.head-main {
  display: flex;
  gap: 24px;
  margin-top: 14px;
}

.head-cover {
  width: 118px;
  height: 162px;
  flex-shrink: 0;
  border-radius: 10px 14px 14px 10px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 12px 10px;
  text-align: center;
  box-shadow: var(--shadow);
}

.head-cover.en {
  background: linear-gradient(150deg, #525e45, #3d4734 130%);
}

.head-cover.de {
  background: linear-gradient(150deg, #261e17, #7a3b2a 140%);
}

.cover-lang {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 800;
  color: rgb(255 255 255 / 0.75);
  letter-spacing: 2px;
  margin-bottom: 8px;
}

.cover-title {
  color: #fff;
  font-family: var(--font-heading);
  font-size: 15px;
  line-height: 1.4;
}

.head-info {
  flex: 1;
  min-width: 0;
}

.head-title {
  font-family: var(--font-heading);
  font-size: 26px;
  margin: 0 0 4px;
}

.head-subtitle {
  font-size: 14px;
  color: var(--muted);
  font-style: italic;
  margin: 0 0 6px;
}

.head-author {
  font-size: 13px;
  color: var(--muted);
  margin: 0 0 10px;
}

.head-desc {
  font-size: 14px;
  color: var(--muted);
  line-height: 1.7;
  margin: 0 0 16px;
}

.chapter-panel {
  padding: 22px 28px;
}

.panel-title {
  font-family: var(--font-heading);
  font-size: 20px;
  font-weight: 700;
  text-align: center;
  color: var(--heading);
  margin: 0 0 16px;
}

/* 三列网格目录：全部章节一屏展示，外层虚线框 */
.chapter-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border: 1px dashed var(--border);
  border-radius: 10px;
  overflow: hidden;
  background: var(--card);
}

.chapter-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  padding: 10px 14px;
  font-size: 13.5px;
  min-width: 0;
  transition: background 0.15s;
}

/* 3 列下偶数行斑马纹 */
.chapter-item:nth-child(6n + 4),
.chapter-item:nth-child(6n + 5),
.chapter-item:nth-child(6n + 6) {
  background: var(--surface);
}

.chapter-item:hover {
  background: var(--accent-soft);
  color: var(--accent);
}

.chapter-no {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--muted);
  flex-shrink: 0;
}

.chapter-item:hover .chapter-no {
  color: var(--accent);
}

.chapter-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .chapter-list {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .chapter-list {
    grid-template-columns: 1fr;
  }
}
</style>
