<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { createComment, fetchComments } from '@/api'
import type { CommentItem } from '@/types'

const props = defineProps<{ articleId: number }>()

const list = ref<CommentItem[]>([])
const submitting = ref(false)

const form = reactive({
  nickname: '',
  email: '',
  content: '',
})

async function load() {
  list.value = (await fetchComments(props.articleId)) || []
}

async function submit() {
  if (!form.nickname.trim() || !form.content.trim()) {
    ElMessage.warning('请填写昵称和评论内容')
    return
  }
  submitting.value = true
  try {
    await createComment(props.articleId, {
      nickname: form.nickname.trim(),
      email: form.email.trim() || undefined,
      content: form.content.trim(),
    })
    ElMessage.success('评论成功！')
    form.content = ''
    load()
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <section class="comments card">
    <h3 class="comments-title">评论 <span>{{ list.length }}</span></h3>

    <div v-if="list.length" class="comment-list">
      <div v-for="cm in list" :key="cm.id" class="comment-item">
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
</template>

<style scoped>
.comments {
  margin-top: 30px;
  padding: 24px 28px;
}

.comments-title {
  font-family: var(--font-heading);
  font-size: 19px;
  margin: 0 0 18px;
}

.comments-title span {
  color: var(--accent);
  font-size: 14px;
  margin-left: 4px;
}

.comment-list {
  display: flex;
  flex-direction: column;
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
  color: #44403c;
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
