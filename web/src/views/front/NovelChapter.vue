<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchBook, fetchChapter } from '@/api'
import type { ChapterDetailData, ChapterItem } from '@/types'
import { pushReadingRecord } from '@/utils/history'
import { dropLeadingTitle, splitParagraphs } from '@/utils/reading'
import { speakTexts, warmUpVoices, type SpeakHandle } from '@/utils/tts'

const route = useRoute()
const router = useRouter()
const data = ref<ChapterDetailData | null>(null)

const chapter = computed(() => data.value?.chapter)
const book = computed(() => data.value?.book)
const paragraphs = computed(() =>
  dropLeadingTitle(splitParagraphs(chapter.value?.content || ''), chapter.value?.title || ''),
)

// 目录抽屉
const catalog = reactive({ visible: false, chapters: [] as ChapterItem[] })

async function openCatalog() {
  if (!book.value) return
  if (!catalog.chapters.length) {
    const d = await fetchBook(book.value.id)
    catalog.chapters = d.chapters || []
  }
  catalog.visible = true
}

function catalogJump(id: number) {
  catalog.visible = false
  if (id !== chapter.value?.id) {
    router.push(`/novel/chapter/${id}`)
  }
}

// 字号
const fontSize = ref(18)
function changeFontSize(delta: number) {
  fontSize.value = Math.min(24, Math.max(14, fontSize.value + delta))
}

// 听书
const tts = reactive({ playing: false, paused: false })
let ttsHandle: SpeakHandle | null = null

function stopTts() {
  ttsHandle?.stop()
  ttsHandle = null
  tts.playing = false
  tts.paused = false
}

function togglePlayPause() {
  if (!paragraphs.value.length) return
  if (tts.playing && ttsHandle) {
    if (tts.paused) {
      ttsHandle.resume()
      tts.paused = false
    } else {
      ttsHandle.pause()
      tts.paused = true
    }
    return
  }
  if (!('speechSynthesis' in window)) {
    ElMessage.warning('当前浏览器不支持语音朗读')
    return
  }
  tts.playing = true
  ttsHandle = speakTexts(paragraphs.value, {
    lang: book.value?.language === 'en' ? 'en' : book.value?.language === 'de' ? 'de' : 'zh',
    rate: 1,
    onDone: () => {
      tts.playing = false
      tts.paused = false
    },
  })
}

onMounted(async () => {
  warmUpVoices()
  await loadChapter()
})

// 栏目内切章（上一章/下一章/目录跳转）时组件不重挂载，手动重载
watch(
  () => route.params.id,
  (id, old) => {
    if (id && id !== old) {
      stopTts()
      loadChapter()
    }
  },
)

async function loadChapter() {
  data.value = await fetchChapter(route.params.id as string)
  if (data.value) {
    pushReadingRecord({
      chapterId: data.value.chapter.id,
      chapterTitle: data.value.chapter.title,
      bookId: data.value.book.id,
      bookTitle: data.value.book.title,
      language: data.value.book.language,
      kind: 'novel',
    })
  }
}

onUnmounted(stopTts)
</script>

<template>
  <div class="container chapter-wrap" v-if="chapter && book">
    <!-- 顶部工具条 -->
    <div class="chapter-toolbar card">
      <router-link :to="`/novel/book/${book.id}`" class="back-link">← {{ book.title }}</router-link>
      <div class="tools">
        <button class="tool-btn" @click="openCatalog">☰ 目录</button>
        <div class="font-ctrl">
          <button class="tool-btn" @click="changeFontSize(-1)">A−</button>
          <button class="tool-btn" @click="changeFontSize(1)">A＋</button>
        </div>
        <button class="tool-btn" @click="togglePlayPause">
          <template v-if="!tts.playing">🎧 听书</template>
          <template v-else>{{ tts.paused ? '▶ 继续' : '⏸ 暂停' }}</template>
        </button>
        <button v-if="tts.playing" class="tool-btn" @click="stopTts">⏹ 停止</button>
      </div>
    </div>

    <!-- 正文（沉浸阅读） -->
    <article class="reader card">
      <header class="chapter-head">
        <h1 class="chapter-title">{{ chapter.title }}</h1>
      </header>

      <div
        v-for="(p, i) in paragraphs"
        :key="i"
        class="para"
        :style="{ fontSize: fontSize + 'px' }"
      >{{ p }}</div>

      <div v-if="!paragraphs.length" class="empty-chapter">本章内容暂未录入</div>
    </article>

    <!-- 上一章 / 下一章 -->
    <nav class="chapter-nav">
      <router-link v-if="data?.prev" :to="`/novel/chapter/${data.prev.id}`" class="nav-btn card">
        <span class="nav-label">← 上一章</span>
        <span class="nav-title">{{ data.prev.title }}</span>
      </router-link>
      <span v-else class="nav-btn card nav-none">已是第一章</span>
      <router-link v-if="data?.next" :to="`/novel/chapter/${data.next.id}`" class="nav-btn card next">
        <span class="nav-label">下一章 →</span>
        <span class="nav-title">{{ data.next.title }}</span>
      </router-link>
      <span v-else class="nav-btn card nav-none nav-next">已是最后一章 🎉</span>
    </nav>

    <!-- 目录抽屉 -->
    <el-drawer v-model="catalog.visible" title="章节目录" size="320px">
      <div class="catalog-list">
        <a
          v-for="(c, i) in catalog.chapters"
          :key="c.id"
          class="catalog-item"
          :class="{ current: c.id === chapter.id }"
          @click="catalogJump(c.id)"
        >
          <span class="catalog-no">{{ String(i + 1).padStart(2, '0') }}</span>
          <span class="catalog-title">{{ c.title }}</span>
        </a>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.chapter-wrap {
  max-width: 800px;
}

.chapter-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  margin-bottom: 18px;
  gap: 10px;
  flex-wrap: wrap;
}

.back-link {
  font-size: 14px;
  color: var(--muted);
}

.back-link:hover {
  color: var(--accent);
}

.tools {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.font-ctrl {
  display: flex;
  gap: 6px;
}

.tool-btn {
  border: 1px solid var(--border);
  background: var(--card);
  border-radius: 7px;
  font-size: 12px;
  padding: 3px 10px;
  cursor: pointer;
  color: var(--muted);
  transition: all 0.15s;
}

.tool-btn:hover {
  color: var(--accent);
  border-color: var(--accent);
}

.reader {
  padding: 40px 52px;
}

.chapter-head {
  text-align: center;
  margin: 0 0 26px;
}

.chapter-title {
  font-family: var(--font-heading);
  font-size: 26px;
  font-weight: 700;
  color: var(--heading);
  margin: 0;
  line-height: 1.5;
}

.para {
  margin: 0 0 1em;
  line-height: 2.05;
  text-align: justify;
  text-indent: 2em;
  color: var(--text);
}

.empty-chapter {
  text-align: center;
  color: var(--muted);
  padding: 60px 0;
  font-size: 15px;
  text-indent: 0;
}

.chapter-nav {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-top: 20px;
}

.nav-btn {
  padding: 14px 18px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  transition: border-color 0.2s;
  text-decoration: none;
  color: inherit;
}

.nav-btn:hover {
  border-color: var(--accent);
}

.nav-btn.next,
.nav-none.nav-next {
  text-align: right;
}

.nav-none {
  color: var(--muted);
  font-size: 13px;
  justify-content: center;
}

.nav-label {
  font-size: 12px;
  color: var(--muted);
}

.nav-title {
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 目录抽屉 */
.catalog-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.catalog-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
}

.catalog-item:hover {
  background: var(--accent-soft);
}

.catalog-item.current {
  background: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

.catalog-no {
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--accent);
  font-weight: 700;
}

.catalog-title {
  flex: 1;
  font-size: 14px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 700px) {
  .reader {
    padding: 24px 18px;
  }

  .para {
    text-indent: 0;
  }
}
</style>
