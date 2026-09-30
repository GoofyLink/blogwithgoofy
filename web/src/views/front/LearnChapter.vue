<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import {
  addWord, createChapterNote, deleteChapterNote, fetchChapter, fetchChapterNotes,
  translateText, updateChapterNote,
} from '@/api'
import type { ChapterDetailData, ChapterNote } from '@/types'
import { useAuthStore } from '@/stores/auth'
import { speakTexts, warmUpVoices, type SpeakHandle } from '@/utils/tts'
import { pushReadingRecord } from '@/utils/history'
import {
  dropLeadingTitle, normalizeWord, sentKey, splitParagraphs, splitSentences, tokenizeWords,
} from '@/utils/reading'

const route = useRoute()
const auth = useAuthStore()

const data = ref<ChapterDetailData | null>(null)

const chapter = computed(() => data.value?.chapter)
const book = computed(() => data.value?.book)
const fromLang = computed(() => book.value?.language || 'en')

// 阅读笔记（段落高亮依赖它，需先声明）
const notes = ref<ChapterNote[]>([])

// 段落 -> 句子 -> 词 三级切分；词带绝对偏移，用于笔记摘录的精准高亮
interface TokenView {
  text: string
  start: number
  end: number
}
interface SentenceView {
  key: string
  raw: string
  tokens: TokenView[]
}
interface ParaView {
  sentences: SentenceView[]
  highlights: { start: number; end: number }[]
}

const paragraphs = computed<ParaView[]>(() => {
  const paras = dropLeadingTitle(
    splitParagraphs(chapter.value?.content || ''),
    chapter.value?.title || '',
  )
  return paras.map((p, pi) => {
    // 本段各条笔记的摘录在段落中的文字范围
    const highlights: { start: number; end: number }[] = []
    const lowerP = p.toLowerCase()
    for (const n of notes.value) {
      if (n.paragraphIndex !== pi) continue
      const q = n.quote.trim()
      if (!q) continue
      let idx = p.indexOf(q)
      if (idx === -1) idx = lowerP.indexOf(q.toLowerCase())
      if (idx !== -1) highlights.push({ start: idx, end: idx + q.length })
    }

    // 句子切分并记录每个词在段落中的绝对偏移
    const sentences: SentenceView[] = []
    let cursor = 0
    splitSentences(p).forEach((s, si) => {
      const found = p.indexOf(s, cursor)
      const sentStart = found === -1 ? cursor : found
      if (found !== -1) cursor = found + s.length
      let off = sentStart
      const tokens: TokenView[] = tokenizeWords(s).map((tk) => {
        const t = { text: tk, start: off, end: off + tk.length }
        off += tk.length
        return t
      })
      sentences.push({ key: sentKey(pi, si), raw: s, tokens })
    })
    return { sentences, highlights }
  })
})

/** 该词是否落在某条笔记的摘录范围内 */
function isHighlighted(para: ParaView, token: TokenView): boolean {
  return para.highlights.some((h) => token.start < h.end && token.end > h.start)
}

const paraTranslations = computed(() =>
  dropLeadingTitle(
    splitParagraphs(chapter.value?.translation || ''),
    chapter.value?.title || '',
  ),
)

const hasTranslation = computed(() => paraTranslations.value.length > 0)

// ---------- 字号 ----------
const fontSize = ref(18)
function changeFontSize(delta: number) {
  fontSize.value = Math.min(24, Math.max(14, fontSize.value + delta))
}

// ---------- 中文对照（每段独立折叠 + 一键展开全部） ----------
const paraTransOpen = reactive<Record<number, boolean>>({})

const allTransOpen = computed(() =>
  paragraphs.value.length > 0 &&
  paragraphs.value.every((_, pi) => !paraTranslations.value[pi] || paraTransOpen[pi]),
)

function toggleAllTrans() {
  const open = !allTransOpen.value
  paragraphs.value.forEach((_, pi) => {
    if (paraTranslations.value[pi]) paraTransOpen[pi] = open
  })
}

function toggleParaTrans(pi: number) {
  paraTransOpen[pi] = !paraTransOpen[pi]
}

// ---------- 句子点击翻译 ----------
const sentTrans = reactive<Record<string, string>>({}) // key -> 译文（"" 表示请求中）
const sentOpen = reactive<Record<string, boolean>>({})

const sentenceCache = new Map<string, string>()

function toggleSentence(view: SentenceView) {
  const k = view.key
  if (sentOpen[k]) {
    sentOpen[k] = false
    return
  }
  sentOpen[k] = true
  if (!(k in sentTrans)) {
    sentTrans[k] = ''
    const cached = sentenceCache.get(view.raw)
    if (cached) {
      sentTrans[k] = cached
      return
    }
    translateText(view.raw, fromLang.value)
      .then((r) => {
        sentenceCache.set(view.raw, r.text)
        sentTrans[k] = r.text
      })
      .catch(() => {
        sentOpen[k] = false
        delete sentTrans[k]
      })
  }
}

// ---------- 单词悬停翻译 ----------
const tip = reactive({ show: false, x: 0, y: 0, word: '', trans: '', loading: false, sentence: '' })
const wordCache = new Map<string, string>()
let hoverTimer: ReturnType<typeof setTimeout> | null = null
let tipHideTimer: ReturnType<typeof setTimeout> | null = null

function hideTip() {
  tip.show = false
}

function hoverWord(e: MouseEvent, token: string, sentence = '') {
  const word = normalizeWord(token)
  if (!word) return
  if (hoverTimer) clearTimeout(hoverTimer)
  if (tipHideTimer) clearTimeout(tipHideTimer)
  hoverTimer = setTimeout(() => {
    tip.word = word
    tip.sentence = sentence
    // 直接算好浮层的固定定位坐标，模板无需访问 window
    tip.x = Math.min(e.clientX + 14, window.innerWidth - 240)
    tip.y = e.clientY > 90 ? e.clientY - 54 : e.clientY + 22
    tip.show = true
    const cached = wordCache.get(word.toLowerCase())
    if (cached) {
      tip.trans = cached
      tip.loading = false
      return
    }
    tip.trans = ''
    tip.loading = true
    translateText(word, fromLang.value)
      .then((r) => {
        wordCache.set(word.toLowerCase(), r.text)
        // 鼠标已移到别的词上则不更新
        if (tip.show && tip.word === word) {
          tip.trans = r.text
          tip.loading = false
        }
      })
      .catch(() => {
        if (tip.show && tip.word === word) {
          tip.trans = '查询失败'
          tip.loading = false
        }
      })
  }, 150)
}

function leaveWord() {
  if (hoverTimer) clearTimeout(hoverTimer)
  // 延迟隐藏，给鼠标移入浮层点「收藏」留时间
  if (tipHideTimer) clearTimeout(tipHideTimer)
  tipHideTimer = setTimeout(hideTip, 300)
}

function keepTip() {
  if (tipHideTimer) clearTimeout(tipHideTimer)
}

/** 悬停浮层里收藏当前单词到生词本 */
async function saveTipWord() {
  if (!tip.word || !chapter.value) return
  try {
    await addWord({
      word: tip.word,
      language: fromLang.value,
      translation: tip.loading ? '' : tip.trans,
      sourceText: tip.sentence.slice(0, 480),
    })
    ElMessage.success(`「${tip.word}」已加入生词本`)
    hideTip()
  } catch {
    /* 拦截器已提示 */
  }
}

// ==================== 阅读笔记 ====================
const activePara = ref(-1)

const articleRef = ref<HTMLElement | null>(null)
const noteFloat = reactive({ show: false, top: 0, left: 0 })
let pendingParaIndex = 0
let pendingQuote = ''

function onArticleMouseup() {
  if (!auth.isLoggedIn) return
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || !sel.rangeCount) {
    noteFloat.show = false
    return
  }
  const text = sel.toString().trim()
  if (!text || text.length > 800 || !articleRef.value) {
    noteFloat.show = false
    return
  }
  const node = sel.anchorNode
  const el = node?.nodeType === 3 ? node.parentElement : (node as HTMLElement)
  const para = el?.closest('.para') as HTMLElement | null
  if (!para || !articleRef.value.contains(para)) {
    noteFloat.show = false
    return
  }
  pendingParaIndex = Number(para.dataset.idx)
  pendingQuote = text

  const rect = sel.getRangeAt(0).getBoundingClientRect()
  const containerRect = articleRef.value.getBoundingClientRect()
  noteFloat.top = rect.top - containerRect.top - 38
  noteFloat.left = Math.max(0, rect.left - containerRect.left + rect.width / 2 - 50)
  noteFloat.show = true
}

const noteDialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  quote: '',
  note: '',
  paraIndex: 0,
})
const noteSaving = ref(false)

function openNoteCreate() {
  noteDialog.mode = 'create'
  noteDialog.quote = pendingQuote
  noteDialog.note = ''
  noteDialog.paraIndex = pendingParaIndex
  noteDialog.visible = true
  noteFloat.show = false
}

function openNoteCreateForPara(idx: number) {
  noteDialog.mode = 'create'
  noteDialog.quote = ''
  noteDialog.note = ''
  noteDialog.paraIndex = idx
  noteDialog.visible = true
}

function openNoteEdit(n: ChapterNote) {
  noteDialog.mode = 'edit'
  noteDialog.editId = n.id
  noteDialog.quote = n.quote
  noteDialog.note = n.note
  noteDialog.paraIndex = n.paragraphIndex
  noteDialog.visible = true
}

async function loadNotes() {
  if (!chapter.value) return
  notes.value = (await fetchChapterNotes(chapter.value.id)) || []
}

async function saveNote() {
  if (!noteDialog.note.trim()) {
    ElMessage.warning('请填写笔记内容')
    return
  }
  noteSaving.value = true
  try {
    if (noteDialog.mode === 'create') {
      await createChapterNote({
        chapterId: chapter.value!.id,
        paragraphIndex: noteDialog.paraIndex,
        quote: noteDialog.quote.trim(),
        note: noteDialog.note.trim(),
      })
      ElMessage.success('笔记已保存')
    } else {
      await updateChapterNote(noteDialog.editId, {
        paragraphIndex: noteDialog.paraIndex,
        quote: noteDialog.quote.trim(),
        note: noteDialog.note.trim(),
      })
      ElMessage.success('笔记已更新')
    }
    noteDialog.visible = false
    await loadNotes()
  } finally {
    noteSaving.value = false
  }
}

async function removeNote(n: ChapterNote) {
  await deleteChapterNote(n.id)
  ElMessage.success('已删除')
  await loadNotes()
}

function scrollToPara(idx: number) {
  activePara.value = idx
  const el = articleRef.value?.querySelector(`.para[data-idx="${idx}"]`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

// ==================== 朗读（浏览器原生 TTS） ====================
const tts = reactive({ playing: false, paused: false, rate: 1 })
const speakingKey = ref('')
let ttsHandle: SpeakHandle | null = null
let ttsRateRef = 1 // 供 speakTexts 每句读取，实现中途改语速下一句生效

function stopTts() {
  ttsHandle?.stop()
  ttsHandle = null
  tts.playing = false
  tts.paused = false
  speakingKey.value = ''
}

/** 展平（或截取）要朗读的句子，返回 { keys, raws } */
function collectSentences(fromPara?: number) {
  const keys: string[] = []
  const raws: string[] = []
  paragraphs.value.forEach((para, pi) => {
    if (fromPara !== undefined && pi !== fromPara) return
    para.sentences.forEach((s) => {
      keys.push(s.key)
      raws.push(s.raw)
    })
  })
  return { keys, raws }
}

function startSpeaking(keys: string[], raws: string[]) {
  stopTts()
  if (!raws.length || !('speechSynthesis' in window)) {
    ElMessage.warning('当前浏览器不支持语音朗读')
    return
  }
  // 语音列表尚未加载完成时稍作等待；完全没有目标语言语音则提示
  const voices = window.speechSynthesis.getVoices()
  const langName = fromLang.value === 'de' ? '德语' : '英语'
  if (!voices.length || !voices.some((v) => v.lang.toLowerCase().startsWith(fromLang.value))) {
    ElMessage.warning(`当前浏览器没有可用的${langName}语音，请用 Chrome/Edge 打开本站体验朗读`)
    warmUpVoices()
    return
  }
  tts.playing = true
  ttsHandle = speakTexts(raws, {
    lang: fromLang.value,
    get rate() {
      return ttsRateRef
    },
    onStart: (i) => {
      speakingKey.value = keys[i] || ''
      // 正在读的句子滚入视野（尽量少跳动）
      articleRef.value
        ?.querySelector(`.sent[data-key="${keys[i]}"]`)
        ?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    },
    onDone: () => {
      tts.playing = false
      tts.paused = false
      speakingKey.value = ''
    },
  })
}

function speakAll() {
  const { keys, raws } = collectSentences()
  startSpeaking(keys, raws)
}

function speakParagraph(pi: number) {
  const { keys, raws } = collectSentences(pi)
  startSpeaking(keys, raws)
}

function toggleTtsPause() {
  if (!ttsHandle) return
  if (tts.paused) {
    ttsHandle.resume()
    tts.paused = false
  } else {
    ttsHandle.pause()
    tts.paused = true
  }
}

function changeRate() {
  ttsRateRef = tts.rate
}

function onScroll() {
  tip.show = false
  noteFloat.show = false
}

onMounted(async () => {
  window.addEventListener('scroll', onScroll, { passive: true })
  warmUpVoices()
  await loadChapter()
})

// 栏目内切换章节（下一章）时组件不重挂载，需手动重载数据
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
  // 记录阅读历史（本地，用于书架「最近阅读」继续读）
  if (data.value) {
    pushReadingRecord({
      chapterId: data.value.chapter.id,
      chapterTitle: data.value.chapter.title,
      bookId: data.value.book.id,
      bookTitle: data.value.book.title,
      language: data.value.book.language,
      kind: 'learn',
    })
  }
  loadNotes()
}

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
  if (hoverTimer) clearTimeout(hoverTimer)
  stopTts()
})
</script>

<template>
  <div class="container chapter-wrap" v-if="chapter && book">
    <!-- 顶部工具条 -->
    <div class="chapter-toolbar card">
      <router-link :to="`/learn/book/${book.id}`" class="back-link">← {{ book.title }}</router-link>
      <div class="tools">
        <div class="font-ctrl">
          <button class="tool-btn" @click="changeFontSize(-1)">A−</button>
          <button class="tool-btn" @click="changeFontSize(1)">A＋</button>
        </div>
        <div class="tts-ctrl">
          <button v-if="!tts.playing" class="tool-btn" title="朗读全章" @click="speakAll">▶ 朗读</button>
          <template v-else>
            <button class="tool-btn" @click="toggleTtsPause">{{ tts.paused ? '▶ 继续' : '⏸ 暂停' }}</button>
            <button class="tool-btn" @click="stopTts">⏹ 停止</button>
          </template>
          <select
            class="rate-select"
            v-model.number="tts.rate"
            title="语速"
            @change="changeRate"
          >
            <option :value="0.75">0.75x</option>
            <option :value="1">1x</option>
            <option :value="1.25">1.25x</option>
            <option :value="1.5">1.5x</option>
          </select>
        </div>
        <button v-if="hasTranslation" class="tool-btn" @click="toggleAllTrans">
          {{ allTransOpen ? '收起全部译文' : '展开全部译文' }}
        </button>
        <span v-if="auth.isLoggedIn" class="note-hint">✍ 选中文字可记笔记</span>
      </div>
    </div>

    <div class="chapter-grid">
      <div class="chapter-main">
        <!-- 章节标题（居中） -->
        <header class="chapter-head">
          <h1 class="chapter-title">{{ chapter.title }}</h1>
          <p class="chapter-hint">🖱 悬停单词看释义 · 点击句子翻译 / 再点隐藏</p>
        </header>

        <!-- 正文 -->
        <article ref="articleRef" class="reader card" @mouseup="onArticleMouseup">
          <div v-if="!paragraphs.length" class="empty-chapter">
            📝 本章内容还没有录入，请稍后再来看看～
          </div>
          <div
            v-for="(para, pi) in paragraphs"
            :key="pi"
            class="para"
            :class="{ 'para-active': activePara === pi }"
            :data-idx="pi"
            :style="{ fontSize: fontSize + 'px' }"
          >
            <button
              v-if="auth.isLoggedIn"
              class="para-add-note"
              title="给本段记笔记"
              @click="openNoteCreateForPara(pi)"
            >＋</button>

            <template v-for="(s, si) in para.sentences" :key="s.key">
              <span v-if="si > 0" class="sent-gap">{{ ' ' }}</span>
              <span
                class="sent"
                :class="{ open: sentOpen[s.key], speaking: speakingKey === s.key }"
                :data-key="s.key"
                @click="toggleSentence(s)"
              >
                <template v-for="(tk, ti) in s.tokens" :key="ti">
                  <span
                    v-if="/\S/.test(tk.text)"
                    class="w"
                    :class="{ hl: isHighlighted(para, tk) }"
                    @mouseenter="hoverWord($event, tk.text, s.raw)"
                    @mouseleave="leaveWord"
                  >{{ tk.text }}</span>
                  <template v-else>{{ tk.text }}</template>
                </template>
                <span v-if="sentOpen[s.key]" class="sent-trans">
                  <template v-if="sentTrans[s.key]">{{ sentTrans[s.key] }}</template>
                  <template v-else>翻译中…</template>
                </span>
              </span>
            </template>

            <!-- 本段操作：朗读 / 中文翻译（折叠） -->
            <div class="para-actions" @click.stop>
              <a class="speak-link" title="朗读本段" @click.prevent="speakParagraph(pi)">🔊 朗读本段</a>
              <template v-if="paraTranslations[pi]">
                <a class="trans-toggle" @click.prevent="toggleParaTrans(pi)">
                  {{ paraTransOpen[pi] ? '▲ 收起翻译' : '▼ 查看中文翻译' }}
                </a>
              </template>
            </div>
            <div v-if="paraTransOpen[pi]" class="para-trans">
              {{ paraTranslations[pi] }}
            </div>
          </div>

          <!-- 划选文字后的浮动「记笔记」按钮 -->
          <div
            v-if="noteFloat.show"
            class="note-float"
            :style="{ top: noteFloat.top + 'px', left: noteFloat.left + 'px' }"
            @mousedown.prevent
            @click="openNoteCreate"
          >✍ 记笔记</div>
        </article>

        <!-- 上一章 / 下一章 -->
        <nav class="chapter-nav">
          <router-link v-if="data?.prev" :to="`/learn/chapter/${data.prev.id}`" class="nav-btn card">
            <span class="nav-label">← 上一章</span>
            <span class="nav-title">{{ data.prev.title }}</span>
          </router-link>
          <span v-else class="nav-btn card nav-none">已是第一章</span>
          <router-link v-if="data?.next" :to="`/learn/chapter/${data.next.id}`" class="nav-btn card next">
            <span class="nav-label">下一章 →</span>
            <span class="nav-title">{{ data.next.title }}</span>
          </router-link>
          <span v-else class="nav-btn card nav-none nav-next">已是最后一章 🎉</span>
        </nav>
      </div>

      <!-- 右侧：我的笔记面板 -->
      <aside class="note-panel">
        <div class="note-head">
          <h3>我的笔记 <em>{{ notes.length }}</em></h3>
          <span class="head-links">
            <router-link to="/learn/words" class="words-entry" title="查看生词本">📚 生词本</router-link>
            <router-link to="/learn" class="words-entry" title="最近阅读">🕘 最近阅读</router-link>
          </span>
        </div>

        <div v-if="notes.length" class="note-list">
          <div
            v-for="n in notes"
            :key="n.id"
            class="note-card"
            :class="{ 'note-active': activePara === n.paragraphIndex }"
            @click="scrollToPara(n.paragraphIndex)"
          >
            <div class="note-para">¶ 第 {{ n.paragraphIndex + 1 }} 段</div>
            <blockquote v-if="n.quote" class="note-quote">{{ n.quote }}</blockquote>
            <p class="note-text">{{ n.note }}</p>
            <div class="note-foot">
              <span class="note-time">{{ dayjs(n.createdAt).format('MM-DD HH:mm') }}</span>
              <span v-if="auth.isLoggedIn" class="note-ops" @click.stop>
                <el-button link size="small" @click="openNoteEdit(n)">编辑</el-button>
                <el-popconfirm title="删除这条笔记？" @confirm="removeNote(n)">
                  <template #reference>
                    <el-button link size="small" type="danger">删除</el-button>
                  </template>
                </el-popconfirm>
              </span>
            </div>
          </div>
        </div>
        <p v-else class="note-empty">
          还没有笔记。<template v-if="auth.isLoggedIn">选中左侧文字点「记笔记」，或悬停段落点「＋」。</template>
        </p>
      </aside>
    </div>

    <!-- 笔记编辑弹窗 -->
    <el-dialog v-model="noteDialog.visible" :title="noteDialog.mode === 'create' ? '记笔记' : '编辑笔记'" width="520px">
      <div class="note-form">
        <div class="note-field">
          <label>段落</label>
          <el-tag size="small" type="info">第 {{ noteDialog.paraIndex + 1 }} 段</el-tag>
        </div>
        <div class="note-field">
          <label>摘录原文（可选）</label>
          <el-input v-model="noteDialog.quote" type="textarea" :rows="2" placeholder="划选的原文会自动带出，也可修改" />
        </div>
        <div class="note-field">
          <label>笔记内容 *</label>
          <el-input
            v-model="noteDialog.note"
            type="textarea"
            :rows="5"
            placeholder="记下你的理解、词汇、语法点…"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="noteDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="noteSaving" @click="saveNote">保存</el-button>
      </template>
    </el-dialog>

    <!-- 单词翻译浮层 -->
    <teleport to="body">
      <div
        v-if="tip.show"
        class="word-tip"
        :style="{ left: tip.x + 'px', top: tip.y + 'px' }"
        @mouseenter="keepTip"
        @mouseleave="hideTip"
      >
        <span class="tip-word">{{ tip.word }}</span>
        <span class="tip-trans">{{ tip.loading ? '查询中…' : tip.trans }}</span>
        <button
          v-if="auth.isLoggedIn && !tip.loading"
          class="tip-save"
          title="收藏到生词本"
          @click.stop="saveTipWord"
        >＋ 收藏</button>
      </div>
    </teleport>
  </div>
</template>

<style scoped>
.chapter-wrap {
  max-width: 1100px;
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
  gap: 14px;
  flex-wrap: wrap;
}

.font-ctrl {
  display: flex;
  gap: 6px;
}

.tool-btn {
  border: 1px solid var(--border);
  background: #fff;
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

.tts-ctrl {
  display: flex;
  align-items: center;
  gap: 6px;
}

.rate-select {
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  border-radius: 7px;
  font-size: 12px;
  padding: 3px 6px;
  cursor: pointer;
}

.rate-select:hover {
  color: var(--accent);
  border-color: var(--accent);
}

.note-hint {
  font-size: 12.5px;
  color: var(--muted);
}

/* ===== 双栏布局：正文 + 笔记 ===== */
.chapter-grid {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 22px;
  align-items: start;
}

.chapter-main {
  min-width: 0;
}

.chapter-head {
  margin: 0 0 18px;
  text-align: center;
}

.chapter-title {
  font-family: var(--font-heading);
  font-size: 26px;
  font-weight: 700;
  color: var(--heading);
  margin: 0;
  line-height: 1.5;
}

.chapter-hint {
  margin: 10px 0 0;
  font-size: 13px;
  color: var(--muted);
}

.reader {
  position: relative;
  padding: 36px 44px;
}

.empty-chapter {
  text-align: center;
  color: var(--muted);
  padding: 60px 0;
  font-size: 15px;
}

.para {
  position: relative;
  margin: 0 0 0.9em;
  padding: 4px 10px;
  line-height: 2.05;
  text-align: justify;
  border-radius: 8px;
  transition: background 0.3s;
}

.para:hover {
  background: color-mix(in srgb, var(--accent) 3%, transparent);
}

.para-active {
  animation: flash 1.2s ease;
}

@keyframes flash {
  0% { background: color-mix(in srgb, var(--accent) 18%, transparent); }
  100% { background: transparent; }
}

.para-add-note {
  position: absolute;
  top: 6px;
  right: 8px;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  border: 1px dashed var(--border);
  background: #fff;
  color: var(--muted);
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  opacity: 0;
  transition: all 0.15s;
}

.para:hover .para-add-note {
  opacity: 1;
}

.para-add-note:hover {
  color: var(--accent);
  border-color: var(--accent);
}

/* 句子 */
.sent {
  cursor: pointer;
  border-bottom: 1px dashed transparent;
  transition: background 0.15s;
}

.sent:hover {
  background: color-mix(in srgb, var(--accent) 6%, transparent);
  border-bottom-color: color-mix(in srgb, var(--accent) 35%, transparent);
}

.sent.open {
  background: color-mix(in srgb, var(--accent) 7%, transparent);
}

/* 正在朗读的句子 */
.sent.speaking {
  background: color-mix(in srgb, var(--accent) 16%, transparent);
  border-bottom: 1px solid var(--accent);
  border-radius: 3px;
}

/* 词 */
.w {
  border-radius: 3px;
  transition: background 0.1s;
}

.w:hover {
  background: var(--accent-soft);
  color: var(--accent-hover);
}

/* 笔记摘录的文字：加深颜色标识 */
.w.hl {
  color: var(--accent-hover);
  font-weight: 600;
}

/* 句内译文 */
.sent-trans {
  display: inline;
  font-size: 0.82em;
  color: var(--accent-hover);
  background: var(--accent-soft);
  border: 1px dashed #ecc2b5;
  border-radius: 6px;
  padding: 1px 8px;
  margin-left: 8px;
  line-height: 1.6;
  vertical-align: 1px;
  cursor: pointer;
}

/* 每段下方的操作行（朗读 / 译文折叠） */
.para-actions {
  display: flex;
  align-items: center;
  gap: 18px;
  margin-top: 8px;
  font-size: 13px;
}

.speak-link {
  color: var(--muted);
  cursor: pointer;
  user-select: none;
}

.speak-link:hover {
  color: var(--accent);
}

/* 每段下方的译文折叠入口 */
.trans-toggle {
  display: inline-block;
  color: var(--accent);
  cursor: pointer;
  user-select: none;
}

.trans-toggle:hover {
  color: var(--accent-hover);
  text-decoration: underline;
}

/* 展开的段落译文 */
.para-trans {
  margin-top: 8px;
  padding: 10px 16px;
  border-left: 3px solid var(--accent);
  background: var(--surface);
  color: var(--text);
  font-size: 0.82em;
  line-height: 1.9;
  border-radius: 0 8px 8px 0;
}

/* 划选后的浮动记笔记按钮 */
.note-float {
  position: absolute;
  z-index: 20;
  background: var(--accent);
  color: #fff;
  font-size: 13px;
  padding: 6px 14px;
  border-radius: 999px;
  cursor: pointer;
  box-shadow: var(--shadow-lift);
  user-select: none;
}

.note-float:hover {
  background: var(--accent-hover);
}

/* 章节导航 */
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

/* ===== 右侧笔记面板 ===== */
.note-panel {
  position: sticky;
  top: 86px;
  max-height: calc(100vh - 110px);
  display: flex;
  flex-direction: column;
}

.note-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.note-head h3 {
  font-family: var(--font-heading);
  font-size: 17px;
  margin: 0 0 12px;
}

.words-entry {
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 12px;
  white-space: nowrap;
}

.words-entry:hover {
  color: var(--accent);
}

.head-links {
  display: flex;
  gap: 12px;
}

.note-head em {
  font-style: normal;
  font-size: 13px;
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 999px;
  padding: 1px 9px;
  margin-left: 6px;
}

.note-list {
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 4px;
}

.note-card {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 12px 14px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
}

.note-card:hover {
  background: var(--surface);
  border-color: #d6c9b0;
}

.note-active {
  border-color: var(--accent);
  box-shadow: var(--shadow-lift);
}

.note-para {
  font-size: 11px;
  color: var(--accent);
  font-family: var(--font-mono);
  margin-bottom: 6px;
}

.note-quote {
  margin: 0 0 8px;
  font-size: 12.5px;
  color: var(--muted);
  font-style: italic;
  border-left: 2px solid var(--border);
  padding-left: 10px;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.note-text {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.note-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.note-time {
  font-size: 11px;
  color: var(--muted);
}

.note-empty {
  color: var(--muted);
  font-size: 13.5px;
  background: var(--card);
  border: 1px dashed var(--border);
  border-radius: 10px;
  padding: 22px;
}

/* 笔记弹窗 */
.note-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.note-field label {
  display: block;
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 6px;
}

/* 单词浮层（teleport 到 body，非 scoped 不可） */
:global(.word-tip) {
  position: fixed;
  z-index: 3000;
  background: #261e17;
  color: #f5efe2;
  border-radius: 10px;
  padding: 8px 14px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 8px 24px rgb(38 30 23 / 0.28);
  max-width: 320px;
}

:global(.tip-word) {
  font-family: var(--font-mono);
  font-weight: 700;
  color: #f2b3a5;
}

:global(.tip-trans) {
  color: #f5efe2;
}

:global(.tip-save) {
  border: 1px solid #6b5947;
  background: transparent;
  color: #f2b3a5;
  font-size: 12px;
  border-radius: 999px;
  padding: 2px 10px;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
}

:global(.tip-save:hover) {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

@media (max-width: 1000px) {
  .chapter-grid {
    grid-template-columns: 1fr;
  }

  .note-panel {
    position: static;
    max-height: none;
  }

  .reader {
    padding: 24px 18px;
  }
}
</style>
