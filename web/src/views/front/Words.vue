<script setup lang="ts">
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { deleteWord, fetchWords } from '@/api'
import type { WordItem } from '@/types'
import { useAuthStore } from '@/stores/auth'
import { speakTexts } from '@/utils/tts'

const auth = useAuthStore()
const words = ref<WordItem[]>([])
const loading = ref(false)
const lang = ref('')
const keyword = ref('')

async function load() {
  loading.value = true
  try {
    words.value =
      (await fetchWords({
        language: lang.value || undefined,
        keyword: keyword.value.trim() || undefined,
      })) || []
  } finally {
    loading.value = false
  }
}

function switchLang(l: string) {
  lang.value = lang.value === l ? '' : l
  load()
}

async function remove(w: WordItem) {
  await deleteWord(w.id)
  ElMessage.success('已删除')
  load()
}

/** 点击单词朗读 */
function pronounce(w: WordItem) {
  if (!('speechSynthesis' in window)) return
  speakTexts([w.word], { lang: w.language, rate: 1 })
}

onMounted(load)
</script>

<template>
  <div class="container words-wrap">
    <div class="words-head card">
      <router-link to="/learn" class="back-link">← 双语书房</router-link>
      <div class="head-row">
        <h1 class="page-title">📚 生词本 <em>{{ words.length }}</em></h1>
        <p class="page-sub">阅读时悬停单词点「＋收藏」自动攒进这里；点单词可发音。</p>
      </div>
      <div class="head-tools">
        <div class="lang-tabs">
          <button class="lang-tab" :class="{ active: lang === '' }" @click="switchLang('')">全部</button>
          <button class="lang-tab" :class="{ active: lang === 'en' }" @click="switchLang('en')">🇬🇧 英语</button>
          <button class="lang-tab" :class="{ active: lang === 'de' }" @click="switchLang('de')">🇩🇪 德语</button>
        </div>
        <el-input
          v-model="keyword"
          placeholder="搜索单词或释义…"
          :prefix-icon="Search"
          clearable
          class="search-input"
          @keyup.enter="load"
          @clear="load"
        />
        <router-link
          v-if="words.length"
          :to="`/learn/words/review${lang ? `?lang=${lang}` : ''}`"
          class="review-entry"
        >
          🎯 复习模式（{{ words.length }} 词）
        </router-link>
      </div>
    </div>

    <div class="word-grid" v-loading="loading">
      <div v-for="w in words" :key="w.id" class="word-card card">
        <div class="word-main">
          <span class="word-text" :title="`点击发音`" @click="pronounce(w)">🔊 {{ w.word }}</span>
          <span class="word-lang" :class="w.language">{{ w.language === 'de' ? 'DE' : 'EN' }}</span>
        </div>
        <p class="word-trans">{{ w.translation || '—' }}</p>
        <p v-if="w.sourceText" class="word-source">“{{ w.sourceText.slice(0, 60) }}{{ w.sourceText.length > 60 ? '…' : '' }}”</p>
        <div class="word-foot">
          <span class="word-time">{{ dayjs(w.createdAt).format('YYYY-MM-DD') }}</span>
          <el-popconfirm v-if="auth.isLoggedIn" :title="`删除「${w.word}」？`" @confirm="remove(w)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>
    </div>
    <el-empty v-if="!loading && !words.length" description="生词本还是空的，去阅读时收藏几个单词吧" />
  </div>
</template>

<style scoped>
.words-wrap {
  max-width: 1020px;
}

.words-head {
  padding: 26px 32px;
  margin-bottom: 20px;
}

.back-link {
  font-size: 13px;
  color: var(--muted);
}

.back-link:hover {
  color: var(--accent);
}

.head-row {
  margin-top: 10px;
}

.page-title {
  font-family: var(--font-heading);
  font-size: 28px;
  margin: 0 0 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-title em {
  font-style: normal;
  font-size: 14px;
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 999px;
  padding: 1px 10px;
}

.page-sub {
  color: var(--muted);
  margin: 0 0 16px;
  font-size: 14px;
}

.head-tools {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.lang-tabs {
  display: inline-flex;
  gap: 6px;
  background: var(--surface);
  padding: 4px;
  border-radius: 999px;
}

.lang-tab {
  border: none;
  background: transparent;
  padding: 6px 16px;
  border-radius: 999px;
  font-size: 13.5px;
  color: var(--muted);
  cursor: pointer;
  transition: all 0.2s;
}

.lang-tab.active {
  background: var(--accent);
  color: #fff;
  font-weight: 600;
}

.search-input {
  width: 240px;
}

.review-entry {
  font-size: 13.5px;
  font-weight: 600;
  color: #fff;
  background: var(--accent);
  border-radius: 999px;
  padding: 7px 18px;
  transition: background 0.2s;
  white-space: nowrap;
}

.review-entry:hover {
  background: var(--accent-hover);
}

.word-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  min-height: 150px;
}

.word-card {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.word-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.word-text {
  font-family: var(--font-heading);
  font-size: 19px;
  font-weight: 700;
  color: var(--heading);
  cursor: pointer;
}

.word-text:hover {
  color: var(--accent);
}

.word-lang {
  font-size: 11px;
  font-family: var(--font-mono);
  font-weight: 700;
  border-radius: 6px;
  padding: 1px 8px;
}

.word-lang.en {
  color: var(--olive);
  background: var(--badge-en-bg);
}

.word-lang.de {
  color: var(--accent-hover);
  background: var(--accent-soft);
}

.word-trans {
  margin: 0;
  font-size: 14px;
  color: var(--text);
}

.word-source {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
  font-style: italic;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.word-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
  padding-top: 6px;
}

.word-time {
  font-size: 11px;
  color: var(--muted);
}

@media (max-width: 860px) {
  .word-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .word-grid {
    grid-template-columns: 1fr;
  }
}
</style>
