<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { fetchWords } from '@/api'
import type { WordItem } from '@/types'
import { speakTexts } from '@/utils/tts'

const route = useRoute()
const router = useRouter()

const phase = ref<'loading' | 'review' | 'done'>('loading')
const queue = ref<WordItem[]>([])
const pos = ref(0)
const flipped = ref(false)
const knownCount = ref(0)
const unknownSet = ref<Set<number>>(new Set())
const rounds = ref(1)

const current = computed(() => queue.value[pos.value] || null)
const total = computed(() => queue.value.length)
const remaining = computed(() => total.value - pos.value)

function shuffle<T>(arr: T[]): T[] {
  const a = [...arr]
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[a[i], a[j]] = [a[j], a[i]]
  }
  return a
}

async function load() {
  const lang = (route.query.lang as string) || ''
  const words = (await fetchWords(lang ? { language: lang } : undefined)) || []
  if (!words.length) {
    ElMessage.warning('生词本还是空的，先去收藏几个单词吧')
    router.replace('/learn/words')
    return
  }
  queue.value = shuffle(words)
  phase.value = 'review'
}

function flip() {
  flipped.value = !flipped.value
}

/** 认识：进入下一个；全部过完则进入总结 */
function markKnown() {
  if (!flipped.value) {
    flip()
    return
  }
  knownCount.value++
  next()
}

/** 不认识：当前词重新排到队尾，再来一遍 */
function markUnknown() {
  if (!flipped.value) {
    flip()
    return
  }
  const w = queue.value[pos.value]
  if (w) {
    unknownSet.value.add(w.id)
    queue.value.push(w)
  }
  next()
}

function next() {
  flipped.value = false
  if (pos.value + 1 >= total.value) {
    phase.value = 'done'
    return
  }
  pos.value++
}

/** 只复习刚才不认识的词 */
function reviewUnknown() {
  const unknown = queue.value.filter((w) => unknownSet.value.has(w.id))
  if (!unknown.length) return
  queue.value = shuffle(unknown)
  unknownSet.value = new Set()
  pos.value = 0
  knownCount.value = 0
  rounds.value++
  flipped.value = false
  phase.value = 'review'
}

function restartAll() {
  queue.value = shuffle(queue.value)
  unknownSet.value = new Set()
  pos.value = 0
  knownCount.value = 0
  rounds.value = 1
  flipped.value = false
  phase.value = 'review'
}

function pronounce(w: WordItem) {
  if (!('speechSynthesis' in window)) return
  speakTexts([w.word], { lang: w.language, rate: 1 })
}

function onKey(e: KeyboardEvent) {
  if (phase.value !== 'review') return
  if (e.code === 'Space' || e.code === 'Enter') {
    e.preventDefault()
    flip()
  } else if (e.key === '1') {
    markUnknown()
  } else if (e.key === '2') {
    markKnown()
  }
}

onMounted(() => {
  load()
  window.addEventListener('keydown', onKey)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKey)
})
</script>

<template>
  <div class="container review-wrap">
    <div class="review-head card">
      <router-link to="/learn/words" class="back-link">← 生词本</router-link>
      <div class="head-row">
        <h1 class="page-title">🎯 复习模式</h1>
        <p class="page-sub">
          看到单词先回想意思，点击卡片翻面核对。<b>认识</b>过下一个，<b>不认识</b>的会自动排到队尾再来。
          键盘：空格翻面 · 1 不认识 · 2 认识
        </p>
      </div>
    </div>

    <!-- 复习中 -->
    <template v-if="phase === 'review' && current">
      <div class="progress-row">
        <el-progress
          :percentage="Math.round((pos / total) * 100)"
          :stroke-width="10"
          class="progress"
        />
        <span class="progress-text">{{ pos + 1 }} / {{ total }} · 第 {{ rounds }} 轮</span>
      </div>

      <!-- 翻转卡片 -->
      <div class="card-area" @click="flip">
        <div class="flip-card" :class="{ flipped }">
          <div class="face front">
            <span class="word-lang" :class="current.language">{{ current.language === 'de' ? 'DE' : 'EN' }}</span>
            <h2 class="face-word" @click.stop="pronounce(current)">🔊 {{ current.word }}</h2>
            <p class="face-hint">想出意思后，点击卡片翻面</p>
          </div>
          <div class="face back">
            <p class="back-trans">{{ current.translation || '（无释义）' }}</p>
            <p v-if="current.sourceText" class="back-source">“{{ current.sourceText }}”</p>
            <p class="face-hint">1 = 不认识 · 2 = 认识</p>
          </div>
        </div>
      </div>

      <div class="judge-row">
        <el-button size="large" class="judge-btn unknown" @click="markUnknown">😵 不认识（1）</el-button>
        <el-button size="large" type="primary" class="judge-btn" @click="markKnown">😆 认识（2）</el-button>
      </div>
    </template>

    <!-- 完成 -->
    <div v-if="phase === 'done'" class="done-card card">
      <h2 class="done-title">🎉 本轮完成！</h2>
      <div class="done-stats">
        <div class="done-stat">
          <span class="stat-num">{{ knownCount }}</span>
          <span class="stat-label">认识</span>
        </div>
        <div class="done-stat">
          <span class="stat-num warn">{{ unknownSet.size }}</span>
          <span class="stat-label">待巩固</span>
        </div>
        <div class="done-stat">
          <span class="stat-num">{{ rounds }}</span>
          <span class="stat-label">轮次</span>
        </div>
      </div>
      <div class="done-actions">
        <el-button v-if="unknownSet.size" type="primary" size="large" @click="reviewUnknown">
          只复习待巩固的 {{ unknownSet.size }} 个
        </el-button>
        <el-button size="large" @click="restartAll">全部重来</el-button>
        <el-button size="large" @click="router.push('/learn/words')">返回生词本</el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.review-wrap {
  max-width: 720px;
}

.review-head {
  padding: 22px 30px;
  margin-bottom: 20px;
}

.back-link {
  font-size: 13px;
  color: var(--muted);
}

.back-link:hover {
  color: var(--accent);
}

.page-title {
  font-family: var(--font-heading);
  font-size: 26px;
  margin: 8px 0 4px;
}

.page-sub {
  color: var(--muted);
  font-size: 13.5px;
  margin: 0;
}

.page-sub b {
  color: var(--accent);
}

.progress-row {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 18px;
}

.progress {
  flex: 1;
}

.progress-text {
  font-size: 12.5px;
  color: var(--muted);
  white-space: nowrap;
  font-family: var(--font-mono);
}

/* ===== 翻转卡片 ===== */
.card-area {
  perspective: 1200px;
  cursor: pointer;
}

.flip-card {
  position: relative;
  height: 300px;
  transform-style: preserve-3d;
  transition: transform 0.55s cubic-bezier(0.3, 1.15, 0.45, 1);
}

.flip-card.flipped {
  transform: rotateY(180deg);
}

.face {
  position: absolute;
  inset: 0;
  backface-visibility: hidden;
  border-radius: 16px;
  border: 1px solid var(--border);
  background: var(--card);
  box-shadow: var(--shadow-lift);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 30px;
}

.face.back {
  transform: rotateY(180deg);
  background: linear-gradient(150deg, var(--card), var(--accent-soft));
}

.word-lang {
  font-size: 11px;
  font-family: var(--font-mono);
  font-weight: 700;
  border-radius: 6px;
  padding: 1px 8px;
  color: var(--olive);
  background: var(--badge-en-bg);
}

.face-word {
  font-family: var(--font-heading);
  font-size: 38px;
  color: var(--heading);
  margin: 0;
  cursor: pointer;
}

.face-word:hover {
  color: var(--accent);
}

.face-hint {
  font-size: 12.5px;
  color: var(--muted);
  margin: 0;
}

.back-trans {
  font-size: 26px;
  font-weight: 700;
  color: var(--heading);
  margin: 0;
  text-align: center;
  font-family: var(--font-heading);
}

.back-source {
  font-size: 13px;
  color: var(--muted);
  font-style: italic;
  margin: 0;
  text-align: center;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ===== 判定按钮 ===== */
.judge-row {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin-top: 24px;
}

.judge-btn {
  min-width: 160px;
  height: 48px;
  font-size: 15px;
  border-radius: 12px;
}

.judge-btn.unknown:hover {
  color: var(--accent);
  border-color: var(--accent);
}

/* ===== 完成页 ===== */
.done-card {
  padding: 40px;
  text-align: center;
}

.done-title {
  font-family: var(--font-heading);
  font-size: 26px;
  margin: 0 0 24px;
}

.done-stats {
  display: flex;
  justify-content: center;
  gap: 44px;
  margin-bottom: 28px;
}

.done-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-num {
  font-family: var(--font-heading);
  font-size: 34px;
  font-weight: 800;
  color: var(--heading);
}

.stat-num.warn {
  color: var(--accent);
}

.stat-label {
  font-size: 13px;
  color: var(--muted);
}

.done-actions {
  display: flex;
  justify-content: center;
  gap: 10px;
  flex-wrap: wrap;
}
</style>
