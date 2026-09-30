<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { fetchCategories, fetchTags } from '@/api'
import type { Category, Tag } from '@/types'

const route = useRoute()
const categories = ref<Category[]>([])
const tags = ref<Tag[]>([])

onMounted(async () => {
  const [cats, tgs] = await Promise.all([fetchCategories(), fetchTags()])
  categories.value = cats || []
  tags.value = (tgs || []).sort((a, b) => (b.articleCount || 0) - (a.articleCount || 0))
})
</script>

<template>
  <aside class="sidebar">
    <div class="card side-block">
      <h3 class="side-title">分类</h3>
      <div class="side-list">
        <router-link
          v-for="cat in categories"
          :key="cat.id"
          :to="`/category/${cat.id}`"
          class="side-row"
          :class="{ active: route.name === 'category' && route.params.id === String(cat.id) }"
        >
          <span>{{ cat.name }}</span>
          <span class="side-count">{{ cat.articleCount || 0 }}</span>
        </router-link>
      </div>
    </div>

    <div class="card side-block">
      <h3 class="side-title">标签</h3>
      <div class="tag-cloud">
        <router-link
          v-for="tag in tags"
          :key="tag.id"
          :to="`/tag/${tag.id}`"
          class="tag-chip"
          :class="{ active: route.name === 'tag' && route.params.id === String(tag.id) }"
        >
          {{ tag.name }}
          <em>{{ tag.articleCount || 0 }}</em>
        </router-link>
        <p v-if="!tags.length" class="side-empty">还没有标签</p>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.side-block {
  padding: 18px 20px;
}

.side-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 700;
  color: var(--muted);
  letter-spacing: 2px;
  text-transform: uppercase;
  display: flex;
  align-items: center;
  gap: 8px;
}

.side-title::before {
  content: '';
  width: 4px;
  height: 14px;
  background: var(--accent);
  border-radius: 2px;
}

.side-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.side-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 7px 10px;
  border-radius: 8px;
  font-size: 14px;
  color: var(--text);
  transition: background 0.15s;
}

.side-row:hover {
  background: var(--accent-soft);
  color: var(--accent);
}

.side-row.active {
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

.side-count {
  font-size: 12px;
  color: var(--muted);
  background: var(--surface);
  border-radius: 999px;
  padding: 0 8px;
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-chip {
  font-size: 13px;
  padding: 3px 11px;
  border: 1px solid var(--border);
  border-radius: 999px;
  color: var(--muted);
  transition: all 0.15s;
}

.tag-chip em {
  font-style: normal;
  font-size: 11px;
  opacity: 0.7;
  margin-left: 2px;
}

.tag-chip:hover,
.tag-chip.active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-soft);
}

.side-empty {
  color: var(--muted);
  font-size: 13px;
  margin: 4px 0;
}
</style>
