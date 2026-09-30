<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchPage } from '@/api'
import type { SinglePage } from '@/types'
import { renderMarkdown } from '@/utils/markdown'

const route = useRoute()
// /about 固定渲染 about 单页；/p/:slug 渲染任意单页
const slug = computed(() => (route.params.slug as string) || 'about')

const page = ref<SinglePage | null>(null)
const html = ref('')

async function load() {
  page.value = await fetchPage(slug.value)
  html.value = renderMarkdown(page.value?.content || '')
}

onMounted(load)
watch(slug, load)
</script>

<template>
  <div class="container about-wrap">
    <div class="card about-head">
      <h1 class="page-title">{{ page?.title || '关于' }}</h1>
    </div>
    <div class="card about-body md-content" v-html="html" />
  </div>
</template>

<style scoped>
.about-wrap {
  max-width: 820px;
}

.about-head {
  padding: 28px 34px;
  margin-bottom: 20px;
}

.page-title {
  font-family: var(--font-heading);
  font-size: 28px;
  margin: 0;
}

.about-body {
  padding: 32px 38px;
}
</style>
