<script setup lang="ts">
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { fetchArchives } from '@/api'
import type { ArchiveGroup } from '@/types'

const groups = ref<ArchiveGroup[]>([])
const keyword = ref('')

onMounted(async () => {
  groups.value = (await fetchArchives()) || []
})

const filtered = () => {
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return groups.value
  return groups.value
    .map((g) => ({
      ...g,
      items: g.items.filter((it) => it.title.toLowerCase().includes(kw)),
    }))
    .filter((g) => g.items.length)
}
</script>

<template>
  <div class="container archive-wrap">
    <div class="card archive-head">
      <h1 class="page-title">归档</h1>
      <p class="page-sub">时间线里的每一篇文字</p>
      <el-input v-model="keyword" placeholder="在归档中搜索标题…" clearable class="archive-search" />
    </div>

    <div class="archive-body card" v-if="filtered().length">
      <section v-for="g in filtered()" :key="g.month" class="month-group">
        <h2 class="month-title">{{ g.month }} <em>{{ g.items.length }} 篇</em></h2>
        <ul class="month-list">
          <li v-for="it in g.items" :key="it.id" class="month-item">
            <span class="item-date">{{ dayjs(it.createdAt).format('MM-DD') }}</span>
            <router-link :to="`/article/${it.id}`" class="item-title">
              {{ it.title }}
              <em v-if="it.type === 1" class="item-badge">{{ it.language === 'de' ? 'DE' : 'EN' }}</em>
            </router-link>
          </li>
        </ul>
      </section>
    </div>
    <el-empty v-else description="没有匹配的文章" />
  </div>
</template>

<style scoped>
.archive-wrap {
  max-width: 860px;
}

.archive-head {
  padding: 30px 34px;
  margin-bottom: 20px;
}

.page-title {
  font-family: var(--font-heading);
  font-size: 28px;
  margin: 0 0 4px;
}

.page-sub {
  color: var(--muted);
  margin: 0 0 16px;
  font-size: 14px;
}

.archive-search {
  max-width: 320px;
}

.archive-body {
  padding: 26px 34px;
}

.month-group {
  position: relative;
  padding-left: 22px;
  border-left: 2px solid var(--border);
  margin-bottom: 26px;
}

.month-group:last-child {
  margin-bottom: 0;
}

.month-group::before {
  content: '';
  position: absolute;
  left: -6px;
  top: 6px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-soft);
}

.month-title {
  font-family: var(--font-heading);
  font-size: 19px;
  margin: 0 0 12px;
}

.month-title em {
  font-style: normal;
  font-size: 13px;
  color: var(--muted);
  font-weight: 400;
  margin-left: 6px;
}

.month-list {
  list-style: none;
  margin: 0;
  padding: 0;
}

.month-item {
  display: flex;
  align-items: baseline;
  gap: 14px;
  padding: 7px 2px;
}

.item-date {
  font-size: 13px;
  color: var(--muted);
  font-family: var(--font-mono);
  flex-shrink: 0;
}

.item-title {
  font-size: 15px;
  transition: color 0.15s;
}

.item-title:hover {
  color: var(--accent);
}

.item-badge {
  font-style: normal;
  font-size: 11px;
  color: var(--olive);
  border: 1px solid var(--badge-en-border);
  background: var(--badge-en-bg);
  border-radius: 999px;
  padding: 0 7px;
  margin-left: 8px;
  vertical-align: 2px;
}
</style>
