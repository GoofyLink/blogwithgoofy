<script setup lang="ts">
import { onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { adminFetchComments, deleteComment } from '@/api'
import AdminPagination from '@/components/AdminPagination.vue'
import type { CommentItem } from '@/types'

const list = ref<CommentItem[]>([])
const total = ref(0)
const page = ref(1)
const size = 10
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data = await adminFetchComments({ page: page.value, size })
    list.value = data.list || []
    total.value = data.total
  } finally {
    loading.value = false
  }
}

async function remove(row: CommentItem) {
  await deleteComment(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <el-table :data="list" v-loading="loading" stripe class="card table-card">
      <el-table-column label="昵称" prop="nickname" width="120" />
      <el-table-column label="内容" min-width="280">
        <template #default="{ row }">
          <span class="comment-text">{{ row.content }}</span>
        </template>
      </el-table-column>
      <el-table-column label="文章" min-width="160">
        <template #default="{ row }">
          <router-link
            v-if="row.articleId"
            :to="`/article/${row.articleId}`"
            class="article-link"
            target="_blank"
          >
            {{ row.articleTitle || `#${row.articleId}` }}
          </router-link>
        </template>
      </el-table-column>
      <el-table-column label="时间" width="150">
        <template #default="{ row }">{{ dayjs(row.createdAt).format('YYYY-MM-DD HH:mm') }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-popconfirm title="删除这条评论？" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrap" v-if="total > size">
      <AdminPagination :page="page" :total="total" :size="size" @update:page="(p: number) => { page = p; load() }" /></div>
  </div>
</template>

<style scoped>
.table-card {
  width: 100%;
}

.comment-text {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  font-size: 13.5px;
}

.article-link {
  color: var(--el-color-primary);
  font-size: 13px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
