<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { adminFetchGames, createGame, deleteGame, updateGame, uploadImage } from '@/api'
import type { GameItem } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()
const list = ref<GameItem[]>([])
const { page, pagedList } = usePagination(list)

const loading = ref(false)
const uploading = ref(false)

const newCats = ref<string[]>([])
const catOptions = computed(() => {
  const set = new Set<string>()
  for (const g of list.value) if (g.category) set.add(g.category)
  return [...new Set([...set, ...newCats.value])]
})

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    title: '',
    cover: '',
    category: '电竞',
    platform: 'PC',
    description: '',
    tags: '',
    hot: 0,
    sort: 0,
    status: 1 as 0 | 1,
  },
})

async function load() {
  loading.value = true
  try {
    list.value = (await adminFetchGames()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  const maxSort = list.value.reduce((m, g) => Math.max(m, g.sort), 0)
  dialog.mode = 'create'
  dialog.editId = 0
  dialog.data = {
    title: '', cover: '', category: '电竞', platform: 'PC',
    description: '', tags: '', hot: 0, sort: maxSort + 1, status: 1,
  }
  dialog.visible = true
}

function openEdit(row: GameItem) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    title: row.title, cover: row.cover, category: row.category, platform: row.platform,
    description: row.description, tags: row.tags, hot: row.hot, sort: row.sort, status: row.status,
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
  if (dialog.mode === 'create') {
    await createGame({ ...dialog.data })
    ElMessage.success('已添加')
  } else {
    await updateGame(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: GameItem) {
  await deleteGame(row.id)
  ElMessage.success('已删除')
  load()
}

function formatHot(n: number): string {
  if (!n) return '0'
  return n >= 10000 ? (n / 10000).toFixed(1) + ' 万' : String(n)
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar card">
      <el-button type="primary" @click="openCreate">＋ 添加游戏</el-button>
      <span class="tip">分类自动生成前台筛选标签；热度决定右侧热门榜排序</span>
    </div>

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
      <el-table-column label="热度" width="90">
        <template #default="{ row }">{{ formatHot(row.hot) }}</template>
      </el-table-column>
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
    <AdminPagination v-model:page="page" :total="list.length" />

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '添加游戏' : '编辑游戏'" width="560px">
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
        <el-form-item label="热度">
          <el-input-number v-model="dialog.data.hot" :min="0" :step="10000" style="width: 200px" />
          <span class="tip-inline">热门榜按此排序</span>
        </el-form-item>
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
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
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
