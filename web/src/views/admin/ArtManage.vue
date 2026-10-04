<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AdminPagination from '@/components/AdminPagination.vue'
import { ElMessage } from 'element-plus'
import { adminFetchArtComments, adminFetchArts, createArt, deleteArt, deleteArtComment, updateArt, uploadImage } from '@/api'
import type { ArtCommentItem, ArtworkItem } from '@/types'
import { usePagination } from '@/composables/usePagination'

const list = ref<ArtworkItem[]>([])
const { page, pagedList } = usePagination(list)

const loading = ref(false)
const uploading = ref(false)

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: {
    title: '',
    image: '',
    description: '',
    content: '',
    author: '',
    sort: 0,
    status: 1 as 0 | 1,
  },
})

async function load() {
  loading.value = true
  try {
    list.value = (await adminFetchArts()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  const maxSort = list.value.reduce((m, a) => Math.max(m, a.sort), 0)
  dialog.mode = 'create'
  dialog.editId = 0
  dialog.data = { title: '', image: '', description: '', content: '', author: '', sort: maxSort + 1, status: 1 }
  dialog.visible = true
}

function openEdit(row: ArtworkItem) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = {
    title: row.title,
    image: row.image,
    description: row.description,
    content: row.content || '',
    author: row.author,
    sort: row.sort,
    status: row.status,
  }
  dialog.visible = true
}

async function onImageChange(file: { raw?: File }) {
  if (!file.raw) return
  uploading.value = true
  try {
    dialog.data.image = await uploadImage(file.raw)
    ElMessage.success('图片已上传，地址已填入')
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
    await createArt({ ...dialog.data })
    ElMessage.success('已添加')
  } else {
    await updateArt(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: ArtworkItem) {
  await deleteArt(row.id)
  ElMessage.success('已删除')
  load()
}

const comments = ref<ArtCommentItem[]>([])
const cTotal = ref(0)
const cPage = ref(1)
const cSize = 10

async function loadComments() {
  const data = await adminFetchArtComments({ page: cPage.value, size: cSize })
  comments.value = data.list || []
  cTotal.value = data.total
}

async function removeComment(row: ArtCommentItem) {
  await deleteArtComment(row.id)
  ElMessage.success('已删除')
  loadComments()
}

onMounted(() => {
  load()
  loadComments()
})
</script>

<template>
  <div>
    <div class="toolbar card">
      <el-button type="primary" @click="openCreate">＋ 添加作品</el-button>
      <span class="tip">图片可粘贴地址或直接上传；排序小的在前</span>
    </div>

    <el-table :data="pagedList" v-loading="loading" stripe class="card table-card">
      <el-table-column label="图片" width="100">
        <template #default="{ row }">
          <el-image
            v-if="row.image"
            :src="row.image"
            fit="cover"
            class="img-thumb"
            :preview-src-list="[row.image]"
            preview-teleported
            hide-on-click-modal
          />
          <div v-else class="img-thumb placeholder">{{ row.title.slice(0, 1) }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="180" />
      <el-table-column prop="author" label="作者" width="120" />
      <el-table-column prop="description" label="简介" min-width="220" show-overflow-tooltip />
      <el-table-column prop="sort" label="排序" width="70" />
      <el-table-column label="状态" width="85">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '展示' : '隐藏' }}
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

    <!-- 评论管理 -->
    <div class="toolbar card gap-top">
      <span class="tip">作品评论（{{ cTotal }}）</span>
    </div>
    <el-table :data="comments" stripe class="card table-card">
      <el-table-column prop="nickname" label="昵称" width="110" />
      <el-table-column prop="content" label="内容" min-width="240" show-overflow-tooltip />
      <el-table-column prop="artTitle" label="所属作品" min-width="160" show-overflow-tooltip />
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
    <AdminPagination v-model:page="page" :total="list.length" />

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '添加作品' : '编辑作品'" width="560px">
      <el-form label-width="90px">
        <el-form-item label="标题 *">
          <el-input v-model="dialog.data.title" placeholder="作品名称" />
        </el-form-item>
        <el-form-item label="图片">
          <div class="img-row">
            <el-image
              v-if="dialog.data.image"
              :src="dialog.data.image"
              fit="cover"
              class="img-preview"
              :preview-src-list="[dialog.data.image]"
              preview-teleported
              hide-on-click-modal
            />
            <div v-else class="img-preview placeholder">图</div>
            <div class="img-ops">
              <el-input
                v-model="dialog.data.image"
                placeholder="粘贴图片地址（如 /uploads/xxx.png），或点右侧按钮直接上传"
              />
              <el-upload
                accept="image/*"
                :show-file-list="false"
                :auto-upload="false"
                :on-change="onImageChange"
              >
                <el-button :loading="uploading" class="up-btn">上传图片</el-button>
              </el-upload>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="作者">
          <el-input v-model="dialog.data.author" placeholder="作者/来源（选填）" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="dialog.data.description" type="textarea" :rows="3" placeholder="作品介绍" />
        </el-form-item>
        <el-form-item label="详细介绍">
          <el-input v-model="dialog.data.content" type="textarea" :rows="6" placeholder="详情页的作品解读（支持 Markdown），选填" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="dialog.data.sort" :min="0" />
          <span class="sort-tip">数字小的排前面</span>
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
.gap-top { margin-top: 24px; }

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

.table-card {
  width: 100%;
}

.img-thumb {
  width: 56px;
  height: 42px;
  border-radius: 4px;
  display: block;
}

.img-thumb.placeholder {
  background: var(--surface);
  display: grid;
  place-items: center;
  color: var(--muted);
  font-family: var(--font-heading);
  font-weight: 700;
}

.img-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.img-preview {
  width: 72px;
  height: 54px;
  flex-shrink: 0;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--surface);
}

.img-preview.placeholder {
  display: grid;
  place-items: center;
  color: var(--muted);
  font-family: var(--font-heading);
  font-size: 20px;
  font-weight: 700;
}

.img-ops {
  flex: 1;
  display: flex;
  gap: 10px;
}

.up-btn {
  white-space: nowrap;
}

.sort-tip {
  margin-left: 10px;
  font-size: 12px;
  color: var(--muted);
}
</style>
