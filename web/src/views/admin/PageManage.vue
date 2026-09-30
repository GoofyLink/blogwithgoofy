<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { MdEditor } from 'md-editor-v3'
import { adminFetchPages, createPage, deletePage, updatePage } from '@/api'
import type { SinglePage } from '@/types'
import AdminPagination from '@/components/AdminPagination.vue'
import { usePagination } from '@/composables/usePagination'

const list = ref<SinglePage[]>([])
const { page, pagedList } = usePagination(list)

const loading = ref(false)

const emptyForm = () => ({ slug: '', title: '', content: '' })

const dialog = reactive({
  visible: false,
  mode: 'create' as 'create' | 'edit',
  editId: 0,
  data: emptyForm(),
})

async function load() {
  loading.value = true
  try {
    list.value = (await adminFetchPages()) || []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  dialog.mode = 'create'
  dialog.data = emptyForm()
  dialog.visible = true
}

function openEdit(row: SinglePage) {
  dialog.mode = 'edit'
  dialog.editId = row.id
  dialog.data = { slug: row.slug, title: row.title, content: row.content }
  dialog.visible = true
}

async function save() {
  if (!dialog.data.slug.trim() || !dialog.data.title.trim()) {
    ElMessage.warning('slug 和标题不能为空')
    return
  }
  if (dialog.mode === 'create') {
    await createPage({ ...dialog.data })
    ElMessage.success('已创建')
  } else {
    await updatePage(dialog.editId, { ...dialog.data })
    ElMessage.success('已更新')
  }
  dialog.visible = false
  load()
}

async function remove(row: SinglePage) {
  await deletePage(row.id)
  ElMessage.success('已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="toolbar card tip-bar">
      <el-button type="primary" @click="openCreate">＋ 新建单页</el-button>
      <span>单页用于「关于」等固定页面，前台地址为 <code>/p/:slug</code> 由导航引用；内置 <code>about</code> 即关于页。</span>
    </div>

    <el-table :data="pagedList" v-loading="loading" stripe class="card table-card">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="slug" label="Slug" width="160">
        <template #default="{ row }"><code>{{ row.slug }}</code></template>
      </el-table-column>
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column label="更新时间" width="160">
        <template #default="{ row }">{{ row.updatedAt?.slice(0, 10) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-popconfirm :title="`删除单页「${row.title}」？`" @confirm="remove(row)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <AdminPagination v-model:page="page" :total="list.length" />

    <el-dialog v-model="dialog.visible" :title="dialog.mode === 'create' ? '新建单页' : '编辑单页'" width="760px" top="6vh">
      <el-form label-width="70px">
        <el-form-item label="Slug">
          <el-input v-model="dialog.data.slug" placeholder="例如 about、links-policy（唯一标识）" />
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="dialog.data.title" />
        </el-form-item>
        <el-form-item label="内容">
          <MdEditor v-model="dialog.data.content" :style="{ height: '360px' }" :toolbars-exclude="['github', 'htmlPreview', 'catalog']" />
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
.tip-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 18px;
  margin-bottom: 14px;
  font-size: 13px;
  color: var(--muted);
}

.tip-bar code {
  background: var(--surface);
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.table-card {
  width: 100%;
}
</style>
