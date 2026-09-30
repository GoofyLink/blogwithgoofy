<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  adminFetchGameComments, adminFetchGamePosts, adminFetchGames,
  createGamePost, deleteGameComment, deleteGamePost, updateGamePost,
} from '@/api'
import type { GameCommentItem, GameItem, GamePostItem } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

const route = useRoute()

const games = ref<GameItem[]>([])
const posts = ref<GamePostItem[]>([])
const comments = ref<GameCommentItem[]>([])
const loading = ref(false)

const gameFilter = ref<string>(route.query.gameId as string || '')

const filteredPosts = computed(() =>
  gameFilter.value ? posts.value.filter((p) => String(p.gameId) === gameFilter.value) : posts.value,
)
const { page, pagedList } = usePagination(filteredPosts)

const gameName = (gid: number): string => games.value.find((g) => g.id === gid)?.title || `#${gid}`

// 帖子类型选项（已有 + 可新建）
const newTypes = ref<string[]>([])
const typeOptions = computed(() => {
  const set = new Set<string>()
  for (const p of posts.value) if (p.type) set.add(p.type)
  return [...new Set([...set, ...newTypes.value])]
})

// 评论分页（服务端）
const cPage = ref(1)
const cSize = 10
const cTotal = ref(0)

async function load() {
  loading.value = true
  try {
    const [g, p] = await Promise.all([adminFetchGames(), adminFetchGamePosts()])
    games.value = g || []
    posts.value = p || []
  } finally {
    loading.value = false
  }
}

async function loadComments() {
  const data = await adminFetchGameComments({ page: cPage.value, size: cSize })
  comments.value = data.list || []
  cTotal.value = data.total
}

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    gameId: 0,
    title: '',
    type: '攻略',
    summary: '',
    content: '',
    status: 1 as 0 | 1,
  },
})

function openCreate() {
  dialog.mode = 'create'
  dialog.editId = 0
  dialog.data = {
    gameId: Number(gameFilter.value) || games.value[0]?.id || 0,
    title: '', type: '攻略', summary: '', content: '', status: 1,
  }
  dialog.visible = true
}

function openEdit(row: GamePostItem) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    gameId: row.gameId, title: row.title, type: row.type,
    summary: row.summary, content: row.content || '', status: row.status,
  }
  dialog.visible = true
}

async function save() {
  if (!dialog.data.title.trim() || !dialog.data.content.trim() || !dialog.data.gameId) {
    ElMessage.warning('请选择游戏并填写标题和内容')
    return
  }
  if (dialog.mode === 'create') {
    await createGamePost({ ...dialog.data })
    ElMessage.success('已发布')
  } else {
    await updateGamePost(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: GamePostItem) {
  await deleteGamePost(row.id)
  ElMessage.success('已删除')
  load()
}

async function removeComment(row: GameCommentItem) {
  await deleteGameComment(row.id)
  ElMessage.success('已删除')
  loadComments()
}

const wc = (s: string): number => (s ? s.trim().split(/\s+/).length : 0)

onMounted(() => {
  load()
  loadComments()
})
</script>

<template>
  <div>
    <!-- 攻略资讯管理 -->
    <div class="toolbar card">
      <el-button type="primary" @click="openCreate">＋ 发布攻略/资讯</el-button>
      <el-select v-model="gameFilter" placeholder="全部游戏" clearable style="width: 180px">
        <el-option v-for="g in games" :key="g.id" :label="g.title" :value="String(g.id)" />
      </el-select>
      <span class="tip">前台展示在对应游戏的「攻略资讯圈」，读者可评论</span>
    </div>

    <el-table :data="pagedList" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="id" label="ID" width="65" />
      <el-table-column prop="title" label="标题" min-width="220">
        <template #default="{ row }">
          <el-tag size="small" effect="plain" type="warning">{{ row.type }}</el-tag>
          <span class="row-gap">{{ row.title }}</span>
        </template>
      </el-table-column>
      <el-table-column label="游戏" width="120">
        <template #default="{ row }">{{ gameName(row.gameId) }}</template>
      </el-table-column>
      <el-table-column prop="views" label="浏览" width="75" />
      <el-table-column label="字数" width="75">
        <template #default="{ row }">{{ wc(row.content || '') }}</template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="145" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除「${row.title}」及其评论？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="page" :total="filteredPosts.length" />

    <!-- 评论管理 -->
    <div class="toolbar card gap-top">
      <el-button :loading="loading" disabled plain>评论管理（{{ cTotal }}）</el-button>
      <span class="tip">游戏圈读者评论</span>
    </div>
    <el-table :data="comments" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="nickname" label="昵称" width="110" />
      <el-table-column prop="content" label="内容" min-width="220" show-overflow-tooltip />
      <el-table-column prop="postTitle" label="所属帖子" min-width="160" show-overflow-tooltip />
      <el-table-column prop="gameTitle" label="游戏" width="110" />
      <el-table-column label="时间" width="150">
        <template #default="{ row }">{{ row.createdAt?.replace('T', ' ').slice(0, 16) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-popconfirm title="删除这条评论？" @confirm="removeComment(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination :page="cPage" :total="cTotal" :size="cSize" @update:page="(p: number) => { cPage = p; loadComments() }" />

    <!-- 帖子弹窗 -->
    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '发布攻略/资讯' : '编辑'" width="760px" top="4vh">
      <el-form label-width="80px">
        <el-form-item label="游戏 *">
          <el-select v-model="dialog.data.gameId" filterable placeholder="选择游戏">
            <el-option v-for="g in games" :key="g.id" :label="g.title" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="dialog.data.type" filterable allow-create default-first-option>
            <el-option v-for="t in typeOptions" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题 *">
          <el-input v-model="dialog.data.title" maxlength="190" show-word-limit />
        </el-form-item>
        <el-form-item label="摘要">
          <el-input v-model="dialog.data.summary" type="textarea" :rows="2" maxlength="1000" />
        </el-form-item>
        <el-form-item label="内容 *">
          <el-input v-model="dialog.data.content" type="textarea" :rows="14" placeholder="支持 Markdown 语法" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dialog.data.status">
            <el-radio-button :value="1">发布</el-radio-button>
            <el-radio-button :value="0">草稿</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 18px;
  margin-bottom: 14px;
}

.gap-top {
  margin-top: 24px;
}

.tip {
  font-size: 12.5px;
  color: var(--muted);
}

.table-card {
  width: 100%;
}

.row-gap {
  margin-left: 6px;
}
</style>
