<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import { fetchArticle } from '@/api'
import type { ArticleDetailData } from '@/types'
import { extractToc, renderMarkdown } from '@/utils/markdown'
import CommentSection from '@/components/CommentSection.vue'

const route = useRoute()
const data = ref<ArticleDetailData | null>(null)
const html = ref('')
const toc = ref<{ id: string; text: string; level: number }[]>([])

const article = computed(() => data.value?.article)

onMounted(async () => {
  data.value = await fetchArticle(route.params.id as string)
  if (data.value) {
    html.value = renderMarkdown(data.value.article.content || '')
    toc.value = extractToc(data.value.article.content || '')
  }
})
</script>

<template>
  <div class="container" v-if="article">
    <div class="detail-grid">
      <article class="detail-main">
        <header class="detail-head card">
          <div class="detail-meta">
            <span>{{ dayjs(article.createdAt).format('YYYY年MM月DD日') }}</span>
            <span v-if="article.category" class="meta-cat">{{ article.category.name }}</span>
            <span>{{ article.views }} 次阅读</span>
            <span v-if="article.status === 0" class="meta-draft">草稿预览</span>
          </div>
          <h1 class="detail-title">{{ article.title }}</h1>
          <div class="detail-tags" v-if="article.tags?.length">
            <router-link v-for="t in article.tags" :key="t.id" :to="`/tag/${t.id}`" class="detail-tag">
              # {{ t.name }}
            </router-link>
          </div>
        </header>

        <div class="detail-body card md-content" v-html="html" />

        <nav class="neighbors">
          <router-link v-if="data?.prev" :to="`/article/${data.prev.id}`" class="neighbor card">
            <span class="neighbor-label">← 上一篇</span>
            <span class="neighbor-title">{{ data.prev.title }}</span>
          </router-link>
          <span v-else class="neighbor card neighbor-none">没有更早的了</span>
          <router-link v-if="data?.next" :to="`/article/${data.next.id}`" class="neighbor card next">
            <span class="neighbor-label">下一篇 →</span>
            <span class="neighbor-title">{{ data.next.title }}</span>
          </router-link>
          <span v-else class="neighbor card neighbor-none">没有更新的了</span>
        </nav>

        <CommentSection :article-id="article.id" />
      </article>

      <aside class="detail-toc" v-if="toc.length">
        <div class="toc-card card">
          <h4 class="toc-title">目录</h4>
          <a
            v-for="t in toc"
            :key="t.id"
            :href="`#${t.id}`"
            class="toc-item"
            :style="{ paddingLeft: (t.level - 1) * 12 + 10 + 'px' }"
          >
            {{ t.text }}
          </a>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 240px;
  gap: 24px;
  align-items: start;
}

.detail-main {
  min-width: 0;
}

.detail-head {
  padding: 28px 32px;
}

.detail-meta {
  display: flex;
  gap: 14px;
  font-size: 13px;
  color: var(--muted);
}

.meta-cat {
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 999px;
  padding: 0 10px;
}

.meta-draft {
  color: #b45309;
  background: #fef3c7;
  border-radius: 999px;
  padding: 0 10px;
}

.detail-title {
  font-family: var(--font-heading);
  font-size: 30px;
  line-height: 1.4;
  margin: 14px 0 12px;
}

.detail-tags {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.detail-tag {
  font-size: 13px;
  color: var(--muted);
}

.detail-tag:hover {
  color: var(--accent);
}

.detail-body {
  margin-top: 18px;
  padding: 32px 36px;
}

.neighbors {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-top: 22px;
}

.neighbor {
  padding: 14px 18px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  transition: all 0.2s;
}

.neighbor:hover {
  border-color: var(--accent);
}

.neighbor.next {
  text-align: right;
}

.neighbor-none {
  color: var(--muted);
  font-size: 13px;
  justify-content: center;
}

.neighbor-label {
  font-size: 12px;
  color: var(--muted);
}

.neighbor-title {
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-toc {
  position: sticky;
  top: 86px;
}

.toc-card {
  padding: 16px 8px;
  max-height: calc(100vh - 120px);
  overflow-y: auto;
}

.toc-title {
  margin: 0 0 8px 10px;
  font-size: 13px;
  color: var(--muted);
  letter-spacing: 2px;
}

.toc-item {
  display: block;
  font-size: 13px;
  color: var(--muted);
  padding: 5px 10px;
  border-radius: 6px;
  border-left: 2px solid transparent;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.toc-item:hover {
  color: var(--accent);
  background: var(--accent-soft);
}

@media (max-width: 1000px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .detail-toc {
    display: none;
  }
}
</style>
