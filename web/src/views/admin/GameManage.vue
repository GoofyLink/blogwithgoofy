<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { adminFetchGames, createGame, deleteGame, updateGame, uploadImage } from '@/api'
import type { GameItem } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

import { MdEditor } from 'md-editor-v3'
import { playStatuses, playLabel } from '@/utils/game'

const router = useRouter()
const list = ref<GameItem[]>([])
const keyword = ref('')
const categoryFilter = ref('')
const statusFilter = ref('')
const saving = ref(false)
const filteredList = computed(() => list.value.filter(g =>
  g.title.toLowerCase().includes(keyword.value.trim().toLowerCase()) &&
  (!categoryFilter.value || g.category === categoryFilter.value) &&
  (!statusFilter.value || g.playStatus === statusFilter.value)))
const { page, pagedList } = usePagination(filteredList)
watch([keyword, categoryFilter, statusFilter], () => { page.value = 1 })
async function uploadContentImages(files: File[], callback: (urls: string[]) => void) {
  uploading.value = true
  try { callback(await Promise.all(files.map(uploadImage))) }
  finally { uploading.value = false }
}

const loading = ref(false)
const uploading = ref(false)
const loadError = ref(false)

const newCats = ref<string[]>([])
const catOptions = computed(() => {
  const set = new Set<string>()
  for (const g of list.value) if (g.category) set.add(g.category)
  return [...new Set(['动作', '角色扮演', '射击', '策略', '模拟经营', ...set, ...newCats.value])]
})

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    content: '', playStatus: '' as GameItem['playStatus'], review: '', progress: '', recommended: false, recommendOrder: 0,
    title: '',
    cover: '',
    category: '',
    platform: 'PC',
    description: '',
    tags: '',
    sort: 0,
    status: 1 as 0 | 1,
  },
})

async function load() {
  loadError.value = false
  loading.value = true
  try {
    list.value = (await adminFetchGames()) || []
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
}

function openCreate() {
  const maxSort = list.value.reduce((m, g) => Math.max(m, g.sort), 0)
  dialog.mode = 'create'
  dialog.editId = 0
  dialog.data = {
    content: '', playStatus: '' as GameItem['playStatus'], review: '', progress: '', recommended: false, recommendOrder: 0,
    title: '', cover: '', category: '', platform: 'PC',
    description: '', tags: '', sort: maxSort + 1, status: 1,
  }
  dialog.visible = true
}

function openEdit(row: GameItem) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    content: row.content || '', playStatus: row.playStatus || '', review: row.review || '', progress: row.progress || '', recommended: !!row.recommended, recommendOrder: row.recommendOrder || 0,
    title: row.title, cover: row.cover, category: row.category, platform: row.platform,
    description: row.description, tags: row.tags, sort: row.sort, status: row.status,
  }
  dialog.visible = true
}

async function onCoverChange(file: { raw?: File }) {
  if (!file.raw) return
  uploading.value = true
  try {
    dialog.data.cover = await uploadImage(file.raw)
    ElMessage.success('封面已上传，地址已填入')
  } finally {
    uploading.value = false
  }
}

async function save() {
  if (!dialog.data.title.trim()) {
    ElMessage.warning('请填写标题')
    return
  }
  if (saving.value) return
  saving.value = true
  try {
    if (dialog.mode === 'create') {
      await createGame({ ...dialog.data })
      ElMessage.success('已添加')
    } else {
      await updateGame(dialog.editId, { ...dialog.data })
      ElMessage.success('已更新')
    }
    dialog.visible = false
    await load()
  } finally { saving.value = false }
}

async function remove(row: GameItem) {
  await deleteGame(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar card">
      <el-button type="primary" @click="openCreate">＋ 添加游戏</el-button>
      <span class="tip">记录游玩心得，勾选推荐后展示在前台站长推荐中</span>
    </div>

    <div class="toolbar card">
      <el-input v-model="keyword" placeholder="搜索游戏名称" clearable aria-label="搜索游戏名称" />
      <el-select v-model="categoryFilter" placeholder="全部分类" clearable><el-option v-for="c in catOptions" :key="c" :value="c" :label="c" /></el-select>
      <el-select v-model="statusFilter" placeholder="全部游玩状态" clearable><el-option v-for="s in playStatuses" :key="s.value" :value="s.value" :label="s.label" /></el-select>
    </div>
    <div v-if="loadError" role="alert">游戏列表加载失败。<el-button text @click="load">重新加载</el-button></div>
    <el-table :data="pagedList" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="sort" label="排序" width="65" />
      <el-table-column label="封面" width="100">
        <template #default="{ row }">
          <div class="cover-thumb" :style="row.cover ? { background: `url(${row.cover}) center/cover` } : {}">
            <span v-if="!row.cover">{{ row.title.slice(0, 1) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="160" />
      <el-table-column prop="category" label="分类" width="95" />
      <el-table-column prop="platform" label="平台" width="90" />
      <el-table-column label="游玩状态" width="110"><template #default="{ row }">{{ playLabel(row.playStatus) }}</template></el-table-column>
      <el-table-column label="推荐" width="65"><template #default="{ row }">{{ row.recommended ? '是' : '—' }}</template></el-table-column>
      <el-table-column prop="tags" label="标签" width="120" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '展示' : '隐藏' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="145" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button link type="success" size="small" @click="router.push(`/admin/game-posts?gameId=${row.id}`)">攻略</el-button>
          <el-popconfirm :title="`删除「${row.title}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="page" :total="filteredList.length" />

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '添加游戏' : '编辑游戏'" width="min(960px, 95vw)">
      <el-form label-width="90px">
        <el-form-item label="标题 *">
          <el-input v-model="dialog.data.title" placeholder="游戏名称" />
        </el-form-item>
        <el-form-item label="封面">
          <div class="cover-row">
            <div class="cover-preview" :style="dialog.data.cover ? { background: `url(${dialog.data.cover}) center/cover` } : {}">
              <span v-if="!dialog.data.cover">封</span>
            </div>
            <el-input v-model="dialog.data.cover" placeholder="粘贴封面地址，或点右侧上传" />
            <el-upload accept="image/*" :show-file-list="false" :auto-upload="false" :on-change="onCoverChange">
              <el-button :loading="uploading" class="up-btn">上传</el-button>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="dialog.data.category" filterable allow-create default-first-option placeholder="下拉选或输入新分类">
            <el-option v-for="c in catOptions" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="平台">
          <el-input v-model="dialog.data.platform" placeholder="如：PC / 手机 / 主机,PC" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="dialog.data.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="dialog.data.tags" placeholder="逗号分隔，如：MOBA,免费" />
        </el-form-item>
        <el-form-item label="游玩状态"><el-select v-model="dialog.data.playStatus" clearable><el-option v-for="s in playStatuses" :key="s.value" :value="s.value" :label="s.label" /></el-select></el-form-item>
        <el-form-item label="一句话评价"><el-input v-model="dialog.data.review" maxlength="300" show-word-limit placeholder="这款游戏最吸引我的地方" /></el-form-item>
        <el-form-item label="游玩进度"><el-input v-model="dialog.data.progress" maxlength="200" placeholder="例如：主线第三章 / 正在练习辅助位" /></el-form-item>
        <el-form-item label="站长推荐"><el-switch v-model="dialog.data.recommended" /></el-form-item>
        <el-form-item v-if="dialog.data.recommended" label="推荐顺序"><el-input-number v-model="dialog.data.recommendOrder" :min="0" /><span class="tip-inline">数字越小越靠前</span></el-form-item>
        <el-form-item label="详细介绍"><MdEditor v-model="dialog.data.content" :preview="true" @on-upload-img="uploadContentImages" style="height: 420px" /></el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="dialog.data.sort" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dialog.data.status">
            <el-radio-button :value="1">展示</el-radio-button>
            <el-radio-button :value="0">隐藏</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="uploading" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  margin-bottom: 14px;
}

.tip {
  font-size: 12.5px;
  color: var(--muted);
}

.tip-inline {
  margin-left: 10px;
  font-size: 12px;
  color: var(--muted);
}

.table-card {
  width: 100%;
}

.cover-thumb {
  width: 56px;
  height: 34px;
  border-radius: 4px;
  background: var(--surface);
  display: grid;
  place-items: center;
  color: var(--muted);
  font-family: var(--font-heading);
  font-weight: 700;
}

.cover-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.cover-preview {
  width: 80px;
  height: 48px;
  flex-shrink: 0;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface);
  display: grid;
  place-items: center;
  color: var(--muted);
  font-family: var(--font-heading);
  font-size: 16px;
  font-weight: 700;
}

.up-btn {
  white-space: nowrap;
}
</style>
