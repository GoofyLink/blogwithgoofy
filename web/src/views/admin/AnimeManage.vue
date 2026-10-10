<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  adminFetchAnimes,
  createAnime,
  updateAnime,
  deleteAnime,
  advanceAnime,
} from '@/api'
import type { AnimeItem } from '@/types'
import AnimeEditor from '@/components/AnimeEditor.vue'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'
import {
  newAnimeForm,
  watchStatuses,
  watchLabel,
  progressLabel,
} from '@/utils/anime'
const list = ref<AnimeItem[]>([])
const loading = ref(false),
  error = ref(false),
  saving = ref(false),
  uploading = ref(false),
  visible = ref(false)
const editId = ref(0),
  advancing = ref<number | null>(null)
const form = ref(newAnimeForm())
const keyword = ref(''),
  category = ref(''),
  status = ref('')
const categories = computed(() => [
  ...new Set(list.value.map((a) => a.category).filter(Boolean)),
])
const filtered = computed(() =>
  list.value.filter(
    (a) =>
      a.title.toLowerCase().includes(keyword.value.trim().toLowerCase()) &&
      (!category.value || a.category === category.value) &&
      (!status.value || a.watchStatus === status.value),
  ),
)
const { page, pagedList } = usePagination(filtered)
watch([keyword, category, status], () => {
  page.value = 1
})
async function load() {
  loading.value = true
  error.value = false
  try {
    list.value = (await adminFetchAnimes()) || []
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}
function open(row?: AnimeItem) {
  editId.value = row?.id || 0
  form.value = newAnimeForm()
  if (row) {
    const {
      id,
      createdAt,
      updatedAt,
      progressUpdatedAt,
      reviewUpdatedAt,
      ...data
    } = row
    form.value = { ...form.value, ...data }
  }
  visible.value = true
}
async function save() {
  if (saving.value || uploading.value) return
  if (!form.value.title.trim()) {
    ElMessage.warning('请填写动漫名称')
    return
  }
  if (
    form.value.totalEpisodes != null &&
    form.value.watched > form.value.totalEpisodes
  ) {
    ElMessage.warning('已看集数不能超过总集数')
    return
  }
  if (form.value.watchUrl) {
    try {
      const u = new URL(form.value.watchUrl)
      if (!['http:', 'https:'].includes(u.protocol)) throw new Error()
    } catch {
      ElMessage.warning('请填写有效的 http 或 https 观看链接')
      return
    }
  }
  saving.value = true
  try {
    if (editId.value) await updateAnime(editId.value, form.value)
    else await createAnime(form.value)
    ElMessage.success('已保存')
    visible.value = false
    await load()
  } finally {
    saving.value = false
  }
}
async function advance(row: AnimeItem) {
  if (advancing.value != null) return
  advancing.value = row.id
  try {
    await advanceAnime(row.id)
    ElMessage.success('观看进度已更新')
    await load()
  } finally {
    advancing.value = null
  }
}
async function remove(row: AnimeItem) {
  await deleteAnime(row.id)
  ElMessage.success('已删除')
  await load()
}
onMounted(load)
</script>
<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="open()">添加动漫</el-button
      ><span>记录追番进度、个人评价与观后感</span>
    </div>
    <div class="filters">
      <el-input
        v-model="keyword"
        clearable
        placeholder="搜索动漫名称"
      /><el-select v-model="category" clearable placeholder="全部题材"
        ><el-option
          v-for="c in categories"
          :key="c"
          :value="c"
          :label="c" /></el-select
      ><el-select v-model="status" clearable placeholder="全部观看状态"
        ><el-option
          v-for="s in watchStatuses"
          :key="s.value"
          :value="s.value"
          :label="s.label"
      /></el-select>
    </div>
    <div v-if="error" role="alert">
      动漫列表加载失败。<el-button text @click="load">重试</el-button>
    </div>
    <el-table :data="pagedList" v-loading="loading" class="card" stripe>
      <el-table-column prop="title" label="作品" min-width="180" />
      <el-table-column prop="category" label="题材" width="100" />
      <el-table-column label="观看状态" width="110"
        ><template #default="{ row }">{{
          watchLabel(row.watchStatus)
        }}</template></el-table-column
      >
      <el-table-column label="进度" width="110"
        ><template #default="{ row }">{{
          progressLabel(row)
        }}</template></el-table-column
      >
      <el-table-column label="个人评分" width="95"
        ><template #default="{ row }">{{
          row.rating == null ? '未评分' : row.rating.toFixed(1)
        }}</template></el-table-column
      >
      <el-table-column label="喜爱排名" width="90"
        ><template #default="{ row }">{{
          row.rank || '—'
        }}</template></el-table-column
      >
      <el-table-column label="推荐" width="70"
        ><template #default="{ row }">{{
          row.recommended ? '是' : '—'
        }}</template></el-table-column
      >
      <el-table-column label="展示" width="80"
        ><template #default="{ row }"
          ><el-tag :type="row.status ? 'success' : 'info'">{{
            row.status ? '上架' : '下架'
          }}</el-tag></template
        ></el-table-column
      >
      <el-table-column label="操作" width="250" fixed="right"
        ><template #default="{ row }"
          ><el-button link type="primary" @click="open(row)">编辑</el-button
          ><el-button
            link
            :loading="advancing === row.id"
            :disabled="
              advancing != null ||
              (row.totalEpisodes != null && row.watched >= row.totalEpisodes)
            "
            @click="advance(row)"
            >看完一集 +1</el-button
          ><el-popconfirm
            :title="`删除「${row.title}」？`"
            @confirm="remove(row)"
            ><template #reference
              ><el-button link type="danger">删除</el-button></template
            ></el-popconfirm
          ></template
        ></el-table-column
      >
    </el-table>
    <AdminPagination v-model:page="page" :total="filtered.length" />
    <el-dialog
      v-model="visible"
      :title="editId ? '编辑动漫' : '添加动漫'"
      width="min(1000px, 96vw)"
      :close-on-click-modal="false"
      :close-on-press-escape="!saving && !uploading"
      :show-close="!saving && !uploading"
      destroy-on-close
    >
      <AnimeEditor
        v-model="form"
        :categories="categories"
        @uploading="uploading = $event"
      />
      <template #footer
        ><el-button :disabled="saving || uploading" @click="visible = false"
          >取消</el-button
        ><el-button
          type="primary"
          :loading="saving"
          :disabled="uploading"
          @click="save"
          >保存</el-button
        ></template
      >
    </el-dialog>
  </div>
</template>
<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.toolbar span {
  color: var(--muted);
  font-size: 13px;
}
.filters {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 12px;
  margin-bottom: 18px;
}
@media (max-width: 700px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
