<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { createArtComment, fetchArt, fetchArtComments } from '@/api'
import type { ArtCommentItem, ArtworkItem } from '@/types'
import { renderMarkdown } from '@/utils/markdown'

const route = useRoute()
const art = ref<ArtworkItem | null>(null)
const comments = ref<ArtCommentItem[]>([])
const html = ref('')
const submitting = ref(false)

const form = reactive({ nickname: '', email: '', content: '' })

async function load() {
  const data = await fetchArt(route.params.id as string)
  art.value = data
  html.value = renderMarkdown(data.content || '')
  comments.value = (await fetchArtComments(data.id)) || []
}

async function submit() {
  if (!form.nickname.trim() || !form.content.trim()) {
    ElMessage.warning('请填写昵称和评论内容')
    return
  }
  submitting.value = true
  try {
    await createArtComment(art.value!.id, {
      nickname: form.nickname.trim(),
      email: form.email.trim() || undefined,
      content: form.content.trim(),
    })
    ElMessage.success('评论成功！')
    form.content = ''
    comments.value = (await fetchArtComments(art.value!.id)) || []
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="container art-detail" v-if="art">
    <router-link to="/art" class="back-link">← 艺术鉴赏</router-link>

    <!-- 大图 + 信息 -->
    <div class="card art-card">
      <h1 class="art-title">{{ art.title }}</h1>
      <p class="art-meta">
        <span v-if="art.author" class="author">{{ art.author }}</span>
        <span>{{ dayjs(art.createdAt).format('YYYY-MM-DD') }}</span>
      </p>

      <!-- 大图：点击全屏查看 -->
      <div class="art-figure">
        <el-image
          v-if="art.image"
          :src="art.image"
          :preview-src-list="[art.image]"
          fit="contain"
          hide-on-click-modal
          preview-teleported
          class="art-big"
        >
          <template #placeholder>
            <div class="img-loading">加载中…</div>
          </template>
        </el-image>
        <div v-else class="art-big art-placeholder">{{ art.title.slice(0, 1) }}</div>
        <p class="figure-tip">点击图片可全屏欣赏</p>
      </div>

      <p v-if="art.description" class="art-desc">{{ art.description }}</p>
      <div v-if="html" class="md-content art-content" v-html="html" />
    </div>

    <!-- 评论区 -->
    <section class="card comment-card">
      <h3 class="comment-title">评论 <span>{{ comments.length }}</span></h3>

      <div v-if="comments.length" class="comment-list">
        <div v-for="cm in comments" :key="cm.id" class="comment-item">
          <div class="comment-avatar">{{ cm.nickname.slice(0, 1).toUpperCase() }}</div>
          <div class="comment-main">
            <div class="comment-head">
              <span class="comment-name">{{ cm.nickname }}</span>
              <span class="comment-time">{{ dayjs(cm.createdAt).format('YYYY-MM-DD HH:mm') }}</span>
            </div>
            <p class="comment-content">{{ cm.content }}</p>
          </div>
        </div>
      </div>
      <p v-else class="comment-empty">还没有评论，来说说你看到这幅作品的感受吧～</p>

      <div class="comment-form">
        <div class="form-row">
          <el-input v-model="form.nickname" placeholder="昵称 *" maxlength="100" />
          <el-input v-model="form.email" placeholder="邮箱（选填，不公开）" />
        </div>
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="3"
          placeholder="写下你的感受… *"
          maxlength="2000"
          show-word-limit
        />
        <div class="form-actions">
          <el-button type="primary" :loading="submitting" @click="submit">发表评论</el-button>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.art-detail {
  max-width: 860px;
}

.back-link {
  font-size: 13px;
  color: var(--muted);
  display: inline-block;
  margin-bottom: 12px;
}

.back-link:hover {
  color: var(--accent);
}

.art-card {
  padding: 30px 36px;
}

.art-title {
  font-family: var(--font-heading);
  font-size: 27px;
  margin: 0 0 6px;
  text-align: center;
}

.art-meta {
  display: flex;
  justify-content: center;
  gap: 14px;
  font-size: 13px;
  color: var(--muted);
  margin: 0 0 22px;
}

.author {
  font-style: italic;
}

.art-figure {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.art-big {
  width: 100%;
  max-width: 760px;
  height: auto;
  min-height: 320px;
  border-radius: 10px;
  overflow: hidden;
  cursor: zoom-in;
  background: linear-gradient(150deg, var(--surface-alt), var(--surface));
}

.art-big :deep(img) {
  width: 100%;
  height: auto;
  max-height: 560px;
  object-fit: contain;
}

.art-placeholder {
  display: grid;
  place-items: center;
  min-height: 320px;
  font-family: var(--font-heading);
  font-size: 64px;
  font-weight: 700;
  color: var(--muted);
  opacity: 0.5;
}

.img-loading {
  display: grid;
  place-items: center;
  height: 320px;
  color: var(--muted);
}

.figure-tip {
  font-size: 12px;
  color: var(--muted);
  margin: 8px 0 0;
}

.art-desc {
  font-size: 14.5px;
  color: var(--text);
  line-height: 1.8;
  margin: 20px 0 0;
}

.art-content {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px dashed var(--border);
}

/* 评论区 */
.comment-card {
  margin-top: 20px;
  padding: 24px 28px;
}

.comment-title {
  font-family: var(--font-heading);
  font-size: 19px;
  margin: 0 0 18px;
}

.comment-title span {
  color: var(--accent);
  font-size: 14px;
  margin-left: 4px;
}

.comment-item {
  display: flex;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px dashed var(--border);
}

.comment-item:last-child {
  border-bottom: none;
}

.comment-avatar {
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border-radius: 50%;
  background: var(--accent-soft);
  color: var(--accent);
  display: grid;
  place-items: center;
  font-weight: 700;
}

.comment-main {
  flex: 1;
  min-width: 0;
}

.comment-head {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.comment-name {
  font-weight: 600;
  font-size: 14px;
}

.comment-time {
  font-size: 12px;
  color: var(--muted);
}

.comment-content {
  margin: 6px 0 0;
  font-size: 14px;
  white-space: pre-wrap;
  word-break: break-word;
}

.comment-empty {
  color: var(--muted);
  font-size: 14px;
  padding: 8px 0 16px;
}

.comment-form {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}
</style>
