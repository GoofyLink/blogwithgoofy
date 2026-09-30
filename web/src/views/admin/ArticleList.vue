<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { adminFetchArticles, deleteArticle } from '@/api'
import type { Article } from '@/types'

const router = useRouter()

const list = ref<Article[]>([])
const total = ref(0)
const page = ref(1)
const size = 10
const loading = ref(false)
const keyword = ref('')
const typeFilter = ref<string>('')

async function load() {
  loading.value = true
  try {
    const data = await adminFetchArticles({
      page: page.value,
      size,
      keyword: keyword.value,
      type: typeFilter.value,
    })
    list.value = data.list || []
    total.value = data.total
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  load()
}

async function remove(a: Article) {
  await deleteArticle(a.id)
  ElMessage.success('已删除')
  load()
}

function edit(a: Article) {
  router.push(`/admin/articles/edit/${a.id}`)
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <router-link to="/admin/articles/new">
        <el-button type="primary">＋ 新建文章</el-button>
      </router-link>
      <div class="toolbar-right">
        <el-input
          v-model="keyword"
          placeholder="搜索标题…"
          clearable
          style="width: 220px"
          @keyup.enter="search"
          @clear="search"
        />
        <el-select v-model="typeFilter" placeholder="全部类型" clearable style="width: 130px" @change="search">
          <el-option label="博客文章" value="0" />
          <el-option label="学习文章" value="1" />
        </el-select>
      </div>
    </div>

    <el-table :data="list" v-loading="loading" stripe class="card table-card">
      <el-table-column label="标题" min-width="260">
        <template #default="{ row }">
          <span class="row-title" @click="edit(row)">{{ row.title }}</span>
        </template>
      </el-table-column>
      <el-table-column label="分类" width="110">
        <template #default="{ row }">{{ row.category?.name || '—' }}</template>
      </el-table-column>
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag :type="row.type === 1 ? 'warning' : 'info'" size="small" effect="plain">
            {{ row.type === 1 ? `学习 ${row.language.toUpperCase()}` : '博客' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="views" label="浏览" width="80" />
      <el-table-column label="创建时间" width="110">
        <template #default="{ row }">{{ dayjs(row.createdAt).format('YYYY-MM-DD') }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="edit(row)">编辑</el-button>
          <router-link v-if="row.status === 1" :to="`/article/${row.id}`" class="view-link">
            查看
          </router-link>
          <el-popconfirm title="删除后无法恢复，确定删除？" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrap" v-if="total > size">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        :page-size="size"
        :current-page="page"
        @current-change="(p: number) => { page = p; load() }"
      />
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 14px;
}

.toolbar-right {
  display: flex;
  gap: 10px;
}

.table-card {
  width: 100%;
}

.row-title {
  cursor: pointer;
  font-weight: 500;
}

.row-title:hover {
  color: var(--accent);
}

.view-link {
  font-size: 12px;
  color: var(--el-color-primary);
  margin: 0 8px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
</style>
