<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchDashboard } from '@/api'
import type { DashboardStats } from '@/types'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const stats = ref<DashboardStats | null>(null)

onMounted(async () => {
  stats.value = await fetchDashboard()
})

const cards = [
  { key: 'articleCount', label: '文章总数', icon: 'Document', color: '#c53620' },
  { key: 'publishedCount', label: '已发布', icon: 'CircleCheck', color: '#525e45' },
  { key: 'bookCount', label: '书房书籍', icon: 'Reading', color: '#564729' },
  { key: 'chapterCount', label: '章节总数', icon: 'Collection', color: '#857360' },
  { key: 'commentCount', label: '评论数', icon: 'ChatDotRound', color: '#a72e1b' },
  { key: 'viewTotal', label: '总浏览量', icon: 'View', color: '#6b5947' },
] as const
</script>

<template>
  <div>
    <div class="welcome card">
      <h2>{{ auth.user?.nickname || auth.user?.username }}，今天写点什么？</h2>
      <p>写博客、读原著、学外语 —— 每天进步一点点。</p>
      <router-link to="/admin/articles/new">
        <el-button type="primary">＋ 写新文章</el-button>
      </router-link>
    </div>

    <div class="stat-grid" v-if="stats">
      <div v-for="c in cards" :key="c.key" class="stat-card card">
        <div class="stat-icon" :style="{ background: c.color + '18', color: c.color }">
          <el-icon :size="22"><component :is="c.icon" /></el-icon>
        </div>
        <div class="stat-num">{{ stats[c.key] ?? 0 }}</div>
        <div class="stat-label">{{ c.label }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.welcome {
  padding: 28px 32px;
  margin-bottom: 18px;
  background: linear-gradient(120deg, var(--card), var(--accent-soft));
}

.welcome h2 {
  font-family: var(--font-heading);
  margin: 0 0 6px;
}

.welcome p {
  color: var(--muted);
  margin: 0 0 16px;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
  gap: 16px;
}

.stat-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stat-icon {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  margin-bottom: 4px;
}

.stat-num {
  font-size: 26px;
  font-weight: 800;
  font-family: var(--font-heading);
}

.stat-label {
  font-size: 13px;
  color: var(--muted);
}
</style>
