<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { fetchArticles, fetchCategories, fetchTags } from '@/api'
import type { Article, Category, Tag } from '@/types'
import ArticleCard from '@/components/ArticleCard.vue'
import Sidebar from '@/components/Sidebar.vue'

const route = useRoute()
const router = useRouter()

const list = ref<Article[]>([])
const total = ref(0)
const page = ref(1)
const size = 10
const loading = ref(false)

const activeCategory = ref<Category | null>(null)
const activeTag = ref<Tag | null>(null)

const filterTitle = computed(() => {
  const kw = route.query.keyword as string | undefined
  if (kw) return `搜索「${kw}」`
  if (activeCategory.value) return `分类：${activeCategory.value.name}`
  if (activeTag.value) return `标签：${activeTag.value.name}`
  return '全部随笔'
})

async function load() {
  loading.value = true
  try {
    const data = await fetchArticles({
      page: page.value,
      size,
      categoryId: route.name === 'category' ? (route.params.id as string) : '',
      tagId: route.name === 'tag' ? (route.params.id as string) : '',
      keyword: (route.query.keyword as string) || '',
    })
    list.value = data.list || []
    total.value = data.total
  } finally {
    loading.value = false
  }
}

// 高亮当前筛选项
async function loadFilterInfo() {
  activeCategory.value = null
  activeTag.value = null
  if (route.name === 'category') {
    const cats = await fetchCategories()
    activeCategory.value = cats.find((c) => c.id === Number(route.params.id)) || null
  } else if (route.name === 'tag') {
    const tags = await fetchTags()
    activeTag.value = tags.find((t) => t.id === Number(route.params.id)) || null
  }
}

function changePage(p: number) {
  page.value = p
  load()
}

function clearFilter() {
  if (route.query.keyword) router.push({ path: '/' })
  else if (route.name === 'category') router.push('/')
  else if (route.name === 'tag') router.push('/')
}

onMounted(() => {
  load()
  loadFilterInfo()
})

watch(
  () => route.fullPath,
  () => {
    page.value = 1
    load()
    loadFilterInfo()
  },
)
</script>

<template>
  <div class="container home-grid">
    <div class="home-main" v-loading="loading">
      <div v-if="filterTitle !== '最新文章'" class="filter-bar card">
        <span>{{ filterTitle }} · 共 {{ total }} 篇</span>
        <el-button link type="primary" size="small" @click="clearFilter">清除筛选 ✕</el-button>
      </div>

      <div class="article-list">
        <ArticleCard v-for="a in list" :key="a.id" :article="a" />
        <el-empty v-if="!loading && !list.length" description="没有找到相关文章" />
      </div>

      <div class="pagination-wrap" v-if="total > size">
        <el-pagination
          background
          layout="prev, pager, next"
          :total="total"
          :page-size="size"
          :current-page="page"
          @current-change="changePage"
        />
      </div>
    </div>

    <div class="home-aside">
      <Sidebar />
    </div>
  </div>
</template>

<style scoped>
.home-grid {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 26px;
  align-items: start;
}

.home-main {
  min-height: 300px;
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  margin-bottom: 18px;
  font-size: 14px;
  color: var(--muted);
}

.article-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  margin-top: 28px;
}

@media (max-width: 900px) {
  .home-grid {
    grid-template-columns: 1fr;
  }
}
</style>
