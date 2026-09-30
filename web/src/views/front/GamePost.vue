<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { createGameComment, fetchGameComments, fetchGamePost } from '@/api'
import type { GameCommentItem, GameItem, GamePostItem } from '@/types'
import { renderMarkdown } from '@/utils/markdown'

const route = useRoute()
const post = ref<GamePostItem | null>(null)
const game = ref<GameItem | null>(null)
const comments = ref<GameCommentItem[]>([])
const html = ref('')
const submitting = ref(false)

const form = reactive({ nickname: '', email: '', content: '' })

async function load() {
  const data = await fetchGamePost(route.params.id as string)
  post.value = data.post
  game.value = data.game
  html.value = renderMarkdown(data.post.content || '')
  comments.value = (await fetchGameComments(data.post.id)) || []
}

async function submit() {
  if (!form.nickname.trim() || !form.content.trim()) {
    ElMessage.warning('请填写昵称和评论内容')
    return
  }
  submitting.value = true
  try {
    await createGameComment(post.value!.id, {
      nickname: form.nickname.trim(),
      email: form.email.trim() || undefined,
      content: form.content.trim(),
    })
    ElMessage.success('评论成功！')
    form.content = ''
    comments.value = (await fetchGameComments(post.value!.id)) || []
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="container post-wrap" v-if="post && game">
    <div class="crumb">
      <router-link to="/game">游戏世界</router-link>
      <span>/</span>
      <router-link :to="`/game/${game.id}`">{{ game.title }}</router-link>
    </div>

    <article class="card post-card">
      <div class="post-head">
        <em v-if="post.type" class="type-badge">{{ post.type }}</em>
        <span class="post-meta">
          {{ dayjs(post.createdAt).format('YYYY-MM-DD HH:mm') }} · {{ post.views }} 次浏览
        </span>
      </div>
      <h1 class="post-title">{{ post.title }}</h1>
      <div class="md-content" v-html="html" />
    </article>

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
      <p v-else class="comment-empty">还没有评论，来说点什么吧～</p>

      <div class="comment-form">
        <div class="form-row">
          <el-input v-model="form.nickname" placeholder="昵称 *" maxlength="100" />
          <el-input v-model="form.email" placeholder="邮箱（选填，不公开）" />
        </div>
        <el-input
          v-model="form.content"
          type="textarea"
          :rows="3"
          placeholder="写下你的评论… *"
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
.post-wrap {
  max-width: 820px;
}

.crumb {
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
}

.crumb a:hover {
  color: var(--accent);
}

.post-card {
  padding: 28px 34px;
}

.post-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.type-badge {
  font-style: normal;
  font-size: 12px;
  color: var(--accent);
  background: var(--accent-soft);
  padding: 1px 10px;
  border-radius: 999px;
}

.post-meta {
  font-size: 12.5px;
  color: var(--muted);
}

.post-title {
  font-family: var(--font-heading);
  font-size: 26px;
  margin: 12px 0 18px;
  line-height: 1.5;
}

/* 评论区（与文章评论一致） */
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
