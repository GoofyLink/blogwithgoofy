<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { motion } from 'motion-v'
import { fetchBooks } from '@/api'
import type { Book } from '@/types'
import { getReadingHistory, timeAgo, type ReadingRecord } from '@/utils/history'

const books = ref<Book[]>([])
const loading = ref(false)
const recent = ref<ReadingRecord[]>([])

const hoverAnim = { scale: 1.02, y: -4 }
const spring = { type: 'spring', stiffness: 320, damping: 28 }

onMounted(async () => {
  recent.value = getReadingHistory('novel').slice(0, 4)
  loading.value = true
  try {
    books.value = (await fetchBooks({ type: 1 })) || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container novel-wrap">
    <div class="novel-hero card">
      <h1 class="hero-title">小说专栏</h1>
      <p class="hero-sub">安静读故事的地方。支持目录跳转、字号调节、听书朗读。</p>
    </div>

    <!-- 最近阅读 -->
    <div v-if="recent.length" class="recent-bar card">
      <h3 class="recent-title">🕘 最近阅读</h3>
      <div class="recent-list">
        <router-link
          v-for="r in recent"
          :key="r.chapterId"
          :to="`/novel/chapter/${r.chapterId}`"
          class="recent-item"
        >
          <span class="recent-info">
            <span class="recent-chapter">{{ r.chapterTitle }}</span>
            <span class="recent-book">《{{ r.bookTitle }}》</span>
          </span>
          <span class="recent-time">{{ timeAgo(r.at) }}</span>
        </router-link>
      </div>
    </div>

    <div class="novel-grid" v-loading="loading">
      <motion.div
        v-for="b in books"
        :key="b.id"
        class="novel-cell"
        :while-hover="hoverAnim"
        :transition="spring"
      >
        <router-link :to="`/novel/book/${b.id}`" class="novel-card card">
          <div class="novel-cover" :style="b.cover ? { background: `url(${b.cover}) center/cover` } : {}">
            <template v-if="!b.cover">
              <span class="cover-mark">说</span>
              <span class="cover-title">{{ b.title }}</span>
            </template>
          </div>
          <div class="novel-info">
            <h2 class="novel-title">{{ b.title }}</h2>
            <p class="novel-author">{{ b.author || '佚名' }}</p>
            <p class="novel-desc">{{ b.description }}</p>
            <span class="novel-chapters">{{ b.chapterCount || 0 }} 章</span>
          </div>
        </router-link>
      </motion.div>
    </div>
    <el-empty v-if="!loading && !books.length" description="还没有小说，去后台「书籍管理」添加（类型选「小说专栏」），整本 TXT 批量导入即可" />
  </div>
</template>

<style scoped>
.novel-wrap {
  max-width: 1020px;
}

.novel-hero {
  padding: 34px 40px;
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

.recent-bar {
  padding: 18px 24px;
  margin-bottom: 22px;
}

.recent-title {
  font-family: var(--font-heading);
  font-size: 16px;
  margin: 0 0 12px;
}

.recent-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 18px;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 12px;
  border-radius: 10px;
  transition: background 0.15s;
  min-width: 0;
}

.recent-item:hover {
  background: var(--accent-soft);
}

.recent-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.recent-chapter {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--heading);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recent-item:hover .recent-chapter {
  color: var(--accent);
}

.recent-book {
  font-size: 12px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recent-time {
  flex-shrink: 0;
  font-size: 11.5px;
  color: var(--muted);
}

.novel-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 18px;
  min-height: 200px;
}

.novel-grid > *:last-child:nth-child(odd) {
  grid-column: 1 / -1;
}

.novel-card {
  display: flex;
  gap: 18px;
  padding: 20px;
  height: 100%;
  transition: box-shadow 0.25s ease, border-color 0.2s ease;
}

.novel-card:hover {
  box-shadow: var(--shadow-lift);
  border-color: var(--accent);
}

.novel-cover {
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
  background: linear-gradient(150deg, #525e45, #3d4734 130%);
}

.cover-mark {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  background: rgb(255 255 255 / 0.14);
  color: #f5efe2;
  font-family: var(--font-heading);
  font-size: 22px;
  border-radius: 10px;
  margin-bottom: 10px;
}

.cover-title {
  color: #f5efe2;
  font-family: var(--font-heading);
  font-size: 15px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.novel-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.novel-title {
  font-family: var(--font-heading);
  font-size: 20px;
  margin: 0 0 2px;
}

.novel-author {
  font-size: 13px;
  color: var(--muted);
  margin: 0 0 8px;
}

.novel-desc {
  font-size: 13px;
  color: var(--muted);
  line-height: 1.65;
  margin: 0 0 10px;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.novel-chapters {
  font-size: 12px;
  color: var(--muted);
}

@media (max-width: 760px) {
  .novel-grid {
    grid-template-columns: 1fr;
  }

  .recent-list {
    grid-template-columns: 1fr;
  }
}
</style>
