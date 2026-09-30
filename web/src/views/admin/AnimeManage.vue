<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminFetchAnimes, createAnime, deleteAnime, updateAnime, uploadImage } from '@/api'
import type { AnimeItem } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

const list = ref<AnimeItem[]>([])
const loading = ref(false)
const uploading = ref(false)
const activeCat = ref('')

const categories = computed(() => {
  const set = new Set<string>()
  for (const a of list.value) if (a.category) set.add(a.category)
  return [...set]
})

/** 下拉选项 = 已有分类 + 本次会话新建的分类 */
const newCats = ref<string[]>([])
const catOptions = computed(() => [...new Set([...categories.value, ...newCats.value])])

/** 新建分类按钮：输入名称后自动选中 */
function addCategory() {
  ElMessageBox.prompt('输入新分类名称（如：AI漫剧 / 漫画改 / 原创）', '新建分类', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /\S+/,
    inputErrorMessage: '分类名不能为空',
  })
    .then(({ value }) => {
      const name = value.trim()
      if (name && !catOptions.value.includes(name)) newCats.value.push(name)
      dialog.data.category = name
    })
    .catch(() => {})
}

const filteredList = computed(() =>
  activeCat.value ? list.value.filter((a) => a.category === activeCat.value) : list.value
)
const { page, pagedList } = usePagination(filteredList)

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    title: '',
    cover: '',
    category: 'AI漫剧',
    region: '日本',
    episodes: '',
    playCount: 0,
    description: '',
    rank: 0,
    status: 1 as 0 | 1,
  },
})

function formatCount(n: number): string {
  if (!n) return '0'
  return n >= 10000 ? (n / 10000).toFixed(1) + ' 万' : String(n)
}

async function load() {
  loading.value = true
  try {
    list.value = (await adminFetchAnimes()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  const maxRank = list.value.reduce((m, a) => Math.max(m, a.rank), 0)
  dialog.mode = 'create'
  dialog.editId = 0
  dialog.data = {
    title: '',
    cover: '',
    category: 'AI漫剧',
    region: '日本',
    episodes: '',
    playCount: 0,
    description: '',
    rank: maxRank + 1,
    status: 1,
  }
  dialog.visible = true
}

function openEdit(row: AnimeItem) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    title: row.title,
    cover: row.cover,
    category: row.category,
    region: row.region,
    episodes: row.episodes,
    playCount: row.playCount || 0,
    description: row.description,
    rank: row.rank,
    status: row.status,
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
    await createAnime({ ...dialog.data })
    ElMessage.success('已添加')
  } else {
    await updateAnime(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: AnimeItem) {
  await deleteAnime(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">＋ 添加动漫</el-button>
      <div v-if="categories.length" class="cat-tabs">
        <span class="cat-label">分类筛选：</span>
        <el-tag
          :type="activeCat === '' ? 'primary' : 'info'"
          class="cat-tag"
          @click="activeCat = ''"
        >全部</el-tag>
        <el-tag
          v-for="cat in categories"
          :key="cat"
          :type="activeCat === cat ? 'primary' : 'info'"
          class="cat-tag"
          @click="activeCat = activeCat === cat ? '' : cat"
        >{{ cat }}</el-tag>
      </div>
      <span class="tip">排名小的在前；封面可粘贴图片地址或直接上传</span>
    </div>

    <el-table :data="filteredList" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="rank" label="排名" width="70" />
      <el-table-column label="封面" width="80">
        <template #default="{ row }">
          <div class="cover-thumb" :style="row.cover ? { background: `url(${row.cover}) center/cover` } : {}">
            <span v-if="!row.cover">{{ row.title.slice(0, 1) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="180" />
      <el-table-column prop="category" label="分类" width="100" />
      <el-table-column prop="region" label="地区" width="90" />
      <el-table-column prop="episodes" label="集数" width="130" />
      <el-table-column label="播放量" width="100">
        <template #default="{ row }">{{ formatCount(row.playCount) }}</template>
      </el-table-column>
      <el-table-column prop="description" label="简介" min-width="200" show-overflow-tooltip />
      <el-table-column label="状态" width="85">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '上架' : '下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除「${row.title}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="page" :total="filteredList.length" />

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '添加动漫' : '编辑动漫'" width="560px">
      <el-form label-width="90px">
        <el-form-item label="标题 *">
          <el-input v-model="dialog.data.title" placeholder="动漫名称" />
        </el-form-item>
        <el-form-item label="封面">
          <div class="cover-row">
            <div class="cover-preview" :style="dialog.data.cover ? { background: `url(${dialog.data.cover}) center/cover` } : {}">
              <span v-if="!dialog.data.cover">{{ dialog.data.title.slice(0, 1) || '封' }}</span>
            </div>
            <el-input
              v-model="dialog.data.cover"
              placeholder="粘贴封面图片地址，或点右侧按钮直接上传"
            />
            <el-upload
              accept="image/*"
              :show-file-list="false"
              :auto-upload="false"
              :on-change="onCoverChange"
            >
              <el-button :loading="uploading" class="up-btn">上传</el-button>
            </el-upload>
          </div>
        </el-form-item>
        <el-form-item label="分类">
          <div class="cat-row">
            <el-select
              v-model="dialog.data.category"
              filterable
              allow-create
              default-first-option
              placeholder="下拉选择已有分类，或输入新分类"
              class="cat-select"
            >
              <el-option v-for="cat in catOptions" :key="cat" :label="cat" :value="cat" />
            </el-select>
            <el-button class="add-btn" @click="addCategory">＋ 新建分类</el-button>
          </div>
        </el-form-item>
        <el-form-item label="地区">
          <el-input v-model="dialog.data.region" placeholder="如：日本 / 中国" />
        </el-form-item>
        <el-form-item label="集数">
          <el-input v-model="dialog.data.episodes" placeholder="如：更新至第12集 / 全13集" />
        </el-form-item>
        <el-form-item label="播放量">
          <el-input-number v-model="dialog.data.playCount" :min="0" :step="10000" style="width: 200px" />
          <span class="rank-tip">右侧热播榜按此数值排序</span>
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="dialog.data.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="排名">
          <el-input-number v-model="dialog.data.rank" :min="0" />
          <span class="rank-tip">数字小的排前面</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="dialog.data.status">
            <el-radio-button :value="1">上架</el-radio-button>
            <el-radio-button :value="0">下架</el-radio-button>
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
  margin-bottom: 14px;
}

.tip {
  font-size: 12.5px;
  color: var(--muted);
}

.cat-tabs {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.cat-label {
  font-size: 13px;
  color: var(--muted);
}

.cat-tag {
  cursor: pointer;
}

.sort-btns {
  display: flex;
  flex-direction: column;
  gap: 0;
  line-height: 1;
}

.sort-btn {
  padding: 0 4px;
}

.table-card {
  width: 100%;
}

.cover-thumb {
  width: 42px;
  height: 56px;
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
  width: 60px;
  height: 80px;
  flex-shrink: 0;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface);
  display: grid;
  place-items: center;
  color: var(--muted);
  font-family: var(--font-heading);
  font-size: 22px;
  font-weight: 700;
}

.up-btn {
  white-space: nowrap;
}

.rank-tip {
  margin-left: 10px;
  font-size: 12px;
  color: var(--muted);
}

.cat-row {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}

.cat-select {
  flex: 1;
}
</style>
