<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { motion } from 'motion-v'
import { fetchLinks } from '@/api'
import type { LinkItem } from '@/types'

const links = ref<LinkItem[]>([])

const hoverAnim = { scale: 1.02, y: -3 }
const spring = { type: 'spring', stiffness: 320, damping: 28 }

onMounted(async () => {
  links.value = (await fetchLinks()) || []
})
</script>

<template>
  <div class="container links-wrap">
    <div class="card links-head">
      <h1 class="page-title">友链</h1>
      <p class="page-sub">好博客，值得一个位置。欢迎交换友链 🤝</p>
    </div>

    <div class="links-grid" v-if="links.length">
      <motion.div
        v-for="link in links"
        :key="link.id"
        :while-hover="hoverAnim"
        :transition="spring"
      >
        <a
          :href="link.url"
          target="_blank"
          rel="noopener"
          class="link-card card"
        >
        <div class="link-avatar" :style="{ background: link.logo ? `url(${link.logo}) center/cover` : '' }">
          <span v-if="!link.logo">{{ link.name.slice(0, 1) }}</span>
        </div>
        <div class="link-info">
          <h3 class="link-name">{{ link.name }}</h3>
          <p class="link-desc">{{ link.description || link.url }}</p>
        </div>
        <span class="link-arrow">↗</span>
        </a>
      </motion.div>
    </div>
    <el-empty v-else description="还没有友链，去后台添加吧" />
  </div>
</template>

<style scoped>
.links-wrap {
  max-width: 900px;
}

.links-head {
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
  margin: 0;
  font-size: 14px;
}

.links-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.link-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 22px;
  height: 100%;
  transition: box-shadow 0.25s ease, border-color 0.2s ease;
}

.link-card:hover {
  box-shadow: var(--shadow-lift);
  border-color: var(--accent);
}

.link-avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: var(--accent-soft);
  color: var(--accent);
  display: grid;
  place-items: center;
  font-size: 22px;
  font-family: var(--font-heading);
  font-weight: 700;
  flex-shrink: 0;
}

.link-info {
  flex: 1;
  min-width: 0;
}

.link-name {
  margin: 0 0 2px;
  font-size: 16px;
}

.link-desc {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-arrow {
  color: var(--muted);
  transition: color 0.2s, transform 0.2s;
}

.link-card:hover .link-arrow {
  color: var(--accent);
  transform: translate(2px, -2px);
}

@media (max-width: 700px) {
  .links-grid {
    grid-template-columns: 1fr;
  }
}
</style>
