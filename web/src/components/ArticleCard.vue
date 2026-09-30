<script setup lang="ts">
import dayjs from 'dayjs'
import { motion } from 'motion-v'
import type { Article } from '@/types'

const props = defineProps<{ article: Article }>()

const date = (v: string) => dayjs(v).format('YYYY-MM-DD')

const hoverAnim = { scale: 1.015, y: -4 }
const spring = { type: 'spring', stiffness: 320, damping: 28 }
</script>

<template>
  <motion.div :while-hover="hoverAnim" :transition="spring">
    <article class="article-card card">
    <router-link :to="`/article/${article.id}`" class="card-link">
      <div class="card-body">
        <div class="card-meta">
          <span class="meta-date">{{ date(article.createdAt) }}</span>
          <span v-if="article.category" class="meta-cat">{{ article.category.name }}</span>
          <span v-if="article.type === 1" class="meta-learn">学习 · {{ article.language === 'de' ? 'DE' : 'EN' }}</span>
        </div>
        <h2 class="card-title">{{ article.title }}</h2>
        <p v-if="article.summary" class="card-summary">{{ article.summary }}</p>
        <div class="card-foot">
          <div class="card-tags">
            <span v-for="tag in article.tags || []" :key="tag.id" class="card-tag"># {{ tag.name }}</span>
          </div>
          <span class="card-views">{{ article.views }} 次阅读</span>
        </div>
      </div>
    </router-link>
    </article>
  </motion.div>
</template>

<style scoped>
.article-card {
  transition: box-shadow 0.25s ease;
  overflow: hidden;
}

.article-card:hover {
  box-shadow: var(--shadow-lift);
}

.card-link {
  display: block;
  padding: 22px 26px;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--muted);
}

.meta-cat {
  color: var(--accent);
  background: var(--accent-soft);
  padding: 1px 10px;
  border-radius: 999px;
  font-size: 12px;
}

.meta-learn {
  color: var(--olive);
  background: var(--badge-en-bg);
  padding: 1px 10px;
  border-radius: 999px;
  font-size: 12px;
  border: 1px solid var(--badge-en-border);
}

.card-title {
  font-family: var(--font-heading);
  font-size: 21px;
  margin: 10px 0 8px;
  line-height: 1.45;
}

.article-card:hover .card-title {
  color: var(--accent);
}

.card-summary {
  color: var(--muted);
  font-size: 14px;
  margin: 0 0 12px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-tags {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.card-tag {
  font-size: 12px;
  color: var(--muted);
}

.card-views {
  font-size: 12px;
  color: var(--muted);
  white-space: nowrap;
}
</style>
