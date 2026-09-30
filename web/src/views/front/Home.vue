<script setup lang="ts">
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { fetchArticles, fetchBook, fetchBooks } from '@/api'
import type { Article, Book, ChapterItem } from '@/types'
import { useSiteStore } from '@/stores/site'

const site = useSiteStore()

const latest = ref<Article[]>([])
const enLatest = ref<Article | null>(null)
const deLatest = ref<Article | null>(null)
const novels = ref<{ book: Book; lastChapter?: ChapterItem }[]>([])

const typeLabel = (a: Article) => {
  if (a.type === 1) return a.language === 'de' ? '学习 · DE' : '学习 · EN'
  return a.category?.name || '随笔'
}

const date = (v: string) => dayjs(v).format('YYYY-MM-DD')

onMounted(async () => {
  site.load()
  const [latestRes, enRes, deRes, booksRes] = await Promise.all([
    fetchArticles({ size: 5 }),
    fetchArticles({ type: 1, language: 'en', size: 1 }),
    fetchArticles({ type: 1, language: 'de', size: 1 }),
    fetchBooks({ type: 1 }),
  ])
  latest.value = latestRes.list || []
  enLatest.value = enRes.list?.[0] || null
  deLatest.value = deRes.list?.[0] || null

  const books = (booksRes || []).slice(0, 3)
  novels.value = (
    await Promise.all(
      books.map(async (book) => {
        const d = await fetchBook(book.id).catch(() => null)
        return { book, lastChapter: d?.chapters?.[d.chapters.length - 1] }
      }),
    )
  ).filter((n) => (n.book.chapterCount || 0) > 0)
})
</script>

<template>
  <div class="container home">
    <!-- ① 自我介绍 -->
    <section class="hero">
      <div class="hero-row">
        <div class="hero-text">
          <h1 class="hero-title">你好，我是<span class="hero-name">{{ site.get('nickname', '小天') }}</span>。</h1>
          <p class="hero-desc">
            {{ site.get('bio', '我在这里记录编程、英语与德语学习，也写一些故事和生活随笔。目前正在用 Vue 3 和 Go 持续完善这个博客。') }}
          </p>
        </div>
        <div class="hero-avatar">
          <img v-if="site.get('avatar')" :src="site.get('avatar')" alt="头像" class="avatar-img" />
          <span v-else class="avatar-placeholder">天</span>
        </div>
      </div>
      <nav class="hero-entries">
        <a href="#writing" class="hero-entry">最近写作 <span>↓</span></a>
        <router-link to="/learn" class="hero-entry">学习笔记 <span>→</span></router-link>
        <router-link to="/novel" class="hero-entry">小说与故事 <span>→</span></router-link>
      </nav>
    </section>

    <!-- ② 最近写作 -->
    <section id="writing" class="sec">
      <header class="sec-head">
        <h2 class="sec-title">最近写作</h2>
        <router-link to="/essays" class="sec-more">查看全部文章 →</router-link>
      </header>
      <div v-if="latest.length" class="post-list">
        <router-link
          v-for="a in latest"
          :key="a.id"
          :to="`/article/${a.id}`"
          class="post-row"
        >
          <div class="post-main">
            <span class="post-title">{{ a.title }}</span>
            <span v-if="a.summary" class="post-summary">{{ a.summary }}</span>
          </div>
          <div class="post-side">
            <span class="post-tag">{{ typeLabel(a) }}</span>
            <span class="post-date">{{ date(a.createdAt) }}</span>
          </div>
        </router-link>
      </div>
      <p v-else class="empty">还没有文章，去后台写第一篇吧。</p>
    </section>

    <!-- ③ 最近在学 -->
    <section id="learning" class="sec">
      <header class="sec-head">
        <h2 class="sec-title">最近在学</h2>
        <router-link to="/learn" class="sec-more">进入双语书房 →</router-link>
      </header>
      <div class="learn-grid">
        <router-link
          v-if="enLatest"
          :to="`/article/${enLatest.id}`"
          class="learn-card"
        >
          <div class="learn-lang">🇬🇧 英语学习</div>
          <div class="learn-body">
            <span class="learn-latest">{{ enLatest.title }}</span>
            <span class="learn-topic">最近在练习英文朗读</span>
          </div>
          <span class="learn-time">{{ date(enLatest.createdAt) }}</span>
        </router-link>
        <router-link
          v-if="deLatest"
          :to="`/article/${deLatest.id}`"
          class="learn-card"
        >
          <div class="learn-lang">🇩🇪 德语学习</div>
          <div class="learn-body">
            <span class="learn-latest">{{ deLatest.title }}</span>
            <span class="learn-topic">最近在学习德语基础阅读</span>
          </div>
          <span class="learn-time">{{ date(deLatest.createdAt) }}</span>
        </router-link>
      </div>
    </section>

    <!-- ④ 小说与故事 -->
    <section v-if="novels.length" id="novel" class="sec">
      <header class="sec-head">
        <h2 class="sec-title">小说与故事</h2>
        <router-link to="/novel" class="sec-more">全部作品 →</router-link>
      </header>
      <div class="novel-list">
        <router-link
          v-for="n in novels"
          :key="n.book.id"
          :to="`/novel/book/${n.book.id}`"
          class="novel-row"
        >
          <div class="novel-main">
            <span class="novel-title">
              {{ n.book.title }}
              <em class="novel-status">{{ (n.book.chapterCount || 0) >= 1 ? '连载中' : '' }}</em>
            </span>
            <span v-if="n.book.description" class="novel-desc">{{ n.book.description }}</span>
            <span v-if="n.lastChapter" class="novel-last">最新章节：{{ n.lastChapter.title }}</span>
          </div>
          <span class="novel-time">{{ date(n.book.updatedAt) }}</span>
        </router-link>
      </div>
    </section>

    <!-- ⑤ 最近在做 -->
    <section class="sec">
      <header class="sec-head">
        <h2 class="sec-title">最近在做</h2>
      </header>
      <div class="now-grid">
        <div class="now-item">
          <span class="now-label">正在开发</span>
          <span class="now-value">{{ site.get('now_dev', '小天 Goofy Blog · Vue 3 + Go') }}</span>
        </div>
        <div class="now-item">
          <span class="now-label">正在学习</span>
          <span class="now-value">{{ site.get('now_learning', '英语 · 德语') }}</span>
        </div>
        <div class="now-item">
          <span class="now-label">正在阅读</span>
          <span class="now-value">{{ site.get('now_reading', '《三体》 刘慈欣') }}</span>
        </div>
        <div class="now-item">
          <span class="now-label">最近更新</span>
          <span class="now-value">{{ site.get('now_update', '博客上线了多主题切换与小说专栏') }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home {
  max-width: 780px;
}

/* ===== 自我介绍 ===== */
.hero {
  padding: 56px 0 40px;
}

.hero-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 30px;
}

.hero-text {
  flex: 1;
  min-width: 0;
}

/* 头像位：圆形，未配置图片时显示主题色「天」字占位 */
.hero-avatar {
  flex-shrink: 0;
  width: 108px;
  height: 108px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
  background: var(--card);
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: grid;
  place-items: center;
  font-family: var(--font-heading);
  font-size: 46px;
  font-weight: 700;
  color: #fff;
  background: var(--accent);
}

.hero-title {
  font-family: var(--font-heading);
  font-size: 34px;
  font-weight: 700;
  color: var(--heading);
  margin: 0 0 14px;
  line-height: 1.4;
}

.hero-name {
  color: var(--accent);
}

.hero-desc {
  font-size: 16px;
  color: var(--text);
  line-height: 1.9;
  margin: 0 0 26px;
  max-width: 560px;
}

.hero-entries {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.hero-entry {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--heading);
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 8px 18px;
  transition: all 0.2s;
}

.hero-entry span {
  color: var(--muted);
  font-size: 12px;
}

.hero-entry:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.hero-entry:hover span {
  color: var(--accent);
}

/* ===== 通用区块 ===== */
.sec {
  padding: 30px 0;
  border-top: 1px solid var(--border);
}

.sec-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 18px;
}

.sec-title {
  font-family: var(--font-heading);
  font-size: 21px;
  font-weight: 700;
  color: var(--heading);
  margin: 0;
}

.sec-more {
  font-size: 13px;
  color: var(--muted);
}

.sec-more:hover {
  color: var(--accent);
}

.empty {
  color: var(--muted);
  font-size: 14px;
}

/* ===== 最近写作：分割线列表 ===== */
.post-list {
  display: flex;
  flex-direction: column;
}

.post-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 14px 2px;
  border-bottom: 1px solid var(--border);
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
  font-size: 15.5px;
  font-weight: 600;
  color: var(--heading);
}

.post-row:hover .post-title {
  color: var(--accent);
}

.post-summary {
  font-size: 13px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  flex-shrink: 0;
}

.post-tag {
  font-size: 11.5px;
  color: var(--muted);
  background: var(--surface);
  border-radius: 999px;
  padding: 1px 9px;
}

.post-date {
  font-size: 12px;
  color: var(--muted);
  font-family: var(--font-mono);
}

/* ===== 最近在学：两栏 ===== */
.learn-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.learn-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 18px 20px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  transition: border-color 0.2s;
}

.learn-card:hover {
  border-color: var(--accent);
}

.learn-lang {
  font-size: 14px;
  font-weight: 700;
  color: var(--heading);
}

.learn-body {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.learn-latest {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.learn-card:hover .learn-latest {
  color: var(--accent);
}

.learn-topic {
  font-size: 12.5px;
  color: var(--muted);
}

.learn-time {
  font-size: 12px;
  color: var(--muted);
  font-family: var(--font-mono);
}

/* ===== 小说与故事 ===== */
.novel-list {
  display: flex;
  flex-direction: column;
}

.novel-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 14px 2px;
  border-bottom: 1px solid var(--border);
}

.novel-row:last-child {
  border-bottom: none;
}

.novel-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.novel-title {
  font-size: 15.5px;
  font-weight: 600;
  color: var(--heading);
  display: flex;
  align-items: center;
  gap: 8px;
}

.novel-row:hover .novel-title {
  color: var(--accent);
}

.novel-status {
  font-style: normal;
  font-size: 11px;
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 999px;
  padding: 1px 8px;
}

.novel-desc {
  font-size: 13px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.novel-last {
  font-size: 12.5px;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.novel-time {
  font-size: 12px;
  color: var(--muted);
  font-family: var(--font-mono);
  flex-shrink: 0;
}

/* ===== 最近在做 ===== */
.now-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.now-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 16px 18px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
}

.now-label {
  font-size: 12px;
  color: var(--muted);
  letter-spacing: 1px;
}

.now-value {
  font-size: 14.5px;
  font-weight: 600;
  color: var(--heading);
}

@media (max-width: 640px) {
  .hero-row {
    flex-direction: column-reverse;
    align-items: flex-start;
    gap: 20px;
  }

  .hero-avatar {
    width: 84px;
    height: 84px;
  }

  .avatar-placeholder {
    font-size: 34px;
  }

  .hero-title {
    font-size: 27px;
  }

  .learn-grid,
  .now-grid {
    grid-template-columns: 1fr;
  }

  .post-side {
    display: none;
  }
}
</style>
