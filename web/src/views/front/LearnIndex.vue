<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { motion } from 'motion-v'
import { fetchBooks } from '@/api'
import type { Book } from '@/types'
import { getReadingHistory, timeAgo, type ReadingRecord } from '@/utils/history'

const route = useRoute()
const books = ref<Book[]>([])
const loading = ref(false)
const lang = ref<string>((route.query.lang as string) || '')
const recent = ref<ReadingRecord[]>([])
const hovered = ref(-1)

const hoverAnim = { scale: 1.02, y: -4 }
const spring = { type: 'spring', stiffness: 320, damping: 28 }

async function load() {
  loading.value = true
  try {
    books.value = (await fetchBooks({ language: lang.value || undefined, type: 0 })) || []
  } finally {
    loading.value = false
  }
}

function switchLang(l: string) {
  lang.value = lang.value === l ? '' : l
  load()
}

onMounted(() => {
  load()
  recent.value = getReadingHistory('learn').slice(0, 4)
})
</script>

<template>
  <div class="container learn-wrap">
    <div class="learn-hero card">
      <h1 class="hero-title">双语书房</h1>
      <p class="hero-sub">
        读英文名著与德语故事：<b>鼠标移到单词</b>上看释义，<b>点击句子</b>显示整句中文翻译，再点一次隐藏。
      </p>
      <div class="lang-tabs">
        <button class="lang-tab" :class="{ active: lang === '' }" @click="switchLang('')">全部</button>
        <button class="lang-tab" :class="{ active: lang === 'en' }" @click="switchLang('en')">🇬🇧 English</button>
        <button class="lang-tab" :class="{ active: lang === 'de' }" @click="switchLang('de')">🇩🇪 Deutsch</button>
      </div>
    </div>

    <!-- 书架：竖立书脊，悬停向右倒出平面封面 -->
    <div v-if="books.length" class="shelf-scene card">
      <div class="shelf-head">
        <span class="shelf-no">01</span>
        <span class="shelf-name">珍藏书架</span>
      </div>
      <div class="shelf">
        <router-link
          v-for="(b, i) in books"
          :key="b.id"
          :to="`/learn/book/${b.id}`"
          class="bs-book"
          :class="`lang-${b.language}`"
          :style="{ zIndex: hovered === i ? 30 : books.length - i }"
          @mouseenter="hovered = i"
          @mouseleave="hovered = -1"
        >
          <span class="bs-spine"><span class="bs-spine-text">{{ b.subtitle || b.title }}</span></span>
          <span class="bs-cover">
            <span class="bs-cover-title">{{ b.subtitle || b.title }}</span>
            <span class="bs-cover-bottom">
              <span class="bs-cover-name">{{ b.title }}</span>
              <span class="bs-cover-author">{{ b.author || '佚名' }}</span>
            </span>
          </span>
        </router-link>
        <div class="shelf-board" />
      </div>
      <p class="shelf-caption">
        <template v-if="hovered >= 0">《{{ books[hovered].title }}》 ↗</template>
        <template v-else>把鼠标移到书脊上，抽出一本看看</template>
      </p>
    </div>

    <!-- 最近阅读：点击继续读 -->
    <div v-if="recent.length" class="recent-bar card">
      <h3 class="recent-title">🕘 最近阅读</h3>
      <div class="recent-list">
        <router-link
          v-for="r in recent"
          :key="r.chapterId"
          :to="`/learn/chapter/${r.chapterId}`"
          class="recent-item"
        >
          <span class="recent-lang" :class="r.language">{{ r.language === 'de' ? 'DE' : 'EN' }}</span>
          <span class="recent-info">
            <span class="recent-chapter">{{ r.chapterTitle }}</span>
            <span class="recent-book">《{{ r.bookTitle }}》</span>
          </span>
          <span class="recent-time">{{ timeAgo(r.at) }}</span>
        </router-link>
      </div>
    </div>

    <div class="book-grid" v-loading="loading">
      <motion.div
        v-for="b in books"
        :key="b.id"
        class="book-cell"
        :while-hover="hoverAnim"
        :transition="spring"
      >
        <router-link :to="`/learn/book/${b.id}`" class="book-card card">
        <div
          class="book-cover"
          :class="b.language"
          :style="b.cover ? { background: `url(${b.cover}) center/cover` } : {}"
        >
          <template v-if="!b.cover">
            <span class="cover-lang">{{ b.language === 'de' ? 'DE' : 'EN' }}</span>
            <span class="cover-title">{{ b.subtitle || b.title }}</span>
          </template>
        </div>
        <div class="book-info">
          <h2 class="book-title">{{ b.title }}</h2>
          <p v-if="b.subtitle && b.title !== b.subtitle" class="book-subtitle">{{ b.subtitle }}</p>
          <p class="book-author">{{ b.author || '佚名' }}</p>
          <p class="book-desc">{{ b.description }}</p>
          <div class="book-meta">
            <span class="book-lang">{{ b.language === 'de' ? '德语' : '英语' }}</span>
            <span class="book-chapters">{{ b.chapterCount || 0 }} 章</span>
          </div>
        </div>
        </router-link>
      </motion.div>
    </div>
    <el-empty v-if="!loading && !books.length" description="书架空空如也，去后台添加书籍吧" />
  </div>
</template>

<style scoped>
.learn-wrap {
  max-width: 1020px;
}

.learn-hero {
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
  margin: 0 0 18px;
}

.hero-sub b {
  color: var(--accent);
  font-weight: 600;
}

.lang-tabs {
  display: inline-flex;
  gap: 8px;
  background: var(--surface);
  padding: 5px;
  border-radius: 999px;
}

.lang-tab {
  border: none;
  background: transparent;
  padding: 7px 20px;
  border-radius: 999px;
  font-size: 14px;
  color: var(--muted);
  cursor: pointer;
  transition: all 0.2s;
}

.lang-tab.active {
  background: var(--accent);
  color: #fff;
  font-weight: 600;
  box-shadow: var(--shadow);
}

/* ===== 珍藏书架：竖立书脊，悬停向右倒出封面 ===== */
.shelf-scene {
  padding: 22px 26px 14px;
  margin-bottom: 22px;
}

.shelf-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.shelf-no {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 700;
  color: var(--accent);
  border: 1.5px solid var(--accent);
  border-radius: 6px;
  padding: 1px 8px;
  background: repeating-linear-gradient(
    -45deg,
    transparent,
    transparent 4px,
    color-mix(in srgb, var(--accent) 7%, transparent) 4px,
    color-mix(in srgb, var(--accent) 7%, transparent) 8px
  );
}

.shelf-name {
  font-family: var(--font-heading);
  font-size: 17px;
  color: var(--heading);
  font-weight: 700;
  letter-spacing: 2px;
}

.shelf {
  position: relative;
  display: flex;
  align-items: flex-end;
  gap: 4px;
  padding: 6px 34px 18px;
  overflow-x: auto;
  height: 193px;
}

.bs-book {
  position: relative;
  flex-shrink: 0;
  width: 18px;
  height: 184px;
  transform: rotate(3deg); /* 静止时整体向右微倾 */
  transform-origin: bottom center;
  transition:
    width 0.45s cubic-bezier(0.3, 1.15, 0.45, 1),
    transform 0.45s cubic-bezier(0.3, 1.15, 0.45, 1);
  cursor: pointer;
}

/* 悬停：展开成平面封面 126x184，立直 */
.bs-book:hover {
  width: 126px;
  transform: rotate(0deg);
  z-index: 30 !important;
}

/* cali.so 效果：被抽书的邻居向两侧倒开（在静止倾角基础上加/减） */
.bs-book:hover + .bs-book {
  transform: rotate(6.5deg);
  transform-origin: bottom left;
}

.bs-book:has(+ .bs-book:hover) {
  transform: rotate(-1.5deg);
  transform-origin: bottom right;
}

.bs-spine {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: 18px;
  border-radius: 2px 5px 2px 2px;
  border: 1px solid rgb(38 30 23 / 0.16);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: inset -2px 0 5px rgb(0 0 0 / 0.18);
  transition: opacity 0.25s;
}

.bs-book:hover .bs-spine {
  opacity: 0;
}

.bs-book.lang-en .bs-spine {
  background: linear-gradient(180deg, #4d5940, #38422e);
}

.bs-book.lang-de .bs-spine {
  background: linear-gradient(180deg, #5d3c2c, #42291e);
}

.bs-book.lang-zh .bs-spine {
  background: linear-gradient(180deg, #a13a28, #7d2b1c);
}

.bs-spine-text {
  writing-mode: vertical-rl;
  max-height: 94%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #f5efe2;
  font-family: var(--font-heading);
  font-size: 11px;
  letter-spacing: 1px;
  text-shadow: 0 1px 2px rgb(0 0 0 / 0.3);
}

.bs-cover {
  position: absolute;
  inset: 0;
  opacity: 0;
  border-radius: 2px 8px 8px 2px;
  border: 1px solid var(--border);
  background: var(--card);
  padding: 14px 12px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  box-shadow: 6px 4px 16px rgb(38 30 23 / 0.2);
  transition: opacity 0.3s 0.12s;
}

.bs-book:hover .bs-cover {
  opacity: 1;
}

.bs-cover::before {
  content: '';
  position: absolute;
  inset: 7px;
  border: 1px solid var(--border);
  border-radius: 2px 5px 5px 2px;
  pointer-events: none;
}

.bs-cover-title {
  font-family: var(--font-heading);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.45;
  color: var(--heading);
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.bs-cover-bottom {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.bs-cover-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.bs-cover-author {
  font-size: 11px;
  color: var(--muted);
  font-style: italic;
}

.shelf-board {
  position: absolute;
  left: 14px;
  right: 14px;
  bottom: 0;
  height: 13px;
  border-radius: 4px 4px 8px 8px;
  background: linear-gradient(180deg, #8a6a4d, #5f4530);
  box-shadow: 0 6px 14px rgb(38 30 23 / 0.22);
  pointer-events: none;
}

.shelf-caption {
  text-align: center;
  font-size: 13.5px;
  color: var(--muted);
  margin: 14px 0 4px;
  min-height: 20px;
}

/* ===== 最近阅读 ===== */
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

.recent-lang {
  flex-shrink: 0;
  font-size: 11px;
  font-family: var(--font-mono);
  font-weight: 700;
  border-radius: 6px;
  padding: 1px 8px;
}

.recent-lang.en {
  color: var(--olive);
  background: var(--badge-en-bg);
}

.recent-lang.de {
  color: var(--accent-hover);
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

@media (max-width: 760px) {
  .recent-list {
    grid-template-columns: 1fr;
  }
}

.book-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 18px;
  min-height: 200px;
}

/* 最后一行只剩一张卡时（含筛选后仅 1 本），占满整行与上方标题卡对齐 */
.book-grid > *:last-child:nth-child(odd) {
  grid-column: 1 / -1;
}

.book-card {
  display: flex;
  gap: 18px;
  padding: 20px;
  height: 100%;
  transition: box-shadow 0.25s ease, border-color 0.2s ease;
}

.book-card:hover {
  box-shadow: var(--shadow-lift);
  border-color: var(--accent);
}

.book-cover {
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
  position: relative;
  overflow: hidden;
}

.book-cover.en {
  background: linear-gradient(150deg, #525e45, #3d4734 130%);
}

.book-cover.de {
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
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.book-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.book-title {
  font-family: var(--font-heading);
  font-size: 20px;
  margin: 0 0 2px;
}

.book-subtitle {
  font-size: 13px;
  color: var(--muted);
  font-style: italic;
  margin: 0 0 6px;
}

.book-author {
  font-size: 13px;
  color: var(--muted);
  margin: 0 0 8px;
}

.book-desc {
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

.book-meta {
  display: flex;
  gap: 8px;
}

.book-lang {
  font-size: 12px;
  color: var(--olive);
  background: var(--badge-en-bg);
  border: 1px solid var(--badge-en-border);
  padding: 1px 10px;
  border-radius: 999px;
}

.book-chapters {
  font-size: 12px;
  color: var(--muted);
}

@media (max-width: 760px) {
  .book-grid {
    grid-template-columns: 1fr;
  }
}
</style>
