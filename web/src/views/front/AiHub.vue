<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import { fetchAiPrompts, fetchAiTools } from '@/api'
import type { AiPromptItem, AiToolItem } from '@/types'

const tools = ref<AiToolItem[]>([])
const prompts = ref<AiPromptItem[]>([])
const loading = ref(false)

/** 一级大类（AI编程 / AI绘画…），从数据提取；默认聚焦第一个 */
const activeSection = ref('')
/** 二级分类筛选 */
const activeCat = ref('')

const sections = computed(() => {
  const set = new Set<string>()
  for (const t of tools.value) if (t.section) set.add(t.section)
  for (const p of prompts.value) if (p.section) set.add(p.section)
  return [...set]
})

const currentSection = computed(() => activeSection.value || sections.value[0] || '')

const categories = computed(() => {
  const set = new Set<string>()
  for (const t of tools.value) if (t.section === currentSection.value && t.category) set.add(t.category)
  return [...set]
})

const filteredTools = computed(() =>
  tools.value.filter(
    (t) =>
      t.section === currentSection.value &&
      (!activeCat.value || t.category === activeCat.value),
  ),
)

const sectionPrompts = computed(() =>
  prompts.value.filter((p) => p.section === currentSection.value),
)

const parseTags = (tags: string): string[] =>
  tags.split(/[,，]/).map((t) => t.trim()).filter(Boolean)

const expanded = ref<number | null>(null)

async function copyPrompt(p: AiPromptItem) {
  try {
    await navigator.clipboard.writeText(p.content)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动选择复制')
  }
}

onMounted(async () => {
  loading.value = true
  try {
    const [t, p] = await Promise.all([fetchAiTools(), fetchAiPrompts()])
    tools.value = t || []
    prompts.value = p || []
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="container ai-wrap">
    <div class="ai-hero card">
      <h1 class="hero-title">🤖 AI 百宝箱</h1>
      <p class="hero-sub">我在用的 AI 工具和提示词，按领域细分整理，持续更新。</p>
    </div>

    <!-- 一级大类 Tab -->
    <div v-if="sections.length" class="section-tabs">
      <button
        v-for="s in sections"
        :key="s"
        class="section-tab"
        :class="{ active: currentSection === s }"
        @click="activeSection = s; activeCat = ''"
      >
        {{ s }}
      </button>
    </div>

    <div v-loading="loading">
      <!-- 二级分类筛选 -->
      <div v-if="categories.length" class="cat-tabs">
        <button class="cat-tab" :class="{ active: !activeCat }" @click="activeCat = ''">全部</button>
        <button
          v-for="cat in categories"
          :key="cat"
          class="cat-tab"
          :class="{ active: activeCat === cat }"
          @click="activeCat = activeCat === cat ? '' : cat"
        >
          {{ cat }}
        </button>
      </div>

      <!-- 工具卡片 -->
      <div v-if="filteredTools.length" class="tool-grid">
        <a
          v-for="t in filteredTools"
          :key="t.id"
          :href="t.url"
          target="_blank"
          rel="noopener"
          class="tool-card card"
        >
          <div class="tool-icon">
            <img v-if="t.icon" :src="t.icon" alt="" />
            <span v-else>{{ t.name.slice(0, 1) }}</span>
          </div>
          <div class="tool-info">
            <div class="tool-name">
              {{ t.name }}
              <span class="tool-cat" v-if="t.category">{{ t.category }}</span>
            </div>
            <p class="tool-desc">{{ t.description }}</p>
            <div class="tool-tags">
              <span v-for="tag in parseTags(t.tags)" :key="tag" class="tool-tag">{{ tag }}</span>
            </div>
          </div>
        </a>
      </div>
      <el-empty v-else-if="!loading" description="这个分类下还没有工具" />

      <!-- 提示词 -->
      <section v-if="sectionPrompts.length" class="prompt-sec">
        <h3 class="prompt-title">💡 提示词收藏</h3>
        <div
          v-for="p in sectionPrompts"
          :key="p.id"
          class="prompt-card card"
          @click="expanded = expanded === p.id ? null : p.id"
        >
          <div class="prompt-head">
            <span class="prompt-name">{{ p.title }}</span>
            <span class="prompt-desc">{{ p.description }}</span>
            <button
              class="copy-btn"
              title="复制提示词"
              @click.stop="copyPrompt(p)"
            >
              <el-icon><CopyDocument /></el-icon> 复制
            </button>
          </div>
          <pre v-if="expanded === p.id" class="prompt-content">{{ p.content }}</pre>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.ai-wrap {
  max-width: 1020px;
}

.ai-hero {
  padding: 32px 40px;
  margin-bottom: 18px;
  background:
    radial-gradient(circle at 92% -30%, #edeafd, transparent 50%),
    var(--card);
}

.hero-title {
  font-family: var(--font-heading);
  font-size: 30px;
  margin: 0 0 8px;
}

.hero-sub {
  color: var(--muted);
  font-size: 14.5px;
  margin: 0;
}

/* ===== 一级大类 Tab ===== */
.section-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.section-tab {
  border: none;
  background: var(--card);
  border: 1px solid var(--border);
  color: var(--muted);
  font-size: 15px;
  font-weight: 600;
  padding: 8px 22px;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.2s;
}

.section-tab.active {
  background: var(--ai-accent, #5a4df8);
  border-color: var(--ai-accent, #5a4df8);
  color: #fff;
}

/* ===== 二级分类 ===== */
.cat-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.cat-tab {
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  font-size: 13px;
  padding: 4px 14px;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.15s;
}

.cat-tab.active {
  background: color-mix(in srgb, var(--ai-accent, #5a4df8) 12%, transparent);
  border-color: var(--ai-accent, #5a4df8);
  color: var(--ai-accent, #5a4df8);
  font-weight: 600;
}

/* ===== 工具卡片 ===== */
.tool-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.tool-card {
  display: flex;
  gap: 14px;
  padding: 16px 18px;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.tool-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-lift);
  border-color: var(--ai-accent, #5a4df8);
}

.tool-icon {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
  border-radius: 10px;
  background: color-mix(in srgb, var(--ai-accent, #5a4df8) 10%, transparent);
  color: var(--ai-accent, #5a4df8);
  display: grid;
  place-items: center;
  font-family: var(--font-heading);
  font-size: 20px;
  font-weight: 800;
  overflow: hidden;
}

.tool-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.tool-info {
  flex: 1;
  min-width: 0;
}

.tool-name {
  font-size: 15.5px;
  font-weight: 700;
  color: var(--heading);
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-cat {
  font-size: 11px;
  font-weight: 400;
  color: var(--ai-accent, #5a4df8);
  background: color-mix(in srgb, var(--ai-accent, #5a4df8) 10%, transparent);
  padding: 1px 8px;
  border-radius: 999px;
}

.tool-desc {
  font-size: 13px;
  color: var(--muted);
  line-height: 1.6;
  margin: 4px 0 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tool-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tool-tag {
  font-size: 11px;
  color: var(--muted);
  border: 1px solid var(--border);
  padding: 0 7px;
  border-radius: 999px;
}

/* ===== 提示词 ===== */
.prompt-sec {
  margin-top: 28px;
}

.prompt-title {
  font-family: var(--font-heading);
  font-size: 18px;
  margin: 0 0 12px;
}

.prompt-card {
  padding: 14px 18px;
  margin-bottom: 10px;
  cursor: pointer;
}

.prompt-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.prompt-name {
  font-size: 14.5px;
  font-weight: 700;
  color: var(--heading);
  white-space: nowrap;
}

.prompt-desc {
  flex: 1;
  font-size: 13px;
  color: var(--muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--border);
  background: var(--card);
  color: var(--muted);
  font-size: 12px;
  padding: 4px 12px;
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.copy-btn:hover {
  color: var(--ai-accent, #5a4df8);
  border-color: var(--ai-accent, #5a4df8);
}

.prompt-content {
  margin: 12px 0 0;
  padding: 14px 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  font-family: var(--font-mono);
  font-size: 12.5px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 700px) {
  .tool-grid {
    grid-template-columns: 1fr;
  }
}
</style>
